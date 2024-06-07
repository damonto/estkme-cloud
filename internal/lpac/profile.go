package lpac

import (
	"errors"
)

type ActivationCode struct {
	SMDP             string
	MatchingId       string
	ConfirmationCode string
	IMEI             string
}

type Profile struct {
	ICCID        string `json:"iccid"`
	ISDPAid      string `json:"isdpAid"`
	State        string `json:"profileState"`
	Nickname     string `json:"profileNickname"`
	ProviderName string `json:"serviceProviderName"`
	ProfileName  string `json:"profileName"`
	IconType     string `json:"iconType"`
	Icon         string `json:"icon"`
	Class        string `json:"profileClass"`
}

type DiscoveryResponse struct {
	RspServerAddress string `json:"rspServerAddress"`
}

const (
	ErrDeletionNotificationNotFound = "deletion notification not found"
)

func (c *Cmd) ProfileList() ([]Profile, error) {
	var profiles []Profile
	if err := c.Run([]string{"profile", "list"}, &profiles, nil); err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (c *Cmd) ProfileInfo(ICCID string) (Profile, error) {
	var profiles []Profile
	if err := c.Run([]string{"profile", "list"}, &profiles, nil); err != nil {
		return Profile{}, err
	}

	for _, profile := range profiles {
		if profile.ICCID == ICCID {
			return profile, nil
		}
	}
	return Profile{}, nil
}

func (c *Cmd) ProfileDownload(activationCode ActivationCode, progress Progress) error {
	arguments := []string{"profile", "download"}
	if activationCode.SMDP != "" {
		arguments = append(arguments, "-s", activationCode.SMDP)
	}
	if activationCode.MatchingId != "" {
		arguments = append(arguments, "-m", activationCode.MatchingId)
	}
	if activationCode.ConfirmationCode != "" {
		arguments = append(arguments, "-c", activationCode.ConfirmationCode)
	}
	if activationCode.IMEI != "" {
		arguments = append(arguments, "-i", activationCode.IMEI)
	}

	return c.sendNotificationAfterDownloading(func() error {
		return c.Run(arguments, nil, progress)
	})
}

func (c *Cmd) sendNotificationAfterDownloading(action func() error) error {
	notifications, err := c.NotificationList()
	if err != nil {
		return err
	}
	var lastSeqNumber int
	for _, notification := range notifications {
		if notification.SeqNumber > lastSeqNumber {
			lastSeqNumber = notification.SeqNumber
		}
	}

	if err := action(); err != nil {
		return err
	}

	notifications, err = c.NotificationList()
	if err != nil {
		return err
	}

	var installationNotificationSeqNumber int
	for _, notification := range notifications {
		if notification.SeqNumber > lastSeqNumber && notification.ProfileManagementOperation == NotificationProfileManagementOperationInstall {
			installationNotificationSeqNumber = notification.SeqNumber
			break
		}
	}
	if installationNotificationSeqNumber > 0 {
		return c.NotificationProcess(installationNotificationSeqNumber, true, nil)
	}
	return nil
}

func (c *Cmd) ProfileDelete(ICCID string) error {
	if err := c.Run([]string{"profile", "delete", ICCID}, nil, nil); err != nil {
		return err
	}

	notifications, err := c.NotificationList()
	if err != nil {
		return err
	}
	var deletionNotificationSeqNumber int
	for _, notification := range notifications {
		if notification.ICCID == ICCID && notification.ProfileManagementOperation == NotificationProfileManagementOperationDelete {
			deletionNotificationSeqNumber = notification.SeqNumber
			break
		}
	}
	if deletionNotificationSeqNumber > 0 {
		return errors.New(ErrDeletionNotificationNotFound)
	}
	return c.NotificationProcess(deletionNotificationSeqNumber, false, nil)
}

func (c *Cmd) ProfileSetNickname(ICCID string, nickname string) error {
	return c.Run([]string{"profile", "nickname", ICCID, nickname}, nil, nil)
}

func (c *Cmd) ProfileDiscovery() ([]DiscoveryResponse, error) {
	var response []DiscoveryResponse
	if err := c.Run([]string{"profile", "discovery"}, &response, nil); err != nil {
		return response, err
	}
	return response, nil
}
