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

package web

import (
	"bytes"
	// "errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/armory/dinghy/pkg/mock"

	// "github.com/armory/dinghy/pkg/settings"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRouterSanity(t *testing.T) {
	wa := &WebAPI{}
	r := wa.Router()
	assert.Equal(t, "*mux.Router", reflect.TypeOf(r).String())
}

func testHealthCheckLogging(t *testing.T, endpoint string) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Any()).Times(0)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(wa.healthcheck)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"ok"}`, rr.Body.String())
}

func TestHealthCheckLogging(t *testing.T) {
	// This is served by multiple endpoints, so we'll iterate over all of them.
	endpoints := []string{"/", "/health", "/healthcheck"}
	for _, e := range endpoints {
		testHealthCheckLogging(t, e)
	}
}

// Github webhook tests
func TestGithubWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode github webhook: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/github", payload)
	rr := httptest.NewRecorder()
	wa.githubWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}
func TestGithubWebhookHandlerNoRef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Eq("Possibly a non-Push notification received (blank ref)")).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/v1/webhooks/github", payload)
	rr := httptest.NewRecorder()
	wa.githubWebhookHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Legacy Bitbucket ("Stash") webhook tests
func TestStashWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Reading stash payload body"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode stash webhook: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/stash", payload)
	rr := httptest.NewRecorder()
	wa.stashWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestStashWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Reading stash payload body"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode stash webhook: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "stash", "refChanges": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/stash", payload)
	rr := httptest.NewRecorder()
	wa.stashWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

// Bitbucket Server webhook tests
func TestBitbucketWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Unable to determine bitbucket event type: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestBitbucketWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Info(gomock.Eq("Processing bitbucket-server webhook")).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode bitbucket-server webhook: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "repo:refs_changed", "changes": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

// Bitbucket Cloud webhook tests
func TestBitbucketCloudWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Unable to determine bitbucket event type: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestBitbucketCloudWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Info(gomock.Eq("Processing bitbucket-cloud webhook")).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode bitbucket-cloud webhook: %s"), gomock.Any()).Times(1)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "repo:push", "push": {"changes": "not an array"}}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestUnknownEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)

	wa := NewWebAPI(nil, nil, nil, nil, logger, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "", "changes": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 500, rr.Code)
}
