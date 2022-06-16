package selfupdate

import (
	"io"

	"github.com/fynelabs/selfupdate/internal/binarydist"
)

// Patcher defines an interface for applying binary patches to an old item to get an updated item.
type Patcher interface {
	Patch(old io.Reader, new io.Writer, patch io.Reader) error
}

type patchFn func(io.Reader, io.Writer, io.Reader) error

// Patch will call the patchFn function to satisfy a Patcher interface
func (fn patchFn) Patch(old io.Reader, new io.Writer, patch io.Reader) error {
	return fn(old, new, patch)
}

// NewBSDiffPatcher returns a new Patcher that applies binary patches using
// the bsdiff algorithm. See http://www.daemonology.net/bsdiff/
func NewBSDiffPatcher() Patcher {
	return patchFn(binarydist.Patch)
}
