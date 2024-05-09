package cloud

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

func processNotification(ctx context.Context, conn *Conn) error {
	cmder := lpac.NewCmder(ctx, conn.APDU)
	notifications, err := cmder.NotificationList()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := cmder.NotificationProcess(notification.SeqNumber, notification.ProfileManagementOperation != lpac.NotificationProfileManagementOperationDelete, nil); err != nil {
			slog.Error("error processing notification", "seqNumber", notification.SeqNumber, "ICCID", notification.ICCID, "operation", notification.ProfileManagementOperation, "error", err)
			if err := conn.Send(TagMessageBox, []byte(fmt.Sprintf("Process notification %d failed\n%s", notification.SeqNumber, err.Error()))); err != nil {
				return err
			}
		}
		slog.Info("notification processed", "seqNumber", notification.SeqNumber, "iccid", notification.ICCID, "operation", notification.ProfileManagementOperation)
		if err := conn.Send(TagMessageBox, []byte(fmt.Sprintf("Notification has been processed. \n Seq number: %d \n ICCID: %s \n Operation: %s", notification.SeqNumber, notification.ICCID, notification.ProfileManagementOperation))); err != nil {
			return err
		}
	}
	return nil
}
