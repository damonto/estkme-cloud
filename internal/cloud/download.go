package cloud

import (
	"bytes"
	"errors"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

const (
	ErrInvalidActivationCode   = "invalid activation code"
	ErrRequireConfirmationCode = "confirmation code is required\n"
)

func downloadProfile(conn *Conn, data []byte) error {
	var imei string
	var parts [][]byte
	if bytes.Contains(data, []byte{0x23}) {
		index := bytes.Index(data, []byte{0x23})
		imei = string(data[index+1:])
		parts = bytes.Split(data[:index], []byte{0x02})
	} else {
		parts = bytes.Split(data, []byte{0x02})
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
			parts[4] = bytes.Replace([]byte("<confirmation_code>"), []byte("_"), []byte{0x11}, 1)
			return errors.New(ErrRequireConfirmationCode + string(bytes.Join(parts, []byte{0x02})))
		}
	}

	return lpac.NewCmder(conn.APDU).ProfileDownload(lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
		IMEI:             imei,
	}, func(current string) error {
		return conn.Send(TagMessageBox, []byte(current))
	})
}
