package lpac

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/damonto/estkme-cloud/internal/config"
	"github.com/damonto/estkme-cloud/internal/driver"
)

var ErrDownloadCancelled = errors.New("download cancelled")

type Cmd struct {
	ctx  context.Context
	APDU driver.APDU
}

func NewCmd(ctx context.Context, APDU driver.APDU) *Cmd {
	return &Cmd{ctx: ctx, APDU: APDU}
}

func (c *Cmd) Run(arguments []string, dst any, progress Progress) error {
	c.APDU.Lock()
	defer c.APDU.Unlock()
	cmd := exec.CommandContext(c.ctx, filepath.Join(config.C.Dir, "lpac"), arguments...)
	cmd.Dir = config.C.Dir
	cmd.Env = append(cmd.Env, "LPAC_APDU=stdio")
	cmd.Env = append(cmd.Env, "LPAC_CUSTOM_ES10X_MSS=240")

	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	stdout, _ := cmd.StdoutPipe()
	stdin, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	cmdErr := c.process(stdout, stdin, dst, progress)
	if err := cmd.Wait(); err != nil {
		slog.Error("command wait error", "error", err, "stderr", stderr.String())
	}
	if cmdErr != nil {
		return fmt.Errorf("%w %s", cmdErr, &stderr)
	}
	return nil
}

func (c *Cmd) process(output io.ReadCloser, input io.WriteCloser, dst any, progress Progress) error {
	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanLines)
	var cmdErr error
	for scanner.Scan() {
		text := scanner.Text()
		slog.Debug("lpac command output", "output", text)
		if err := c.handleOutput(text, input, dst, progress); err != nil {
			cmdErr = err
		}
	}
	return cmdErr
}

func (c *Cmd) handleOutput(output string, input io.WriteCloser, dst any, progress Progress) error {
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

func (c *Cmd) handleLPAResponse(payload json.RawMessage, dst any) error {
	var lpaPayload LPAPyaload
	if err := json.Unmarshal(payload, &lpaPayload); err != nil {
		return err
	}

	if lpaPayload.Code != 0 {
		if lpaPayload.Message == LPADownloadCancelled {
			return ErrDownloadCancelled
		}
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

func (c *Cmd) handleProgress(payload json.RawMessage, progress Progress) error {
	var progressPayload ProgressPayload
	if err := json.Unmarshal(payload, &progressPayload); err != nil {
		return err
	}

	if progressPayload.Message == ProgressMetadataParse {
		profileMetadata := &Profile{}
		if err := json.Unmarshal(progressPayload.Data, profileMetadata); err != nil {
			return err
		}
		return progress(progressPayload.Message, profileMetadata)
	}

	if text, ok := HumanReadableText[progressPayload.Message]; ok {
		return progress(text, nil)
	}
	return nil
}

func (c *Cmd) handleAPDU(payload json.RawMessage, input io.WriteCloser) error {
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
