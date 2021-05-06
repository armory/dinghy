package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// DinghyURL is the location of our Dinghy service.
	DinghyURL string = os.Getenv("DINGHY_URL")
	// DinghyEnvironmentIDHeader is the name of the header to pass along envID.
	DinghyEnvironmentIDHeader string = "X-Environment-ID"
)

// Buster is our client for expiring Config records in Dinghy.
type Buster struct {
	baseURL string
	client  *http.Client
}

// NewBuster creates a new Buster instance.
func NewBuster(baseURL string) (*Buster, error) {
	if baseURL == "" {
		return nil, errors.New("cannot create a cache buster without a baseURL")
	}

	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("could not assign baseURL: %w", err)
	}

	return &Buster{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}, nil
}

func (b *Buster) makeRequest(envID string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/config/cachebust", b.baseURL), nil)

	if err == nil {
		req.Header.Add(DinghyEnvironmentIDHeader, envID)
	}

	return req, err
}

// Bust expires a Config record from Dinghy's cache.
func (b *Buster) Bust(environmentID string) error {

	req, err := b.makeRequest(environmentID)
	if err != nil {
		return fmt.Errorf("could not create http request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not complete request to Dinghy: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not bust cache for %q. is environmentId actually defined?", environmentID)
	}

	return nil
}

// YetiNotification represents the customer environment that was changed.
type YetiNotification struct {
	// EnvironmentId represents the customer environment that was changed.
	EnvironmentID string `json:"environmentId"`
}

// HandleRequest handles a Lambda request.
func HandleRequest(ctx context.Context, event events.SNSEvent) error {

	buster, err := NewBuster(DinghyURL)

	if err != nil {
		return fmt.Errorf("could not create cache bust client: %w", err)
	}

	errs := []error{}
	for _, record := range event.Records {
		notification := YetiNotification{}
		err := json.Unmarshal([]byte(record.SNS.Message), &notification)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not unmarshal notification: %w", err))
			continue
		}

		err = buster.Bust(notification.EnvironmentID)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not post notification: %w", err))
			continue
		}
	}

	if len(errs) > 0 {
		errMsg := ""
		for _, e := range errs {
			errMsg = errMsg + "\n" + e.Error()
		}
		return errors.New(errMsg)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
