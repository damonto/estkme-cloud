package cloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

const (
	ErrInvalidActivationCode   = "invalid activation code"
	ErrRequireConfirmationCode = "confirmation code is required"
	ErrInvalidDataSize         = "invalid data size"

	CmdUseData = "data" // data$<size in KiB>
)

// GSM 7-bit encoding, see https://en.wikipedia.org/wiki/GSM_03.38
var (
	GSMNumberSign = []byte{0x23} // #
	GSMDollarSign = []byte{0x02} // $
)

func handleDownloadProfile(ctx context.Context, conn *Conn, data []byte) error {
	defer conn.Close()
	cmd := strings.ToLower(string(data[:4]))
	if cmd == CmdUseData {
		err := useData(conn, data)
		if err != nil {
			slog.Error("failed to use data", "error", err)
		}
		return err
	}

	conn.Send(TagMessageBox, []byte("Your profile is being downloaded. \n Please wait..."))
	if err := download(ctx, conn, data); err != nil {
		slog.Error("failed to download profile", "error", err)
		return conn.Send(TagMessageBox, []byte("Download failed \n"+ToTitle(err.Error())))
	}
	return conn.Send(TagMessageBox, []byte("Your profile has been downloaded successfully"))
}

func download(ctx context.Context, conn *Conn, data []byte) error {
	var imei string
	var parts [][]byte
	if bytes.Contains(data, GSMNumberSign) {
		index := bytes.Index(data, GSMNumberSign)
		imei = string(data[index+1:])
		parts = bytes.Split(data[:index], GSMDollarSign)
	} else {
		parts = bytes.Split(data, GSMDollarSign)
	}

	if len(parts) < 2 && string(parts[0]) != "LPA:1" {
		return errors.New(ErrInvalidActivationCode)
	}

	var matchingId string
	if len(parts) > 2 {
		matchingId = string(parts[2])
	}

	var confirmationCode string
	if len(parts) == 5 {
		confirmationCode = string(parts[4])
		if confirmationCode == "1" {
			parts[4] = []byte("<confirmation_code>")
			return errors.New(ErrRequireConfirmationCode + "\n" + string(bytes.Join(parts, GSMDollarSign)))
		}
	}

	return lpac.NewCmd(ctx, conn.APDU).ProfileDownload(&lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
		IMEI:             imei,
	}, func(current string) error {
		return conn.Send(TagMessageBox, []byte(current))
	})
}

func useData(conn *Conn, cmd []byte) error {
	var err error
	kb := 1
	if bytes.Contains(cmd, GSMDollarSign) {
		arguments := bytes.Split(cmd, GSMDollarSign)
		if len(arguments) > 1 {
			kb, err = strconv.Atoi(string(arguments[1]))
			if err != nil {
				return err
			}
		}
	}

	if kb > 1024 || kb < 0 {
		return errors.New(ErrInvalidDataSize)
	}

	message := fmt.Sprintf("You used %d KiB data", kb)
	count := kb * 1024 / 256
	for i := 1; i <= count; i++ {
		var placeholder []byte
		if i == count {
			placeholder = bytes.Repeat([]byte{0}, 253-len(message)-len(cmd)-6)
		} else {
			placeholder = bytes.Repeat([]byte{0}, 253)
		}
		if err := conn.Send(TagMessageBox, placeholder); err != nil {
			return err
		}
	}
	return conn.Send(TagMessageBox, []byte(message))
}
