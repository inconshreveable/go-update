package selfupdate

import (
	"os"
	"syscall"

	"github.com/fynelabs/selfupdate/internal/osext"
)

func restart(exiter func(error), executable string) error {
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
