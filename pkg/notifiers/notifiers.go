package notifiers

type Notifier interface {
	SendSuccess(org, repo, path string)
	SendFailure(org, repo, path string, err error)
}
