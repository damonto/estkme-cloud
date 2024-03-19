package lpac

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/damonto/estkme-rlpa-server/internal/pkg/config"
	"github.com/damonto/estkme-rlpa-server/internal/pkg/transmitter"
)

type CLI interface {
	Run(arguments []string, dst any, progress Progress) error
}

type cli struct {
	APDU transmitter.APDU
}

func NewCLI(APDU transmitter.APDU) CLI {
	return &cli{APDU: APDU}
}

func (c *cli) Run(arguments []string, dst any, progress Progress) error {
	c.APDU.Lock()
	defer c.APDU.Unlock()
	cmd := exec.Command(c.binName(), arguments...)
	cmd.Dir = config.C.DataDir
	// TODO: Will be deprecated in future versions
	cmd.Env = append(cmd.Env, "APDU_INTERFACE="+filepath.Join(config.C.DataDir, "libapduinterface_stdio"+c.libExtension()))
	// Windows requires libcurl.dll to be in the same directory as the binary
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "LIBCURL"+filepath.Join(config.C.DataDir, "libcurl.dll"))
	}

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

func (c *cli) handleOutput(output string, input io.WriteCloser, dst any, progress Progress) error {
	var commandOutput CommandOutput
	if err := json.Unmarshal([]byte(output), &commandOutput); err != nil {
		return err
	}

	switch commandOutput.Type {
	case CommandStdioAPDU:
		return c.handleAPDU(commandOutput.Payload, input)
	case CommandStdioLPA:
		if dst != nil {
			return json.Unmarshal(commandOutput.Payload, dst)
		}
	case CommandStdioProgress:
		if progress != nil {
			return c.handleProgress(commandOutput.Payload, progress)
		}
	}
	return nil
}

func (c *cli) handleProgress(payload json.RawMessage, progress Progress) error {
	var progressPayload ProgressPayload
	if err := json.Unmarshal(payload, &progressPayload); err != nil {
		return err
	}
	return progress(progressPayload.Message)
}

func (c *cli) handleAPDU(payload json.RawMessage, input io.WriteCloser) error {
	var command CommandAPDUPayload
	if err := json.Unmarshal(payload, &command); err != nil {
		return err
	}

	switch command.Func {
	case CommandAPDUFuncConnect, CommandAPDUFuncDisconnect, CommandAPDUOpenLogicalChannel, CommandAPDUCloseLogicalChannel:
		return json.NewEncoder(input).Encode(CommandAPDUInput{
			Type: CommandStdioAPDU,
			Payload: CommandAPDUInputPayload{
				ECode: 0,
			},
		})
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
	}
	return nil
}

func (c *cli) binName() string {
	var binName string
	switch runtime.GOOS {
	case "windows":
		binName = "lpac.exe"
	default:
		binName = "lpac"
	}
	return filepath.Join(config.C.DataDir, binName)
}

func (c *cli) libExtension() string {
	switch runtime.GOOS {
	case "windows":
		return ".dll"
	case "darwin":
		return ".dylib"
	default:
		return ".so"
	}
}
