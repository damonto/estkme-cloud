//go:build windows

package lpac

import (
	"os/exec"
	"syscall"
)

func (c *Cmder) forSystem(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func (c *Cmder) bin() string {
	return "lpac.exe"
}
