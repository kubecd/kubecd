package exec

import (
	osexec "os/exec"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type RealRunner struct{}

func (r RealRunner) Run(cmd string, args ...string) ([]byte, error) {
	return osexec.Command(cmd, args...).Output()
}
