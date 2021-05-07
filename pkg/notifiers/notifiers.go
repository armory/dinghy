package notifiers

import "github.com/armory/plank/v4"

type Notifier interface {
	SendSuccess(org, repo, path string, notifications plank.NotificationsType, content map[string]interface{})
	SendFailure(org, repo, path string, err error, notifications plank.NotificationsType, content map[string]interface{})
	SendOnValidation() bool
}
