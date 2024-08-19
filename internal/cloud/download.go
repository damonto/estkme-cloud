package cloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

var (
	ErrInvalidActivationCode    = errors.New("invalid activation code")
	ErrInvalidDataSize          = errors.New("invalid data size")
	ErrCommandUnsupported       = errors.New("command unsupported")
	ErrCommandSeqNumberRequired = errors.New("seq number is required")
	ErrNoNotificationFound      = errors.New("no notification found")

	// GSM 7-bit encoding, see https://en.wikipedia.org/wiki/GSM_03.38
	GSMNumberSign = []byte{0x23} // #
	GSMDollarSign = []byte{0x02} // $
	GSMSlashSign  = []byte{0x2F} // /

	CommandConsumeData         = []byte("/data")
	CommandProcessNotification = []byte("/process")
	CommandListNotifications   = []byte("/list")

	CommandArgumentSplitter       = []byte(" ")
	CommandListNotificationsLimit = 5

	ActivationCodeSchemaLPA    = []byte("LPA:")
	ActivationCodeSchemaQRCyou = []byte("qr.esim.cyou/")
)

func handleDownloadProfile(ctx context.Context, conn *Conn, data []byte) error {
	defer conn.Close()
	if bytes.HasPrefix(data, GSMSlashSign) {
		err := handleCommand(ctx, conn, data)
		if err != nil {
			slog.Error("failed to handle command", "error", err)
			return conn.Send(TagMessageBox, []byte("Command failed \n"+ToTitle(err.Error())))
		}
		return err
	}

	conn.Send(TagMessageBox, []byte("Your profile is being downloaded.\nIt may take a few minutes.\nTo avoid download failure, please do not lock your phone."))
	if err := handleCommandDownload(ctx, conn, data); err != nil {
		slog.Error("failed to download profile", "error", err)
		return conn.Send(TagMessageBox, []byte("Download failed \n"+ToTitle(err.Error())))
	}
	return conn.Send(TagMessageBox, []byte("Your profile has been downloaded successfully."))
}

func handleCommandDownload(ctx context.Context, conn *Conn, data []byte) error {
	activationCode, err := decodeActivationCode(data)
	if err != nil {
		return err
	}
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
		return nil
	})
}

func decodeActivationCode(activationCode []byte) (*lpac.ActivationCode, error) {
	if bytes.Contains(activationCode, ActivationCodeSchemaQRCyou) {
		activationCode, _ = bytes.CutPrefix(activationCode, []byte("https://"))
		activationCode, _ = bytes.CutPrefix(activationCode, ActivationCodeSchemaQRCyou)
		if !bytes.HasPrefix(activationCode, ActivationCodeSchemaLPA) {
			activationCode = append(ActivationCodeSchemaLPA, activationCode...)
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

func handleCommand(ctx context.Context, conn *Conn, data []byte) error {
	arguments := bytes.Split(data, CommandArgumentSplitter)
	if bytes.Equal(arguments[0], CommandConsumeData) {
		return handleCommandConsumeData(conn, arguments[1:])
	}
	if bytes.Equal(arguments[0], CommandListNotifications) {
		return handleListNotifications(ctx, conn, arguments[1:])
	}
	if bytes.Equal(arguments[0], CommandProcessNotification) {
		return handleCommandProcessNotification(ctx, conn, arguments[1:])
	}
	return ErrCommandUnsupported
}

func handleListNotifications(ctx context.Context, conn *Conn, arguments [][]byte) error {
	var filterOperation string
	if len(arguments) > 0 {
		filterOperation = string(arguments[0])
	}
	notifications, err := lpac.NewCmd(ctx, conn.APDU).NotificationList()
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		return conn.Send(TagMessageBox, []byte("No notifications found"))
	}
	count := CommandListNotificationsLimit
	message := "Seq ICCID Operation\n"
	slices.Reverse(notifications)
	for _, notification := range notifications {
		if count == 0 {
			break
		}
		if filterOperation != "" && notification.ProfileManagementOperation != filterOperation {
			continue
		}
		message += fmt.Sprintf(
			"%d %s %s\n",
			notification.SeqNumber,
			notification.ICCID[len(notification.ICCID)-4:],
			notification.ProfileManagementOperation,
		)
		count--
	}
	return conn.Send(TagMessageBox, []byte(strings.TrimRight(message, "\n")))
}

func handleCommandProcessNotification(ctx context.Context, conn *Conn, arguments [][]byte) error {
	if len(arguments) == 0 {
		return ErrCommandSeqNumberRequired
	}
	seqNumber, err := strconv.Atoi(string(arguments[0]))
	if err != nil {
		return err
	}
	if err := conn.Send(TagMessageBox, []byte("Processing notification...")); err != nil {
		return err
	}
	if err := lpac.NewCmd(ctx, conn.APDU).NotificationProcess(seqNumber, false, nil); err != nil {
		return err
	}
	return conn.Send(TagMessageBox, []byte("Notification has been processed."))
}

func handleCommandConsumeData(conn *Conn, arguments [][]byte) error {
	var err error
	var kb int
	if len(arguments) > 0 {
		kb, err = strconv.Atoi(string(arguments[0]))
		if err != nil {
			return err
		}
	}

	if kb > 1024 || kb <= 0 {
		return ErrInvalidDataSize
	}

	message := fmt.Sprintf("You used %d KiB data", kb)
	count := kb * 1024 / 256
	for i := 1; i <= count; i++ {
		var placeholder []byte
		if i == count {
			commandLen := len(CommandConsumeData) + len(arguments[0]) + 1
			placeholder = bytes.Repeat([]byte{0}, 253-len(message)-commandLen-6)
		} else {
			placeholder = bytes.Repeat([]byte{0}, 253)
		}
		if err := conn.Send(TagMessageBox, placeholder); err != nil {
			return err
		}
	}
	return conn.Send(TagMessageBox, []byte(message))
}
