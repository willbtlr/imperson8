package bash

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type BashError struct {
	Cmd string
	Err error
	Out string
}

func (b BashError) Error() string {
	return fmt.Sprintf("Bash error executing %s %s: %s", b.Cmd, b.Err, b.Out)
}

func Execf(cmd string, args ...interface{}) (string, error) {
	cmd = fmt.Sprintf(cmd, args...)
	path, e := exec.LookPath("bash")
	if e != nil {
		return "", e
	}

	output, e := exec.Command(path, "-c", cmd).CombinedOutput()
	if e != nil {
		return "", BashError{Cmd: cmd, Err: e, Out: string(output)}
	}

	return string(output), nil
}


func CheckDeps(cliDeps ...string) error {
	errs := make([]string, 0, len(cliDeps))
	for _, dep := range cliDeps {
		_, e := exec.LookPath(dep)
		if e != nil {
			errs = append(errs, fmt.Sprintf("cant find %s on $PATH: %s", dep, e.Error()))
		}
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}
