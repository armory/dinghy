package notifiers

import "github.com/armory/plank/v3"

type Notifier interface {
	SendSuccess(org, repo, path string, notifications plank.NotificationsType)
	SendFailure(org, repo, path string, err error, notifications plank.NotificationsType)
}
