package selfupdate

import (
	"crypto/ed25519"
	"errors"
	"io"
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

	s, err := u.conf.Source.GetSignature()
	if err != nil {
		return err
	}

	r, contentLength, err := u.conf.Source.Get(v)
	if err != nil {
		return err
	}
	defer r.Close()

	opts := Options{}
	opts.Signature = s[:]
	opts.PublicKey = u.conf.PublicKey

	pr := &progressReader{Reader: r, progressCallback: u.conf.ProgressCallback, contentLength: contentLength}

	err = Apply(pr, opts)
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

	opts := Options{}
	opts.Signature = signature[:]
	opts.PublicKey = publicKey

	return Apply(r, opts)
}
