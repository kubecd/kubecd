package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
)

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type RealRunner struct{}

func (r RealRunner) Run(cmd string, args ...string) ([]byte, error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s %s\n", cmd, strings.Join(args, " "))
	return osexec.Command(cmd, args...).Output()
}
