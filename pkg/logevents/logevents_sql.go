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
	"github.com/armory/dinghy/pkg/database"
	"strings"
	"time"
)

type LogEventSQLClient struct {
	MinutesTTL time.Duration
	SQLClient  *database.SQLClient
}

func (LogEventSQL) TableName() string {
	return "logevents"
}

type LogEventSQL struct {
	Id                 int    `gorm:"primaryKey;column:id"`
	Org                string `gorm:"column:org"`
	Repo               string `gorm:"column:repo"`
	Files              string `gorm:"column:files"`
	Message            string `gorm:"column:message"`
	Date               int64  `gorm:"column:commitdate"`
	Commits            string `gorm:"column:commits"`
	Status             string `gorm:"column:status"`
	RawData            string `gorm:"column:rawdata"`
	Author             string `gorm:"column:author"`
	RenderedDinghyfile string `gorm:"column:rendereddinghyfile"`
}

func (log LogEventSQL) ToLogEvent() LogEvent {
	return LogEvent{
		Org:                log.Org,
		Repo:               log.Repo,
		Files:              strings.Split(log.Files, ","),
		Message:            log.Message,
		Date:               log.Date,
		Commits:            strings.Split(log.Commits, ","),
		Status:             log.Status,
		RawData:            log.RawData,
		RenderedDinghyfile: log.RenderedDinghyfile,
	}
}

func (log LogEvent) ToLogEventSQL() LogEventSQL {
	return LogEventSQL{
		Org:                log.Org,
		Repo:               log.Repo,
		Files:              strings.Join(log.Files, ","),
		Message:            log.Message,
		Date:               log.Date,
		Commits:            strings.Join(log.Commits, ","),
		Status:             log.Status,
		RawData:            log.RawData,
		RenderedDinghyfile: log.RenderedDinghyfile,
	}
}

func (c LogEventSQLClient) GetLogEvents() ([]LogEvent, error) {
	queryLogEvents := []LogEventSQL{}
	result := []LogEvent{}
	now := time.Now().UnixNano()
	reducednanos := now - int64(c.MinutesTTL*time.Minute)
	milis := reducednanos / 1000000

	queryResult := c.SQLClient.Client.Where("commitdate >= ?", int(milis)).Find(&queryLogEvents)
	if queryResult.Error != nil {
		return nil, queryResult.Error
	}

	for _, val := range queryLogEvents {
		result = append(result, val.ToLogEvent())
	}
	return result, nil
}

func (c LogEventSQLClient) SaveLogEvent(logEvent LogEvent) error {
	convert := logEvent.ToLogEventSQL()
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	convert.Date = milis
	return c.SQLClient.Client.Create(&convert).Error
}
