package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestRunner(t *testing.T) {
	expected := "testing the test runner"
	runner := TestRunner{Output: []byte(expected)}
	output, err := runner.Run("command")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(output))
	runner = TestRunner{ExitCode: 42}
	output, err = runner.Run("command")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "exit status 42")
	runner = TestRunner{ExpectedCommand: []string{"foo", "bar"}}
	output, err = runner.Run("foo", "bar", "gazonk")
	assert.Error(t, err)
	assert.Equal(t, "exit status 127", err.Error())
	assert.Equal(t, "expected argv [foo bar], got [foo bar gazonk]", string(output))
}

func TestHelperProcess(t *testing.T) {
	InsideHelperProcess()
}
