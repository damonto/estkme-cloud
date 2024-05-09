//go:build !windows

package lpac

import (
	"os"
	"os/exec"
	"syscall"
)

func (c *Cmder) forSystem(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}
}

func (c *Cmder) bin() string {
	return "lpac"
}

func (c *Cmder) interrupt(cmd *exec.Cmd) error {
	return cmd.Process.Signal(os.Interrupt)
}
