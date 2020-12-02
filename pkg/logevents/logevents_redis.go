package logevents

import (
	"encoding/json"
	"github.com/armory/dinghy/pkg/cache"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type LogEventRedisClient struct {
	MinutesTTL	time.Duration
	RedisClient *cache.RedisCache
}

func (c LogEventRedisClient) GetLogEvents() ([]LogEvent, error) {
	loge := log.WithFields(log.Fields{"func": "GetLogEvents"})
	key := cache.CompileKey("logEvent","*")
	var cursor uint64
	result := []LogEvent{}
	for {
		keys, nextcursor, err := c.RedisClient.Client.Scan(cursor, key , 1000).Result()
		cursor = nextcursor
		if err != nil {
			loge.WithFields(log.Fields{"operation": "scan key", "key": cache.CompileKey("logEvent*")}).Error(err)
			return nil, err
		}
		for _, key := range keys {
			currentEventLog, errorNoKey := c.RedisClient.Client.Get(key).Result()
			if errorNoKey != nil {
				loge.WithFields(log.Fields{"operation": "get key", "key": key}).Error(err)
				continue
			}
			var logEvent LogEvent
			errorUnmarshal := json.Unmarshal([]byte(currentEventLog), &logEvent)
			if errorUnmarshal != nil {
				loge.WithFields(log.Fields{"operation": "unmarshall key " + key, "content": currentEventLog}).Error(err)
				continue
			}
			result = append(result, logEvent)
		}

		if cursor == 0 {
			break
		}
	}

	return result, nil
}

func (c LogEventRedisClient) SaveLogEvent(logEvent LogEvent) error {
	filesSet := make(map[string]bool)
	for _, value := range logEvent.Files {
		filesSet[value] = true
	}
	// Deck does not like null values
	files := []string{}
	for key, _ := range filesSet {
		files = append(files, key)
	}
	// Deck does not like null values
	if logEvent.Commits == nil {
		logEvent.Commits = []string{}
	}
	logEvent.Files = files
	loge := log.WithFields(log.Fields{"func": "SaveLogEvent"})
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	logEvent.Date = milis
	logEventBytes, err := json.Marshal(logEvent)
	if err != nil {
		loge.WithFields(log.Fields{"operation": "marshall logEvent", "content": logEvent}).Error(err)
		return err
	}
	key := cache.CompileKey("logEvent", strconv.FormatInt(milis, 10))
	if _, err := c.RedisClient.Client.Set(key, logEventBytes, c.MinutesTTL * time.Minute).Result(); err != nil {
		loge.WithFields(log.Fields{"operation": "set key", "key": key, "content": logEventBytes}).Error(err)
		return err
	}
	return nil
}
