/*
* Copyright 2020 Armory, Inc.

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

package logevents

type LogEventsClient interface {
	GetLogEvents() ([]LogEvent, error)
	SaveLogEvent(logEvent LogEvent) error
}

type LogEvent struct {
	Org                string   `json:"org" yaml:"org"`
	Repo               string   `json:"repo" yaml:"repo"`
	Files              []string `json:"files" yaml:"files"`
	Message            string   `json:"message" yaml:"message"`
	Date               int64    `json:"date" yaml:"date"`
	Commits            []string `json:"commits" yaml:"commits"`
	Status             string   `json:"status" yaml:"status"`
	RawData            string   `json:"rawdata" yaml:"rawdata"`
	RenderedDinghyfile string   `json:"rendereddinghyfile" yaml:"rendereddinghyfile"`
}
