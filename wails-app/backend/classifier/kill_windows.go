//go:build windows

package classifier

import (
	"fmt"
	"os/exec"
)

func setProcessGroup(cmd *exec.Cmd) {
	// Not required for basic kill on Windows, but taskkill is used below for process trees.
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		// Use taskkill to terminate the process tree (/T) forcefully (/F)
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", cmd.Process.Pid))
		if err := killCmd.Run(); err != nil {
			// Fallback to basic kill if taskkill fails
			cmd.Process.Kill()
		}
	}
}
