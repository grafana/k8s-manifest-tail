package describe_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
)

type cliResult struct {
	Stdout string
	Stderr string
	Error  error
}

func runCLI(args ...string) cliResult {
	commandArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", commandArgs...)
	cmd.Dir = projectRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return cliResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Error:  err,
	}
}

func fixturePath(name string) string {
	return filepath.Join(fixturesDir, name)
}
