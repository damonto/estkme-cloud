//go:build windows

package lpac

import (
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/damonto/estkme-cloud/internal/config"
)

func configureSystemOptions(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, "LIBCURL="+filepath.Join(config.C.DataDir, "libcurl.dll"))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func lpacPath() string {
	return "lpac.exe"
}
