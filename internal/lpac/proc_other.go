//go:build !windows

package lpac

import (
	"os/exec"
)

func (c *Cmd) forSystem(cmd *exec.Cmd) {
	//
}

func (c *Cmd) bin() string {
	return "lpac"
}
