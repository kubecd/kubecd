package updates

import (
	"github.com/kubecd/kubecd/pkg/exec"
	"testing"
)

var runner exec.Runner = exec.RealRunner{}

// TestHelperProcess is required boilerplate (one per package) for using exec.TestRunner
func TestHelperProcess(t *testing.T) {
	exec.InsideHelperProcess()
}
