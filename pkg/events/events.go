package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/armory-io/dinghy/pkg/settings"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Client      *retryablehttp.Client
	LogSettings *settings.RemoteLogging
	Ctx         context.Context
}

type Event struct {
	Start int64  `json:"start_time"`
	End   int64  `json:"end_time"`
	Org   string `json:"org"`
	Repo  string `json:"repo"`
	Path  string `json:"path"`
}

type details struct {
	Source  string `json:"source"`
	Version string `json:"sourceVersion"`
	Type    string `json:"type"`
}

type payload struct {
	CustomerID string  `json:"customer_id"`
	Details    details `json:"details"`
	Event      *Event  `json:"dinghy"`
}

func NewEventClient(ctx context.Context, logSettings *settings.Logging) *Client {
	c := retryablehttp.NewClient()
	c.HTTPClient.Transport = cleanhttp.DefaultPooledTransport() // reuse the client so we can pipeline stuff
	return &Client{
		Client:      c,
		LogSettings: &logSettings.Remote,
		Ctx:         ctx,
	}
}

func (c *Client) SendEvent(eventType string, event *Event) {
	if !c.LogSettings.Enabled {
		return
	}

	payload := payload{
		CustomerID: c.LogSettings.CustomerID,
		Details: details{
			Source:  "dinghy",
			Version: c.LogSettings.Version,
			Type:    eventType,
		},
		Event: event,
	}

	if err := c.postEvent(payload); err != nil {
		log.Errorf(err.Error())
		return
	}
}

func (c *Client) postEvent(event payload) error {
	postData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := retryablehttp.NewRequest(http.MethodPost, c.LogSettings.Endpoint, postData)
	if err != nil {
		return err
	}
	// we need to set Authorization header to talk to debug
	req.Header.Set("Authorization", fmt.Sprintf("Armory %s", c.LogSettings.CustomerID))
	req = req.WithContext(c.Ctx)
	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("debug at %s returned %d", c.LogSettings.Endpoint, res.StatusCode))
	}
	return nil
}
