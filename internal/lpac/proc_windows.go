//go:build windows

package lpac

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/damonto/estkme-cloud/internal/config"
)

func (c *Cmder) forSystem(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, "LIBCURL="+filepath.Join(config.C.DataDir, "libcurl.dll"))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func (c *Cmder) bin() string {
	return "lpac.exe"
}

func (c *Cmder) interrupt(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
