package logevents

import (
	"github.com/armory/dinghy/pkg/database"
	"strings"
	"time"
)

type LogEventSQLClient struct {
	MinutesTTL	time.Duration
	SQLClient 	*database.SQLClient
}


func (LogEventSQL) TableName() string {
	return "logevents"
}

type LogEventSQL struct {
	Id 			int		`gorm:"primaryKey;column:id"`
	Org 		string	`gorm:"column:org"`
	Repo 		string	`gorm:"column:repo"`
	Files		string	`gorm:"column:files"`
	Message		string	`gorm:"column:message"`
	Date		int64	`gorm:"column:commitdate"`
	Commits		string	`gorm:"column:commits"`
	Status		string	`gorm:"column:status"`
	RawData		string	`gorm:"column:rawdata"`
	RenderedDinghyfile	string	`gorm:"column:rendereddinghyfile"`
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
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000

	queryResult := c.SQLClient.Client.Where("date >= ?", milis - int64(c.MinutesTTL * time.Minute) ).Find(&queryLogEvents)
	if queryResult.Error != nil{
		return nil, queryResult.Error
	}

	for _, val := range queryLogEvents {
		result = append(result, val.ToLogEvent())
	}
	return result, nil
}

func (c LogEventSQLClient) SaveLogEvent(logEvent LogEvent) error {
	convert := logEvent.ToLogEventSQL()
	return c.SQLClient.Client.Create(&convert).Error
}