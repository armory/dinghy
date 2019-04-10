package plank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

// Client for working with API servers that accept and return JSON payloads.
type Client struct {
	http           *http.Client
	retryIncrement time.Duration
	maxRetry       int
	URLs           map[string]string
}

type ContentType string

const (
	ApplicationJson        ContentType = "application/json"
	ApplicationContextJson ContentType = "application/context+json"
)

// DefaultURLs
var DefaultURLs = map[string]string{
	"orca":    "http://armory-orca:8083",
	"front50": "http://armory-front50:8080",
	"fiat":    "http://armory-fiat:7003",
}

// New constructs a Client using the given http.Client-compatible client.
// Pass nil if you want to just use the default http.Client structure.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
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
	resp, err := c.http.Get(url)
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
	resp, err := c.http.Post(url, string(contentType), bytes.NewBuffer(jsonBody))
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
