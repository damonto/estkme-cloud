package cloud

import (
	"bytes"
	"context"
	"errors"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

const (
	ErrInvalidActivationCode   = "invalid activation code"
	ErrRequireConfirmationCode = "confirmation code is required\n"
)

// GSM 7-bit encoding, see https://en.wikipedia.org/wiki/GSM_03.38
var (
	GSMNumberSign = []byte{0x23}
	GSMDollarSign = []byte{0x02}
	GSMUnderscore = []byte{0x11}
)

func downloadProfile(ctx context.Context, conn *Conn, data []byte) error {
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
			parts[4] = bytes.Replace([]byte("<confirmation_code>"), []byte("_"), GSMUnderscore, 1)
			return errors.New(ErrRequireConfirmationCode + string(bytes.Join(parts, GSMDollarSign)))
		}
	}

	return lpac.NewCmder(ctx, conn.APDU).ProfileDownload(lpac.ActivationCode{
		SMDP:             string(parts[1]),
		MatchingId:       matchingId,
		ConfirmationCode: confirmationCode,
		IMEI:             imei,
	}, func(current string) error {
		return conn.Send(TagMessageBox, []byte(current))
	})
}
