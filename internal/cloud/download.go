package cloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

var (
	ErrInvalidActivationCode = errors.New("invalid activation code")
	ErrInvalidDataSize       = errors.New("invalid data size")

	// GSM 7-bit encoding, see https://en.wikipedia.org/wiki/GSM_03.38
	GSMNumberSign = []byte{0x23} // #
	GSMDollarSign = []byte{0x02} // $

	ActivationCodeSchemaAC     = []byte("LPA:")
	ActivationCodeSchemaQRCyou = []byte("https://qr.esim.cyou/")
)

func handleDownloadProfile(ctx context.Context, conn *Conn, data []byte) error {
	defer conn.Close()
	if bytes.HasPrefix(data, []byte("data:")) {
		err := useData(conn, data)
		if err != nil {
			slog.Error("failed to use data", "error", err)
		}
		return err
	}

	activationCode, err := decodeActivationCode(data)
	if err != nil {
		return err
	}
	conn.Send(TagMessageBox, []byte("Your profile is being downloaded. \n Please wait..."))
	if err := download(ctx, conn, activationCode); err != nil {
		slog.Error("failed to download profile", "error", err)
		return conn.Send(TagMessageBox, []byte("Download failed \n"+ToTitle(err.Error())))
	}
	return conn.Send(TagMessageBox, []byte("Your profile has been downloaded successfully."))
}

func decodeActivationCode(activationCode []byte) (*lpac.ActivationCode, error) {
	if bytes.HasPrefix(activationCode, ActivationCodeSchemaQRCyou) {
		activationCode, _ = bytes.CutPrefix(activationCode, ActivationCodeSchemaQRCyou)
		if !bytes.HasPrefix(activationCode, ActivationCodeSchemaAC) {
			activationCode = append(ActivationCodeSchemaAC, activationCode...)
		}
	}

	var imei string
	var parts [][]byte
	if bytes.Contains(activationCode, GSMNumberSign) {
		index := bytes.Index(activationCode, GSMNumberSign)
		imei = string(activationCode[index+1:])
		parts = bytes.Split(activationCode[:index], GSMDollarSign)
	} else {
		parts = bytes.Split(activationCode, GSMDollarSign)
		fmt.Println("activationCode", activationCode, "parts", parts)
	}

	if len(parts) < 2 && string(parts[0]) != "LPA:1" {
		return nil, ErrInvalidActivationCode
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
			return nil, errors.New("confirmation code is required" + "\n" + string(bytes.Join(parts, GSMDollarSign)))
		}
	}
	return &lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
		IMEI:             imei,
	}, nil
}

func download(ctx context.Context, conn *Conn, activationCode *lpac.ActivationCode) error {
	return lpac.NewCmd(ctx, conn.APDU).ProfileDownload(activationCode, func(current string, profileMetadata *lpac.Profile) error {
		if current == lpac.ProgressMetadataParse {
			template := `
Downloading Profile...
Provider Name: %s
Profile Name: %s
ICCID: %s
`
			return conn.Send(TagMessageBox, []byte(fmt.Sprintf(template, profileMetadata.ProviderName, profileMetadata.ProfileName, profileMetadata.ICCID)))
		}
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
		return ErrInvalidDataSize
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
