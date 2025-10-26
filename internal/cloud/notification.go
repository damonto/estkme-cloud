package cloud

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/damonto/estkme-cloud/internal/lpac"
)

func handleProcessNotification(ctx context.Context, conn *Conn, _ []byte) error {
	defer conn.Close()
	conn.Send(TagMessageBox, []byte("Processing notifications..."))
	if err := processNotification(ctx, conn); err != nil {
		slog.Error("failed to process notification", "error", err)
		return conn.Send(TagMessageBox, []byte("Process failed \n"+ToTitle(err.Error())))
	}
	return conn.Send(TagMessageBox, []byte("All notifications have been processed successfully."))
}

func processNotification(ctx context.Context, conn *Conn) error {
	cmd := lpac.NewCmd(ctx, conn.APDU)
	notifications, err := cmd.NotificationList()
	if err != nil {
		return err
	}
	for _, notification := range notifications {
		if err := cmd.NotificationProcess(notification.SeqNumber, notification.ProfileManagementOperation != lpac.NotificationProfileManagementOperationDelete, nil); err != nil {
			slog.Error("error processing notification", "seqNumber", notification.SeqNumber, "ICCID", notification.ICCID, "operation", notification.ProfileManagementOperation, "error", err)
			if err := conn.Send(TagMessageBox, fmt.Appendf(nil, "Process notification %d failed\n%s", notification.SeqNumber, err.Error())); err != nil {
				return err
			}
		}
		if notification.ProfileManagementOperation != lpac.NotificationProfileManagementOperationDelete {
			if err := cmd.NotificationDelete(notification.SeqNumber); err != nil {
				slog.Error("error deleting notification", "seqNumber", notification.SeqNumber, "error", err)
				if err := conn.Send(TagMessageBox, fmt.Appendf(nil, "Delete notification %d failed\n%s", notification.SeqNumber, err.Error())); err != nil {
					return err
				}
			}
		}
		slog.Info("notification processed", "seqNumber", notification.SeqNumber, "iccid", notification.ICCID, "operation", notification.ProfileManagementOperation)
		if err := conn.Send(TagMessageBox, fmt.Appendf(nil,
			"Notification has been processed. \n Seq number: %d \n ICCID: %s \n Operation: %s",
			notification.SeqNumber,
			notification.ICCID,
			notification.ProfileManagementOperation,
		)); err != nil {
			return err
		}
	}
	return nil
}
