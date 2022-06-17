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

// Source define an interface that is able to get an update
type Source interface {
	Get(*Version) (io.ReadCloser, int64, error) // Get the executable to be updated to
	GetSignature() ([64]byte, error)            // Get the signature that match the executable
	LatestVersion() (*Version, error)           // Get the latest version information to determine if we should trigger an update
}

// Config define extra parameter necessary to manage the updating process
type Config struct {
	Current   *Version          // If present will define the current version of the executable that need update
	Source    Source            // Necessary Source for update
	Schedule  Schedule          // Define when to trigger an update
	PublicKey ed25519.PublicKey // The public key that match the private key used to generate the signature of future update

	ProgressCallback       func(float64, error) // if present will call back with 0.0 at the start, rising through to 1.0 at the end if the progress is known. A negative start number will be sent if size is unknown, any error will pass as is and the process is considered done
	RestartConfirmCallback func() bool          // if present will ask for user acceptance before restarting app
	UpgradeConfirmCallback func(string) bool    // if present will ask for user acceptance, it can present the message passed
	ExitCallback           func(error)          // if present will be expected to handle app exit procedure
}

// Repeating pattern for scheduling update at a specific time
type Repeating int

const (
	// None will not schedule
	None Repeating = iota
	// Hourly will schedule in the next hour and repeat it every hour after
	Hourly
	// Daily will schedule next day and repeat it every day after
	Daily
	// Monthly will schedule next month and repeat it every month after
	Monthly
)

// ScheduleAt define when a repeating update at a specific time should be triggered
type ScheduleAt struct {
	Repeating // The pattern to enforce for the repeating schedule
	time.Time // Offset time used to define when in a minute/hour/day/month to actually trigger the schedule
}

// Schedule define when to trigger an update
type Schedule struct {
	FetchOnStart bool          // Trigger when the updater is created
	Interval     time.Duration // Trigger at regular interval
	At           ScheduleAt    // Trigger at a specific time
}

// Version define an executable versionning information
type Version struct {
	Number string    // if the app knows its version and supports checking metadata
	Build  int       // if the app has a build number this could be compared
	Date   time.Time // last update, could be mtime
}

// Updater is managing update for your application in the background
type Updater struct {
	lock       sync.Mutex
	conf       *Config
	executable string
}

// CheckNow will manually trigger a check of an update and if one is present will start the update process
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

// Restart once an update is done can trigger a restart of the binary. This is useful to implement a restart later policy.
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
		var delay time.Duration

		if updater.conf.Schedule.Interval != 0 {
			delay = updater.conf.Schedule.Interval
		}
		if updater.conf.Schedule.At.Repeating != None {
			at := delayUntilNextTriggerAt(updater.conf.Schedule.At.Repeating, updater.conf.Schedule.At.Time)
			if delay == 0 || at < delay {
				delay = at
			}
		}

		time.Sleep(delay)
		logInfo("Scheduled upgrade check after %s.\n", delay)
		err := updater.CheckNow()
		if err != nil {
			logError("Upgrade error: %v\n", err)
		}
	}
}

func delayUntilNextTriggerAt(repeating Repeating, offset time.Time) time.Duration {
	now := time.Now().In(offset.Location())
	var next time.Time
	switch repeating {
	case Hourly:
		next = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, offset.Minute(), offset.Second(), offset.Nanosecond(), offset.Location())
	case Daily:
		next = time.Date(now.Year(), now.Month(), now.Day()+1, offset.Hour(), offset.Minute(), offset.Second(), offset.Nanosecond(), offset.Location())
	case Monthly:
		next = time.Date(now.Year(), now.Month()+1, offset.Day(), offset.Hour(), offset.Minute(), offset.Second(), offset.Nanosecond(), offset.Location())
	}

	return next.Sub(now)
}
