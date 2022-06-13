package selfupdate

import (
	"crypto/ed25519"
	"errors"
	"io"
	"sync"
	"time"
)

// ErrNotSupported is returned by `Manage` when it is not possible to manage the current application.
var ErrNotSupported = errors.New("operating system not supported")

type Source interface {
	Get(*Version) (io.ReadCloser, int64, error)
	GetSignature() ([64]byte, error)
	LatestVersion() (*Version, error)
}

type Config struct {
	Current   *Version
	Source    Source
	Schedule  Schedule
	PublicKey ed25519.PublicKey

	ProgressCallback       func(float64, error) // if present will call back with 0.0 at the start, rising through to 1.0 at the end if the progress is known. A negative start number will be sent if size is unknown, any error will pass as is and the process is considered done
	RestartConfirmCallback func() bool          // if present will ask for user acceptance before restarting app
	UpgradeConfirmCallback func(string) bool    // if present will ask for user acceptance, it can present the message passed
	ExitCallback           func(error)          // if present will be expected to handle app exit procedure
}

type Repeating int

const (
	None    Repeating = iota
	Hourly            // Will schedule in the next hour and repeat it every hour after
	Daily             // Will schedule next day and repeat it every day after
	Monthly           // Will schedule next month and repeat it every month after
)

type ScheduleAt struct {
	Repeating
	time.Time // Offset time used to define when in a minute/hour/day/month to actually trigger the schedule
}

type Schedule struct {
	FetchOnStart bool
	Interval     time.Duration
	At           ScheduleAt
}

type Version struct {
	Number string    // if the app knows its version and supports checking metadata
	Build  int       // if the app has a build number this could be compared
	Date   time.Time // last update, could be mtime
}

type Updater struct {
	lock       sync.Mutex
	conf       *Config
	executable string
}

func (u *Updater) CheckNow() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	v := u.conf.Current
	if v == nil {
		mtime, err := lastModifiedExecutable()
		if err != nil {
			return err
		}

		v = &Version{Date: mtime}
	}

	latest, err := u.conf.Source.LatestVersion()
	if err != nil {
		return err
	}
	if !latest.Date.After(v.Date) {
		logDebug("Local binary time (%v) is recent enough compared to the online version (%v).\n", v.Date.Format(time.RFC1123Z), latest.Date.Format(time.RFC1123Z))
		return nil
	}

	if ask := u.conf.UpgradeConfirmCallback; ask != nil {
		if !ask("New version found") {
			logInfo("The user didn't confirm the upgrade.\n")
			return nil
		}
	}

	s, err := u.conf.Source.GetSignature()
	if err != nil {
		return err
	}

	r, contentLength, err := u.conf.Source.Get(v)
	if err != nil {
		return err
	}
	defer r.Close()

	pr := &progressReader{Reader: r, progressCallback: u.conf.ProgressCallback, contentLength: contentLength}

	u.executable, err = applyUpdate(pr, u.conf.PublicKey, s)
	if err != nil {
		return err
	}

	if ask := u.conf.RestartConfirmCallback; ask != nil {
		if !ask() {
			logInfo("The user didn't confirm restarting the application after upgrade.\n")
			return nil
		}
	}
	return u.Restart()
}

func (u *Updater) Restart() error {
	return restart(u.conf.ExitCallback, u.executable)
}

// Manage sets up an Updater and runs it to manage the current executable.
func Manage(conf *Config) (*Updater, error) {
	updater := &Updater{conf: conf}

	go func() {
		if updater.conf.Schedule.FetchOnStart {
			logInfo("Doing an initial upgrade check.\n")
			err := updater.CheckNow()
			if err != nil {
				logError("Upgrade error: %v\n", err)
			}
		}

		if updater.conf.Schedule.Interval != 0 || updater.conf.Schedule.At.Repeating != None {
			go func() {
				triggerSchedule(updater)
			}()
		}
	}()

	// TODO check if we can support the current app!
	return updater, nil
}

// ManualUpdate applies a specific update manually instead of managing the update of this app automatically.
func ManualUpdate(s Source, publicKey ed25519.PublicKey) error {
	v := &Version{}
	r, _, err := s.Get(v)
	if err != nil {
		return err
	}

	signature, err := s.GetSignature()
	if err != nil {
		return err
	}

	_, err = applyUpdate(r, publicKey, signature)
	return err
}

func applyUpdate(r io.Reader, publicKey ed25519.PublicKey, signature [64]byte) (string, error) {
	opts := &Options{}
	opts.Signature = signature[:]
	opts.PublicKey = publicKey

	err := apply(r, opts)
	if err != nil {
		return "", err
	}
	return opts.TargetPath, nil
}

func triggerSchedule(updater *Updater) {
	for {
		var next time.Duration

		if updater.conf.Schedule.Interval != 0 {
			next = updater.conf.Schedule.Interval
		}
		if updater.conf.Schedule.At.Repeating != None {
			at := nextScheduleAt(updater.conf.Schedule.At.Repeating, updater.conf.Schedule.At.Time)
			if next == 0 || at < next {
				next = at
			}
		}

		time.Sleep(next)
		logInfo("Scheduled upgrade check after %s.\n", next)
		err := updater.CheckNow()
		if err != nil {
			logError("Upgrade error: %v\n", err)
		}
	}
}

func nextScheduleAt(repeating Repeating, offset time.Time) time.Duration {
	now := time.Now()
	var next time.Time
	switch repeating {
	case Hourly:
		next = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, offset.Minute(), offset.Second(), offset.Nanosecond(), time.Local)
	case Daily:
		next = time.Date(now.Year(), now.Month(), now.Day()+1, offset.Hour(), offset.Minute(), offset.Second(), offset.Nanosecond(), time.Local)
	case Monthly:
		next = time.Date(now.Year(), now.Month()+1, offset.Day(), offset.Hour(), offset.Minute(), offset.Second(), offset.Nanosecond(), time.Local)
	}

	return next.Sub(now)
}
