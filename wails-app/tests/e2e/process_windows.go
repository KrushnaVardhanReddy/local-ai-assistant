//go:build windows

package e2e

import (
	"os/exec"
	"strconv"
)

// setupProcessGroup is a no-op on Windows since Setpgid doesn't exist
func setupProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup uses taskkill on Windows to kill child processes
func killProcessGroup(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	}
}
