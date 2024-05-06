//go:build !windows

package lpac

import (
	"os/exec"
)

func configureSystemOptions(cmd *exec.Cmd) {
	// Do nothing on non-Windows systems.
}

func lpacPath() string {
	return "lpac"
}
