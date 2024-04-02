package lpac

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/damonto/estkme-rlpa-server/internal/config"
	"github.com/damonto/estkme-rlpa-server/internal/driver"
)

type Cmder struct {
	APDU driver.APDU
}

func NewCmder(APDU driver.APDU) *Cmder {
	return &Cmder{APDU: APDU}
}

func (c *Cmder) Run(arguments []string, dst any, progress Progress) error {
	c.APDU.Lock()
	defer c.APDU.Unlock()
	cmd := exec.Command(c.binName(), arguments...)
	cmd.Dir = config.C.DataDir
	cmd.Env = append(cmd.Env, "LPAC_APDU=stdio")
	// Windows requires libcurl.dll to be in the same directory as the binary
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "LIBCURL="+filepath.Join(config.C.DataDir, "libcurl.dll"))
	}

	// We don't need check the error output, because we are using the stdio interface. (most of the time, the error output is empty.)
	stdout, _ := cmd.StdoutPipe()
	defer stdout.Close()
	stdin, _ := cmd.StdinPipe()
	defer stdin.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		output := scanner.Text()
		if err := c.handleOutput(output, stdin, dst, progress); err != nil {
			return err
		}
	}
	return cmd.Wait()
}

func (c *Cmder) binName() string {
	var binName string
	switch runtime.GOOS {
	case "windows":
		binName = "lpac.exe"
	default:
		binName = "lpac"
	}
	return filepath.Join(config.C.DataDir, binName)
}

func (c *Cmder) handleOutput(output string, input io.WriteCloser, dst any, progress Progress) error {
	var commandOutput CommandOutput
	if err := json.Unmarshal([]byte(output), &commandOutput); err != nil {
		return err
	}

	switch commandOutput.Type {
	case CommandStdioAPDU:
		return c.handleAPDU(commandOutput.Payload, input)
	case CommandStdioLPA:
		return c.handleLPAResponse(commandOutput.Payload, dst)
	case CommandStdioProgress:
		if progress != nil {
			return c.handleProgress(commandOutput.Payload, progress)
		}
	}
	return nil
}

func (c *Cmder) handleLPAResponse(payload json.RawMessage, dst any) error {
	var lpaPayload LPAPyaload
	if err := json.Unmarshal(payload, &lpaPayload); err != nil {
		return err
	}

	if lpaPayload.Code != 0 {
		var errorMessage string
		if err := json.Unmarshal(lpaPayload.Data, &errorMessage); err != nil {
			return errors.New(lpaPayload.Message)
		}
		if errorMessage == "" {
			return errors.New(lpaPayload.Message)
		}
		return errors.New(errorMessage)
	}
	if dst != nil {
		return json.Unmarshal(lpaPayload.Data, dst)
	}
	return nil
}

func (c *Cmder) handleProgress(payload json.RawMessage, progress Progress) error {
	var progressPayload ProgressPayload
	if err := json.Unmarshal(payload, &progressPayload); err != nil {
		return err
	}
	if humanReadableMessage, ok := ProgressMessages[progressPayload.Message]; ok {
		return progress(humanReadableMessage)
	}
	return progress(progressPayload.Message)
}

func (c *Cmder) handleAPDU(payload json.RawMessage, input io.WriteCloser) error {
	var command CommandAPDUPayload
	if err := json.Unmarshal(payload, &command); err != nil {
		return err
	}

	switch command.Func {
	case CommandAPDUFuncTransmit:
		received, err := c.APDU.Transmit(command.Param)
		if err != nil {
			return err
		}
		return json.NewEncoder(input).Encode(CommandAPDUInput{
			Type: CommandStdioAPDU,
			Payload: CommandAPDUInputPayload{
				Data: received,
			},
		})
	default:
		return json.NewEncoder(input).Encode(CommandAPDUInput{
			Type: CommandStdioAPDU,
			Payload: CommandAPDUInputPayload{
				ECode: 0,
			},
		})
	}
}
