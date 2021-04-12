package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

func snsEvent(message string) events.SNSEvent {
	return events.SNSEvent{
		Records: []events.SNSEventRecord{
			events.SNSEventRecord{
				SNS: events.SNSEntity{
					Message: message,
				},
			},
		},
	}
}

func Test_Handler_Error_Cases(t *testing.T) {
	tests := []struct {
		request events.SNSEvent
	}{
		{ // Test that we capture JSON errors appropriately
			request: snsEvent("not valid json"),
		},
		{ // Test that we fail to post a notification
			// NOTE: we're relying on the fact  that the buster is not valid
			request: snsEvent("{\"environmentId\": \"1\"}"),
		},
	}

	for _, test := range tests {
		err := HandleRequest(context.TODO(), test.request)
		assert.Error(t, err)
	}
}

func Test_Buster_Constructors(t *testing.T) {
	tests := []struct {
		baseURL   string
		shouldErr bool
	}{
		{ // An empty baseURL means we likely can't talk to Dinghy
			baseURL:   "",
			shouldErr: true,
		},
		{ // An invalid url should be the only thing that makes our request fail
			baseURL:   ":/\r",
			shouldErr: true,
		},
		{
			baseURL:   "http://dinghy.armory-hosted-services-staging:8081",
			shouldErr: false,
		},
	}

	for _, test := range tests {
		_, err := NewBuster(test.baseURL)

		if test.shouldErr {
			assert.Error(t, err)
		} else {
			assert.Nil(t, err)
		}
	}

}

func Test_Buster_MakeRequest(t *testing.T) {
	props := gopter.NewProperties(nil)

	buster, _ := NewBuster("http://dinghy.armory-hosted-services-staging:8081")

	props.Property("makeRequest should never fail if passed an environment id", prop.ForAll(
		func(envID string) bool {
			req, err := buster.makeRequest(envID)

			hdr := req.Header.Get(DinghyEnvironmentIDHeader)
			if hdr == envID && err == nil {
				return true
			}

			return false
		},
		gen.AlphaString(),
	))

	props.TestingRun(t)
}
