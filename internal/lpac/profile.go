package lpac

import (
	"errors"
	"log/slog"
)

type ActivationCode struct {
	SMDP             string
	MatchingId       string
	ConfirmationCode string
	IMEI             string
}

type Profile struct {
	ICCID        string       `json:"iccid"`
	ISDPAid      string       `json:"isdpAid"`
	State        ProfileState `json:"profileState"`
	Nickname     string       `json:"profileNickname"`
	ProviderName string       `json:"serviceProviderName"`
	ProfileName  string       `json:"profileName"`
	IconType     string       `json:"iconType"`
	Icon         string       `json:"icon"`
	Class        string       `json:"profileClass"`
}

type DiscoveryResponse struct {
	RspServerAddress string `json:"rspServerAddress"`
}

var (
	ErrDeletionNotificationNotFound = errors.New("deletion notification not found")
	ErrProfileNotFound              = errors.New("profile not found")
)

func (c *Cmd) ProfileList() ([]*Profile, error) {
	var profiles []*Profile
	if err := c.Run([]string{"profile", "list"}, &profiles, nil); err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (c *Cmd) ProfileInfo(ICCID string) (*Profile, error) {
	var profiles []*Profile
	if err := c.Run([]string{"profile", "list"}, &profiles, nil); err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		if profile.ICCID == ICCID {
			return profile, nil
		}
	}
	return nil, ErrProfileNotFound
}

func (c *Cmd) ProfileDownload(activationCode *ActivationCode, progress Progress) error {
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

	return c.sendNotificationAfterExecution(func() error {
		return c.Run(arguments, nil, progress)
	}, true)
}

func (c *Cmd) sendNotificationAfterExecution(action func() error, remove bool) error {
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
	for _, notification := range notifications {
		if notification.SeqNumber > lastSeqNumber {
			slog.Debug("processing notification", "seqNumber", notification.SeqNumber, "ICCID", notification.ICCID, "operation", notification.ProfileManagementOperation, "remove", remove)
			if err := c.NotificationProcess(notification.SeqNumber, remove, nil); err != nil {
				slog.Error("failed to process notification", "seqNumber", notification.SeqNumber, "ICCID", notification.ICCID, "operation", notification.ProfileManagementOperation, "remove", remove, "error", err)
				return err
			}
		}
	}
	return nil
}

func (c *Cmd) ProfileDelete(ICCID string) error {
	return c.sendNotificationAfterExecution(func() error {
		return c.Run([]string{"profile", "delete", ICCID}, nil, nil)
	}, false)
}

func (c *Cmd) ProfileEnable(ICCID string) error {
	return c.sendNotificationAfterExecution(func() error {
		return c.Run([]string{"profile", "enable", ICCID}, nil, nil)
	}, true)
}

func (c *Cmd) ProfileSetNickname(ICCID string, nickname string) error {
	return c.Run([]string{"profile", "nickname", ICCID, nickname}, nil, nil)
}

func (c *Cmd) ProfileDiscovery() ([]*DiscoveryResponse, error) {
	var response []*DiscoveryResponse
	if err := c.Run([]string{"profile", "discovery"}, &response, nil); err != nil {
		return response, err
	}
	return response, nil
}
