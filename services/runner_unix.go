//go:build !windows

package services

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	// Negative PID targets the process group, killing the shell and any
	// descendants it spawned (e.g. a sleep inside `sh -c "sleep 5"`).
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		// Fall back to killing just the shell if the group is already gone.
		return cmd.Process.Kill()
	}
	return nil
}
