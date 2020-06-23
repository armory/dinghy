/*
 * Copyright 2020 Armory, Inc.
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
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	// SpinFiatUserHeader is the header name used for representing users.
	SpinFiatUserHeader string = "X-Spinnaker-User"
	// SpinFiatAccountHeader is the header name used for representing accounts.
	SpinFiatAccountHeader string = "X-Spinnaker-Accounts"

	ApplicationJson        ContentType = "application/json"
	ApplicationContextJson ContentType = "application/context+json"
)

type ErrUnsupportedStatusCode struct {
	Code int
}

func (e *ErrUnsupportedStatusCode) Error() string {
	return fmt.Sprintf("unsupported status code: %d", e.Code)
}

// Client for working with API servers that accept and return JSON payloads.
type Client struct {
	http            *http.Client
	retryIncrement  time.Duration
	maxRetry        int
	URLs            map[string]string
	FiatUser        string
	ArmoryEndpoints bool
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

func WithOverrideAllURLs(urls map[string]string) ClientOption {
	return func(c *Client) {
		c.URLs = make(map[string]string)
		for k, v := range urls {
			c.URLs[k] = v
		}
	}
}

// DefaultURLs
var DefaultURLs = map[string]string{
	"orca":    "http://localhost:8083",
	"front50": "http://localhost:8080",
	"fiat":    "http://localhost:7003",
	"gate":    "http://localhost:8084",
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

// FailedResponse captures a 4xx/5xx response from the upstream Spinnaker service.
// It is expected that the caller destructures the response according to the structure they expect.
type FailedResponse struct {
	Response   []byte
	StatusCode int
}

func (e *FailedResponse) Error() string {
	return fmt.Sprintf("%v: %s", http.StatusText(e.StatusCode), string(e.Response))
}

// Method represents a supported HTTP Method in Plank.
type Method string

const (
	// Patch is a PATCH HTTP method
	Patch Method = http.MethodPatch
	// Post is a POST HTTP method
	Post Method = http.MethodPost
	// Put is a PUT HTTP method
	Put Method = http.MethodPut
	// Get is a GET HTTP method
	Get Method = http.MethodGet
)

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

func (c *Client) request(method Method, url string, contentType ContentType, body interface{}, dest interface{}) error {
	jsonBody, err := json.Marshal(body)

	if err != nil {
		return fmt.Errorf("could not %q, body could not be marshalled to json: %w", string(method), err)
	}

	req, err := http.NewRequest(string(method), url, bytes.NewBuffer(jsonBody))

	if err != nil {
		return fmt.Errorf("could not %q, new request could not be made: %w", string(method), err)
	}

	if c.FiatUser != "" {
		req.Header.Set(SpinFiatUserHeader, c.FiatUser)
		req.Header.Set(SpinFiatAccountHeader, c.FiatUser)
	}

	req.Header.Set("Content-Type", string(contentType))

	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if string(b) == "" {
			dest = make(map[string]interface{})
			return nil
		}
		return json.Unmarshal(b, dest)
	} else if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		return &FailedResponse{StatusCode: resp.StatusCode, Response: b}
	} else {
		// If status falls outside the range of 200 - 599 then return an error.
		return &ErrUnsupportedStatusCode{Code: resp.StatusCode}
	}
}

// Get a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Get(url string, dest interface{}) error {
	return c.request(Get, url, ApplicationJson, nil, dest)
}

func (c *Client) GetWithRetry(url string, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Get(url, dest)
	})
}

// Patch updates a resource for the target URL
func (c *Client) Patch(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.request(Patch, url, contentType, body, dest)
}

// PatchWithRetry updates a resource for the target URL
func (c *Client) PatchWithRetry(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Patch(url, contentType, body, dest)
	})
}

// Post a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Post(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.request(Post, url, contentType, body, dest)
}

func (c *Client) PostWithRetry(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.RequestWithRetry(func() error {
		return c.Post(url, contentType, body, dest)
	})
}

// Post a JSON payload from the URL then decode it into the 'dest' arguement.
func (c *Client) Put(url string, contentType ContentType, body interface{}, dest interface{}) error {
	return c.request(Put, url, contentType, body, dest)
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
	return &ErrUnsupportedStatusCode{Code: resp.StatusCode}
}

func (c *Client) DeleteWithRetry(url string) error {
	return c.RequestWithRetry(func() error {
		return c.Delete(url)
	})
}

func (c *Client) ArmoryEndpointsEnabled() bool {
	return c.ArmoryEndpoints
}

func (c *Client) EnableArmoryEndpoints() {
	c.ArmoryEndpoints = true
}

func (c *Client) DisableARmoryEndpoints() {
	c.ArmoryEndpoints = false
}
