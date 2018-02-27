package github

// Status wires up to the green check or red x next to a GitHub commit.
type Status string

const (
	Pending Status = "pending"
	Error          = "error"
	Success        = "success"
	Failure        = "failure"
)
