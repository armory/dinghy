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

package dinghyfile

import (
	"bytes"
	"fmt"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"github.com/armory/dinghy/pkg/log"
	"github.com/sirupsen/logrus"
	"strings"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/mock"
	"github.com/golang/mock/gomock"
)

// mock out events so that it gets passed over and doesn't do anything
type EventsTestClient struct{}

type partialMatcher struct {
	y string
}

func (p partialMatcher) Matches(x interface{}) bool {
	return strings.Contains(x.(string), p.y)
}
func (p partialMatcher) String() string {
	return fmt.Sprintf("contains %v", p.y)
}
func containsString(x string) gomock.Matcher { return partialMatcher{x} }

func (c *EventsTestClient) SendEvent(eventType string, event *events.Event) {}

// This returns a test PipelineBuilder object.
func testBasePipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		Depman:      cache.NewMemoryCache(),
		EventClient: &EventsTestClient{},
		Logger:      NewDinghylog(),
		Ums:         []Unmarshaller{&DinghyJsonUnmarshaller{}},
		TemplateOrg: "armory",
		Action:      pipebuilder.Process,
	}
}

// This sets a mock logger on the pipeline {
func mockLogger(dr *DinghyfileParser, ctrl *gomock.Controller) *mock.MockDinghyLog {
	dr.Builder.Logger = mock.NewMockDinghyLog(ctrl)
	return dr.Builder.Logger.(*mock.MockDinghyLog)
}

func NewDinghylog() log.DinghyLog {
	return log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
		log.SystemLogKey: {
			Logger:         logrus.New(),
			LogEventBuffer: &bytes.Buffer{},
		},
	}}
}
