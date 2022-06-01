package selfupdate

import (
	"os"
	"syscall"

	"github.com/fynelabs/selfupdate/internal/osext"
)

func Restart() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	exe, err := osext.Executable()
	if err != nil {
		return err
	}

	_, err = os.StartProcess(exe, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return err
	}

	return nil
}
