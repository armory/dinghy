/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/dinghy/pkg/settings/global"
	"net/http"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type EventClient interface {
	SendEvent(string, *Event)
}

type Client struct {
	Client        *retryablehttp.Client
	Settings      *global.Settings
	Ctx           context.Context
	IsMultiTenant bool
}

type Event struct {
	Start      int64  `json:"start_time"`
	End        int64  `json:"end_time"`
	Org        string `json:"org"`
	Repo       string `json:"repo"`
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	Dinghyfile string `json:"dinghyfile"`
	Module     bool   `json:"is_module"`
}

type details struct {
	Source  string `json:"source"`
	Version string `json:"sourceVersion"`
	Type    string `json:"type"`
}

type payload struct {
	Details details `json:"details"`
	Event   *Event  `json:"content"`
}

func NewEventClient(ctx context.Context, settings *global.Settings, isMultiTennant bool) *Client {
	c := retryablehttp.NewClient()
	c.HTTPClient.Transport = cleanhttp.DefaultPooledTransport() // reuse the client so we can pipeline stuff
	return &Client{
		Client:        c,
		Settings:      settings,
		Ctx:           ctx,
		IsMultiTenant: isMultiTennant,
	}
}

func (c *Client) SendEvent(eventType string, event *Event) {
	payload := payload{
		Details: details{
			Source:  "dinghy",
			Version: c.Settings.Logging.Remote.Version,
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
	var req *retryablehttp.Request
	if c.IsMultiTenant {
		req, err = retryablehttp.NewRequest(http.MethodPost, c.Settings.SpinnakerSupplied.Gate.BaseURL, postData)
	} else {
		req, err = retryablehttp.NewRequest(http.MethodPost, c.Settings.SpinnakerSupplied.Echo.BaseURL, postData)
	}
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(c.Ctx)
	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("debug at %s returned %d", c.Settings.SpinnakerSupplied.Echo.BaseURL, res.StatusCode))
	}
	return nil
}
