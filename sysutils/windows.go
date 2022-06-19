//go:build windows

package sysutils

import (
	"log"
	"os/exec"
)

func PrepareCmd(cmd *exec.Cmd) {}

// Kill kills a process
func Kill(cmd *exec.Cmd) {
	err := cmd.Process.Kill()
	if err != nil {
		log.Println(err)
	}
}
