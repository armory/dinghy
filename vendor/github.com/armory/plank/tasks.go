package plank

import (
	"errors"
	"time"
)

type Task struct {
	Application string        `json:"application"`
	Description string        `json:"description"`
	Job         []interface{} `json:"job,omitempty"`
}

type TaskRefResponse struct {
	Ref string `json:"ref"`
}

type ExecutionStatusResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	EndTime int    `json:"endTime"`
}

func (c *Client) GetTask(refURL string) (*ExecutionStatusResponse, error) {
	var body ExecutionStatusResponse
	err := c.Get(c.URLs["orca"]+refURL, &body)
	return &body, err
}

func (c *Client) PollTaskStatus(refURL string) (*ExecutionStatusResponse, error) {
	if refURL == "" {
		return nil, errors.New("no taskRef provided to follow")
	}
	timer := time.NewTimer(c.retryIncrement)
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for range t.C {
		var body ExecutionStatusResponse
		c.Get(c.URLs["orca"]+refURL, &body)
		if body.EndTime > 0 {
			return &body, nil
		}
		select {
		case <-timer.C:
			return nil, errors.New("timed out waiting for task to complete")
		default:
		}
	}
	return nil, errors.New("exited poll loop before completion")
}

// Create task puts the payload into the Task wrapper.
func (c *Client) CreateTask(app, desc string, payload interface{}) (*TaskRefResponse, error) {
	task := Task{Application: app, Description: desc, Job: []interface{}{payload}}
	var ref TaskRefResponse
	if err := c.Post(c.URLs["orca"]+"/ops", ApplicationContextJson, task, &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}
