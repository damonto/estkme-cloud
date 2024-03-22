package rlpa

import (
	"bytes"
	"errors"

	"github.com/damonto/estkme-rlpa-server/internal/lpac"
)

const (
	ErrInvalidActivationCode = "invalid activation code"
	ErrNeedConfirmationCode  = "please enter the confirmation code in the following format. "
)

func downloadProfile(conn *Conn, data []byte) error {
	parts := bytes.Split(data, []byte{0x02})
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
			parts[4] = []byte("ConfirmationCode")
			return errors.New(ErrNeedConfirmationCode + "\n" + string(bytes.Join(parts, []byte{0x02})))
		}
	}

	return lpac.NewCmder(conn.APDU).DownloadProfile(lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
	}, func(current string) error {
		return conn.Send(TagMessageBox, []byte(current))
	})
}
