package spinnaker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"time"
)

// Task is the structure posted to the task endpoint in Spinnaker
type Task struct {
	Application string        `json:"application"`
	Description string        `json:"description"`
	Job         []interface{} `json:"job,omitempty"`
}

// TaskRefResponse represents a task ID URL response following a submitted
// orchestration.
type TaskRefResponse struct {
	Ref string `json:"ref"`
}

// ExecutionResponse wraps the generic response format of an orchestration
// execution.
type ExecutionResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"string"`
	Application string              `json:"application"`
	Status      string              `json:"status"`
	BuildTime   int                 `json:"buildTime"`
	StartTime   int                 `json:"startTime"`
	EndTime     int                 `json:"endTime"`
	Execution   interface{}         `json:"execution"`
	Steps       []ExecutionStep     `json:"steps"`
	Variables   []ExecutionVariable `json:"variables"`
}

// ExecutionStep partially represents a single Orca execution step.
type ExecutionStep struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime int    `json:"startTime"`
	EndTime   int    `json:"endTime"`
	Status    string `json:"status"`
}

// ExecutionVariable represents a variable key/value pair from an execution.
type ExecutionVariable struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// RetrofitErrorResponse represents a Retrofit error.
type RetrofitErrorResponse struct {
	Error        string   `mapstructure:"error"`
	Errors       []string `mapstructure:"errors"`
	Kind         string   `mapstructure:"kind"`
	ResponseBody string   `mapstructure:"responseBody"`
	Status       int      `mapstructure:"status"`
	URL          string   `mapstructure:"url"`
}

type exceptionVariable struct {
	Details RetrofitErrorResponse `mapstructure:"details"`
}

func submitTask(task Task) (*TaskRefResponse, error) {
	path := fmt.Sprintf("%s/applications/%s/tasks", settings.S.SpinnakerAPIURL, task.Application)
	b, err := json.Marshal(task)
	if err != nil {
		log.Error("Could not marshal pipeline ", err)
		return nil, err
	}
	resp, err := postWithRetry(path, b)
	if err != nil {
		return nil, err
	}
	var ref TaskRefResponse
	util.ReadJSON(resp.Body, &ref)
	return &ref, nil
}

func pollTaskStatus(refURL string, timeout time.Duration) (*ExecutionResponse, error) {
	timer := time.NewTimer(timeout)
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for range t.C {
		resp, err := getTask(refURL)
		if err != nil {
			return nil, err
		}
		if resp.EndTime > 0 {
			return resp, nil
		}
		select {
		case <-timer.C:
			return nil, errors.New("timed out waiting for task to complete")
		default:
			log.WithField("status", resp.Status).Debug("Polling task")
		}
	}
	return nil, errors.New("exited poll loop before completion")
}

func getTask(refURL string) (*ExecutionResponse, error) {
	resp, err := getWithRetry(fmt.Sprintf("%s/%s", settings.S.SpinnakerAPIURL, refURL))
	if err != nil {
		return nil, fmt.Errorf("error getting task status %v", err)
	}
	defer resp.Body.Close()
	var task ExecutionResponse
	util.ReadJSON(resp.Body, &task)
	return &task, nil
}

func (e ExecutionResponse) ExtractRetrofitError() *RetrofitErrorResponse {
	for _, v := range e.Variables {
		if v.Key == "exception" {
			var exception exceptionVariable
			if err := mapstructure.Decode(v.Value, &exception); err != nil {
				log.WithError(err).Fatal("could not decode exception struct")
			}
			return &exception.Details
		}
	}
	return nil
}
