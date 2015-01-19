package update

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"bitbucket.org/kardianos/osext"
)

type Update struct {
	// The update to apply
	Update io.Reader

	// TargetPath defines the path to the file to update.
	// The emptry string means 'the executable file of the running program'.
	TargetPath string

	// Create TargetPath replacement with this file mode
	TargetMode os.FileMode

	// Checksum of the new binary to verify against. If nil, no checksum or signature verification is done.
	Checksum []byte

	// Public key to use for signature verification. If nil, no signature verification is done.
	PublicKey crypto.PublicKey

	// Signature to verify the updated file. If nil, no signature verification is done.
	Signature []byte

	// Pluggable signature verification algorithm. If nil, ECDSA is used.
	Verifier Verifier

	// Use this hash function to generate the checksum. If not set, SHA256 is used.
	Hash crypto.Hash

	// If non-nil, treat the Update as a patch to apply on the old file
	Patcher Patcher
}

type Verifier interface {
	VerifySignature(checksum, signature []byte, h crypto.Hash, publicKey crypto.PublicKey) error
}

type Patcher interface {
	Patch(old io.Reader, new io.Writer, patch io.Reader) error
}

// Do performs the update of TargetFile with Update.
//
// Do performs the following actions to ensure a safe cross-platform update:
//
// 1. If configured, applies the contents of the Update io.Reader as a binary patch.
//
// 2. If configured, computes the checksum and verifies it matches.
//
// 3. If configured, verifies the signature with a public key.
//
// 4. Creates a new file, /path/to/.target.new with the TargetMode with the contents of the updated file
//
// 5. Renames /path/to/target to /path/to/.target.old
//
// 6. Renames /path/to/.target.new to /path/to/target
//
// 7. If the rename is successful, deletes /path/to/.target.old, returns no error
//
// 8. If the rename fails, attempts to rename /path/to/.target.old back to /path/to/target
// If this operation fails, it is reported in the errRecover return value so as not to
// mask the original error that caused the recovery attempt.
//
// On Windows, the removal of /path/to/.target.old always fails, so instead,
// we just make the old file hidden instead.
func (u *Update) Do() (err error, errRecover error) {
	// validate
	verify := false
	switch {
	case u.Signature != nil && u.PublicKey != nil:
		// okay
		verify = true
	case u.Signature != nil:
		return errors.New("no public key to verify signature with"), nil
	case u.PublicKey != nil:
		return errors.New("No signature to verify with"), nil
	}

	// set defaults
	if u.Hash == 0 {
		u.Hash = crypto.SHA256
	}
	if u.Verifier == nil {
		u.Verifier = ECDSAVerifier
	}

	// get target path
	u.TargetPath, err = u.getPath()
	if err != nil {
		return
	}

	var newBytes []byte
	if u.Patcher != nil {
		if newBytes, err = u.applyPatch(); err != nil {
			return
		}
	} else {
		// no patch to apply, go on through
		if newBytes, err = ioutil.ReadAll(u.Update); err != nil {
			return
		}
	}

	// verify checksum if requested
	if u.Checksum != nil {
		if err = u.verifyChecksum(newBytes); err != nil {
			return
		}
	}

	if verify {
		if err = u.verifySignature(newBytes); err != nil {
			return
		}
	}

	// get the directory the executable exists in
	updateDir := filepath.Dir(u.TargetPath)
	filename := filepath.Base(u.TargetPath)

	// Copy the contents of of newbinary to a the new executable file
	newPath := filepath.Join(updateDir, fmt.Sprintf(".%s.new", filename))
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, u.TargetMode)
	if err != nil {
		return
	}
	defer fp.Close()
	_, err = io.Copy(fp, bytes.NewReader(newBytes))

	// if we don't call fp.Close(), windows won't let us move the new executable
	// because the file will still be "in use"
	fp.Close()

	// this is where we'll move the executable to so that we can swap in the updated replacement
	oldPath := filepath.Join(updateDir, fmt.Sprintf(".%s.old", filename))

	// delete any existing old exec file - this is necessary on Windows for two reasons:
	// 1. after a successful update, Windows can't remove the .old file because the process is still running
	// 2. windows rename operations fail if the destination file already exists
	_ = os.Remove(oldPath)

	// move the existing executable to a new file in the same directory
	err = os.Rename(u.TargetPath, oldPath)
	if err != nil {
		return
	}

	// move the new exectuable in to become the new program
	err = os.Rename(newPath, u.TargetPath)

	if err != nil {
		// copy unsuccessful
		errRecover = os.Rename(oldPath, u.TargetPath)
	} else {
		// copy successful, remove the old binary
		errRemove := os.Remove(oldPath)

		// windows has trouble with removing old binaries, so hide it instead
		if errRemove != nil {
			_ = hideFile(oldPath)
		}
	}

	return
}

// CheckPermissions() determines whether the process has the correct permissions to
// perform the requested update. If the update can proceed, it returns nil, otherwise
// it returns the error that would occur if an update were attempted.
func (u *Update) CheckPermissions() error {
	// get the directory the file exists in
	path, err := u.getPath()
	if err != nil {
		return err
	}

	fileDir := filepath.Dir(path)
	fileName := filepath.Base(path)

	// attempt to open a file in the file's directory
	newPath := filepath.Join(fileDir, fmt.Sprintf(".%s.new", fileName))
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, u.TargetMode)
	if err != nil {
		return err
	}
	fp.Close()

	_ = os.Remove(newPath)
	return nil
}

func (u *Update) getPath() (string, error) {
	if u.TargetPath == "" {
		return osext.Executable()
	} else {
		return u.TargetPath, nil
	}
}

// VerifyWithPEM configures the update to use the given PEM-formatted
// public key to verify the update's signature.
func (u *Update) VerifyWithPEM(publicKeyPEM []byte) error {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return errors.New("couldn't parse PEM data")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	u.PublicKey = pub
	return nil
}

// FromURL makes an HTTP request to the given URL and configures the update
// to use the response body as the update.
func (u *Update) FromURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	u.Update = resp.Body
	return nil
}

func (u *Update) applyPatch() ([]byte, error) {
	// open the file to patch
	old, err := os.Open(u.TargetPath)
	if err != nil {
		return nil, err
	}
	defer old.Close()

	// apply the patch
	var applied bytes.Buffer
	if err = u.Patcher.Patch(old, &applied, u.Update); err != nil {
		return nil, err
	}

	return applied.Bytes(), nil
}

func (u *Update) verifyChecksum(updated []byte) error {
	checksum, err := checksumFor(u.Hash, updated)
	if err != nil {
		return err
	}

	if !bytes.Equal(u.Checksum, checksum) {
		return fmt.Errorf("Updated file has wrong checksum. Expected: %x, got: %x", u.Checksum, checksum)
	}
	return nil
}

func (u *Update) verifySignature(updated []byte) error {
	checksum, err := checksumFor(u.Hash, updated)
	if err != nil {
		return err
	}
	return u.Verifier.VerifySignature(checksum, u.Signature, u.Hash, u.PublicKey)
}

func checksumFor(h crypto.Hash, payload []byte) ([]byte, error) {
	if !h.Available() {
		return nil, errors.New("requested hash function not available")
	}
	hash := h.New()
	hash.Write(payload) // guaranteed not to error
	return hash.Sum([]byte{}), nil
}
