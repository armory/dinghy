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

package web

import (
	"fmt"
	dinghylog "github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/logevents"
)

func saveLogEventError(logeventClient logevents.LogEventsClient, p Push, dinghyLog dinghylog.DinghyLog, logEvent logevents.LogEvent) {
	currentLogEvent := appendPushToLogEvent(logEvent, p)
	saveLogEvent(logeventClient, p, dinghyLog, currentLogEvent, "error")
}

func saveLogEventSuccess(logeventClient logevents.LogEventsClient, p Push, dinghyLog dinghylog.DinghyLog, logEvent logevents.LogEvent) {
	currentLogEvent := appendPushToLogEvent(logEvent, p)
	saveLogEvent(logeventClient, p, dinghyLog, currentLogEvent, "success")
}

func saveLogEvent(logeventClient logevents.LogEventsClient, p Push, dinghyLog dinghylog.DinghyLog, logEvent logevents.LogEvent, status string) {
	if buf, err := dinghyLog.GetBytesBuffByLoggerKey(dinghylog.LogEventKey); err == nil {
		logEvent.Message = fmt.Sprintf("%v", buf)
		logEvent.Status = status
		logeventClient.SaveLogEvent(logEvent)
	}
}

func appendPushToLogEvent(logEvent logevents.LogEvent, push Push) logevents.LogEvent {
	if logEvent.Org == "" {
		logEvent.Org = push.Org()
	}
	if logEvent.Repo == "" {
		logEvent.Repo = push.Repo()
	}
	if logEvent.Files == nil {
		logEvent.Files = push.Files()
	}
	if logEvent.Commits == nil {
		logEvent.Commits = push.GetCommits()
	}
	return logEvent
}
