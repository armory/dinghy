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

import (
	"encoding/json"
	"github.com/armory/dinghy/pkg/cache"
	"time"
)

type LogEventsClient interface {
	GetLogEvents() ([]LogEvent, error)
	SaveLogEvent(logEvent LogEvent) error
}

type LogEventRedisClient struct {
	RedisClient *cache.RedisCache
}

type LogEvent struct {
	Org 	string
	Repo 	string
	File	string
	User	string
	Message	string
	Date	int64
	Commithash	string
}

func (c LogEventRedisClient) GetLogEvents() ([]LogEvent, error) {
	keys, _, err := c.RedisClient.Client.Scan(0, cache.CompileKey("logEventsKeys"), 1000).Result()
	if err != nil {
		return nil, err
	}

	var result []LogEvent
	for _, key := range keys {
		var logEvent LogEvent
		errorUnmarshal := json.Unmarshal([]byte(key), &logEvent)
		if errorUnmarshal != nil {

			continue
		}
		result = append(result, logEvent)
	}

	return result, nil
}

func (c LogEventRedisClient) SaveLogEvent(logEvent LogEvent) error {
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	logEvent.Date = milis
	logEventBytes, err := json.Marshal(logEvent)
	if err != nil {
		return err
	}
	key := cache.CompileKey("logEvent" + string(milis))
	c.RedisClient.Client.Set(key, logEventBytes, 1 * time.Hour)

	return nil
}