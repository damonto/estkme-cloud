package lpac

import "strconv"

type Notification struct {
	SeqNumber                  int    `json:"seqNumber"`
	ProfileManagementOperation string `json:"profileManagementOperation"`
	NotificationAddress        string `json:"notificationAddress"`
	ICCID                      string `json:"iccid"`
}

type Notifications = []Notification

const (
	NotificationProfileManagementOperationDisable   = "disable"
	NotificationProfileManagementOperationEnable    = "enable"
	NotificationProfileManagementOperationInstall = "install"
	NotificationProfileManagementOperationDelete   = "delete"
)

func (c *Cmder) NotificationList() (Notifications, error) {
	var notifications Notifications
	if err := c.Run([]string{"notification", "list"}, &notifications, nil); err != nil {
		return notifications, err
	}
	return notifications, nil
}

func (c *Cmder) NotificationProcess(seqNumber int, remove bool, progress Progress) error {
	arguments := []string{"notification", "process", strconv.Itoa(seqNumber)}
	if remove {
		arguments = append(arguments, "-r")
	}
	return c.Run(arguments, nil, progress)
}

func (c *Cmder) NotificationDelete(seqNumber int) error {
	return c.Run([]string{"notification", "delete", strconv.Itoa(seqNumber)}, nil, nil)
}

func (c *Cmder) NotificationPurge() error {
	notifications, err := c.NotificationList()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := c.NotificationDelete(notification.SeqNumber); err != nil {
			return err
		}
	}
	return nil
}
