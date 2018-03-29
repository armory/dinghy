package git

// Status wires up to the green check or red x next to a GitHub commit.
type Status string

// Status types
const (
	StatusPending Status = "pending"
	StatusError          = "error"
	StatusSuccess        = "success"
	StatusFailure        = "failure"
)
