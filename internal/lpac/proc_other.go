//go:build !windows

package lpac

import (
	"os/exec"
)

func (c *Cmder) forSystem(cmd *exec.Cmd) {
	//
}

func (c *Cmder) bin() string {
	return "lpac"
}
