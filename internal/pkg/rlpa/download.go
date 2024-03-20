package rlpa

import (
	"bytes"
	"fmt"
)

func Download(conn *Connection, data []byte) error {
	parts := bytes.Split(data, []byte{0x02})
	if len(parts) < 3 {
		return conn.Send(TagMessageBox, []byte("Invalid activation code"))
	}

	var matchingId string
	if len(parts) > 3 {
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

	fmt.Println(parts[1], matchingId, confirmationCode)
	return nil
}
