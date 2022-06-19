//go:build !windows

package sysutils

import (
	"log"
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
		err = syscall.Kill(-pgid, 15)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}
