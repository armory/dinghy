package notifiers

import (
	"fmt"
	plank "github.com/armory/plank/v3"
	"github.com/mitchellh/mapstructure"
)

type ApplicationNotification interface {
	GetWhen()	  []string
	GetLevel()	  string
	GetAddress()  string
}

func ToNotifications(notificationsType plank.NotificationsType) ([]ApplicationNotification, error) {
	var result []ApplicationNotification
	for key, valueSlice := range notificationsType {
		switch key {
		case "slack":
			if slackValues, ok := valueSlice.([]interface{}); ok {
				for _, currentSlack := range slackValues {
					var slackNotifValues SlackNotificationValues
					if err := mapstructure.Decode(currentSlack, &slackNotifValues); err != nil {
						return nil, fmt.Errorf("Failed to decode map to SlackNotificationValues %v", err.Error())
					}
					result = append(result, &slackNotifValues)
				}
			}
		}
	}
	return result, nil
}