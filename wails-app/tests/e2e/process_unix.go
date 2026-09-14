//go:build !windows

package e2e

import (
	"os/exec"
	"syscall"
)

// setupProcessGroup ensures the command runs in its own process group
func setupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup kills the command's entire process group
func killProcessGroup(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
