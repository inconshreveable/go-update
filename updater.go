package selfupdate

import (
	"errors"
	"io"
	"time"
)

// ErrNotSupported is returned by `Manage` when it is not possible to manage the current application.
var ErrNotSupported = errors.New("operating system not supported")

type Source interface {
	Get(*Version) (io.Reader, error)
	GetSignature() (string, error)
	LatestVersion() (*Version, error)
}

type Config struct {
	Current  *Version
	Source   Source
	Schedule Schedule

	ProgressCallback       func(float64, error) // if present will call back with 0.0 at the start, rising through to 1.0 at the end if the progress is known. A negative start number will be sent if size is unknown, any error will pass as is and the process is considered done
	RestartConfirmCallback func() bool          // if present will ask for user acceptance before restarting app
	UpgradeConfirmCallback func(string) bool    // if present will ask for user acceptance, it can present the message passed
}

type Schedule struct {
	FetchOnStart bool
	Interval     time.Duration
}

type Version struct {
	Number string    // if the app knows its version and supports checking metadata
	Build  int       // if the app has a build number this could be compared
	Date   time.Time // last update, could be mtime
}

type Updater struct {
	conf *Config
}

func (u *Updater) CheckNow() error {
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
		return nil
	}

	if ask := u.conf.UpgradeConfirmCallback; ask != nil {
		if !ask("New version found") {
			return nil
		}
	}

	r, err := u.conf.Source.Get(latest)
	if err != nil {
		return err
	}
	err = Apply(r, Options{})
	if err != nil {
		return err
	}

	if ask := u.conf.RestartConfirmCallback; ask != nil {
		if !ask() {
			return nil
		}
	}
	return u.Restart()
}

func (u *Updater) Restart() error {
	return Restart()
}

// Manage sets up an Updater and runs it to manage the current executable.
func Manage(conf *Config) (*Updater, error) {
	// TODO check if we can support the current app!
	return &Updater{conf: conf}, nil
}

// ManualUpdate applies a specific update manually instead of managing the update of this app automatically.
func ManualUpdate(s Source) error {
	v := &Version{}
	r, err := s.Get(v)
	if err != nil {
		return err
	}

	return Apply(r, Options{})
}
