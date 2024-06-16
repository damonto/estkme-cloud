package lpac

import (
	"sort"
	"strconv"
)

type Notification struct {
	SeqNumber                  int    `json:"seqNumber"`
	ProfileManagementOperation string `json:"profileManagementOperation"`
	NotificationAddress        string `json:"notificationAddress"`
	ICCID                      string `json:"iccid"`
}

const (
	NotificationProfileManagementOperationDisable = "disable"
	NotificationProfileManagementOperationEnable  = "enable"
	NotificationProfileManagementOperationInstall = "install"
	NotificationProfileManagementOperationDelete  = "delete"
)

func (c *Cmd) NotificationList() ([]*Notification, error) {
	var notifications []*Notification
	if err := c.Run([]string{"notification", "list"}, &notifications, nil); err != nil {
		return notifications, err
	}
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].SeqNumber < notifications[j].SeqNumber
	})
	return notifications, nil
}

func (c *Cmd) NotificationProcess(seqNumber int, remove bool, progress Progress) error {
	arguments := []string{"notification", "process", strconv.Itoa(seqNumber)}
	if remove {
		arguments = append(arguments, "-r")
	}
	return c.Run(arguments, nil, progress)
}

func (c *Cmd) NotificationDelete(seqNumber int) error {
	return c.Run([]string{"notification", "remove", strconv.Itoa(seqNumber)}, nil, nil)
}

func (c *Cmd) NotificationPurge() error {
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
