package lpac

import (
	"errors"
	"math"
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
	ProviderName string `json:"profileProviderName"`
	ProfileName  string `json:"profileName"`
	IconType     string `json:"iconType"`
	Icon         string `json:"icon"`
	Class        string `json:"profileClass"`
}

type Profiles = []Profile

func (c *cli) ProfileList() (Profiles, error) {
	var profiles Profiles
	if err := c.Run([]string{"profile", "list"}, &profiles, nil); err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (c *cli) ProfileInfo(ICCID string) (Profile, error) {
	var profiles Profiles
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

func (c *cli) DownloadProfile(activationCode ActivationCode, progress Progress) error {
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

	return c.sendNotificationAfterDownload(func() error {
		return c.Run(arguments, nil, progress)
	})
}

func (c *cli) sendNotificationAfterDownload(action func() error) error {
	notifications, err := c.NotificationList()
	if err != nil {
		return err
	}

	lastSeqNumber := 0
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

	// Find the notification with the highest sequence number
	installationNotificationSeqNumber := math.MaxInt
	for _, notification := range notifications {
		if notification.SeqNumber > lastSeqNumber && notification.SeqNumber < installationNotificationSeqNumber {
			installationNotificationSeqNumber = notification.SeqNumber
			break
		}
	}
	if installationNotificationSeqNumber != math.MaxInt {
		return c.NotificationProcess(installationNotificationSeqNumber, false, nil)
	}
	return nil
}

func (c *cli) DeleteProfile(ICCID string) error {
	if err := c.Run([]string{"profile", "delete", ICCID}, nil, nil); err != nil {
		return err
	}

	notifications, err := c.NotificationList()
	if err != nil {
		return err
	}

	deletionNotificationSeqNumber := math.MaxInt
	for _, notification := range notifications {
		if notification.ICCID == ICCID && notification.ProfileManagementOperation == "delete" {
			if notification.SeqNumber > deletionNotificationSeqNumber {
				deletionNotificationSeqNumber = notification.SeqNumber
			}
		}
	}
	if deletionNotificationSeqNumber == math.MaxInt {
		return errors.New("deletion notification not found")
	}
	return c.NotificationProcess(deletionNotificationSeqNumber, false, nil)
}
