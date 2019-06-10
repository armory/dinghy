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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/armory/dinghy/pkg/mock"
	// "github.com/armory/dinghy/pkg/settings"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTrue(t *testing.T) {
	assert.Equal(t, true, true)
}

func TestHealthCheckLogging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Any()).Times(0)

	wa := &WebAPI{Logger: logger}
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(wa.healthcheck)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), `{"status":"ok"}`)
}
