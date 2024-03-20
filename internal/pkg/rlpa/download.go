package rlpa

import (
	"bytes"

	"github.com/damonto/estkme-rlpa-server/internal/pkg/lpac"
)

func Download(conn *Connection, data []byte) error {
	parts := bytes.Split(data, []byte{0x02})
	if len(parts) < 2 && string(parts[0]) != "LPA:1" {
		return conn.Send(TagMessageBox, []byte("Invalid activation code"))
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
			return conn.Send(TagMessageBox, []byte(
				"Need a confirmation code. \n Please enter the confirmation code in the following format. \n"+
					string(bytes.Join(parts, []byte{0x02})),
			))
		}
	}

	return lpac.NewCLI(conn.APDU).DownloadProfile(lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
	}, func(current string) error {
		return conn.Send(TagMessageBox, []byte(current))
	})
}
