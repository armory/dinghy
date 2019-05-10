/*
 * Copyright 2019 Armory, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package plank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"runtime"
	"time"
)

// Client for working with API servers that accept and return JSON payloads.
type Client struct {
	http           *http.Client
	retryIncrement time.Duration
	maxRetry       int
	URLs           map[string]string
	FiatUser       string
}

type ContentType string

type ClientOption func(*Client)

func WithClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.http = client
	}
}

func WithTransport(transport *http.Transport) ClientOption {
	return func(c *Client) {
		c.http.Transport = transport
	}
}

func WithRetryIncrement(t time.Duration) ClientOption {
	return func(c *Client) {
		c.retryIncrement = t
	}
}

func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetry = retries
	}
}

func WithFiatUser(user string) ClientOption {
	return func(c *Client) {
		c.FiatUser = user
	}
}

func WithURLs(urls map[string]string) ClientOption {
	return func(c *Client) {
		c.URLs = make(map[string]string)
		for k, v := range urls {
			c.URLs[k] = v
		}
	}
}

const (
	ApplicationJson        ContentType = "application/json"
	ApplicationContextJson ContentType = "application/context+json"
)

// DefaultURLs
var DefaultURLs = map[string]string{
	"orca":    "http://armory-orca:8083",
	"front50": "http://armory-front50:8080",
	"fiat":    "http://armory-fiat:7003",
	"gate":    "http://armory-gate:8084",
}

// New constructs a Client using a default client and sane non-shared http transport
func New(opts ...ClientOption) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
	c := &Client{
		http:           httpClient,
		retryIncrement: 100,
		maxRetry:       20,
		URLs:           make(map[string]string),
	}
	// Have to manually copy the DefaultURLs map because otherwise any changes
	// made to this copy will modify the global.  I can't believe I have to
	// to do this in this day and age...
	for k, v := range DefaultURLs {
		c.URLs[k] = v
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type RequestFunction func() error

func (c *Client) RequestWithRetry(f RequestFunction) error {
	var err error
	for i := 0; i <= c.maxRetry; i++ {
		err = f()
		if err == nil {
			return nil // Success (or non-failure)
		}
		// exponential back-off
		interval := c.retryIncrement * time.Duration(math.Exp2(float64(i)))
		time.Sleep(interval)
	}
	// We get here, we timed out without getting a valid response.
	return err
}

// Get a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Get(url string, dest interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if c.FiatUser != "" {
		req.Header.Set("X-Spinnaker-User", c.FiatUser)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return err
		}
		return nil
	}
	return errors.New(fmt.Sprintf("Unsupported status code: %d", resp.StatusCode))
}

func (c *Client) GetWithRetry(url string, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Get(url, dest)
	})
}

// Post a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Post(url string, contentType ContentType, body interface{}, dest interface{}) error {
	var err error
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("could not post - body could not be marshaled to json - %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	if c.FiatUser != "" {
		req.Header.Set("X-Spinnaker-User", c.FiatUser)
	}
	req.Header.Set("Content-Type", string(contentType))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Only unmarshalling if we actually have a response body (as opposed
		// to, for example, a simple "201 Created" response)
		if len(b) > 0 {
			err := json.Unmarshal(b, &dest)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Unsupported status code: %d", resp.StatusCode))
}

func (c *Client) PostWithRetry(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Post(url, contentType, body, dest)
	})
}

// Post a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Put(url string, contentType ContentType, body interface{}, dest interface{}) error {
	var err error
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("could not put - body could not be marshaled to json - %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	if c.FiatUser != "" {
		req.Header.Set("X-Spinnaker-User", c.FiatUser)
	}
	req.Header.Set("Content-Type", string(contentType))
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Only unmarshalling if we actually have a response body (as opposed
		// to, for example, a simple "201 Created" response)
		if len(b) > 0 {
			err := json.Unmarshal(b, &dest)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Unsupported status code: %d", resp.StatusCode))
}

func (c *Client) PutWithRetry(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Put(url, contentType, body, dest)
	})
}

func (c *Client) Delete(url string) error {
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		// There is no support for receiving a payload back from a DELETE...
		return nil
	}
	return errors.New(fmt.Sprintf("Unsupported status code: %d", resp.StatusCode))
}

func (c *Client) DeleteWithRetry(url string) error {
	return c.RequestWithRetry(func() error {
		return c.Delete(url)
	})
}
