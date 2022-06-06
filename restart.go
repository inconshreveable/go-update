package selfupdate

import (
	"os"
	"syscall"

	"github.com/fynelabs/selfupdate/internal/osext"
)

// Restart will attempt to restar the current application, any error will be returned.
// If the exiter function is passed in it will be responsible for terminating the old processes.
// If exiter is passed an error it can assume the restart failed and handle appropriately.
// It is recommended to provide the executable that need to be started, but if none is
// specified we will try to guess it correctly.
func Restart(exiter func(error), executable string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if executable == "" {
		executable, err = osext.Executable()
		if err != nil {
			return err
		}
	}

	_, err = os.StartProcess(executable, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   &syscall.SysProcAttr{},
	})

	if exiter != nil {
		exiter(err)
	} else if err == nil {
		os.Exit(0)
	}
	return err
}
