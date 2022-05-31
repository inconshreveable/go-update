//go:build !windows
// +build !windows

package selfupdate

func hideFile(path string) error {
	return nil
}
