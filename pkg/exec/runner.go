package exec

import (
	"fmt"
	osexec "os/exec"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type RealRunner struct{}

func (r RealRunner) Run(cmd string, args ...string) ([]byte, error) {
	fmt.Printf("Running: %s %v\n", cmd, args)
	return osexec.Command(cmd, args...).Output()
}
