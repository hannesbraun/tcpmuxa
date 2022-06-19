//go:build !windows

package sysutils

import (
	"os/exec"
	"syscall"
)

// PrepareCmd prepares a command to be executed.
// On Unix, the pgid will be set.
func PrepareCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Kill kills a process and its subprocesses
func Kill(cmd *exec.Cmd) {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, 15)
	}
}
