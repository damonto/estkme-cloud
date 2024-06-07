//go:build windows

package lpac

import (
	"os/exec"
	"syscall"
)

func (c *Cmd) forSystem(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func (c *Cmd) bin() string {
	return "lpac.exe"
}
