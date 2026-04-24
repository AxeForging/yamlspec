//go:build windows

package services

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {
	// No process-group concept on Windows — exec.CommandContext's default
	// Kill behavior is used by leaving cmd.SysProcAttr unchanged.
}

func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
