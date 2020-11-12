package log

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
)

type DinghyLog interface {

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	GetBytesBuffByLoggerKey(key string) (*bytes.Buffer, error)
}

const (
	SystemLogKey = "system"
	LogEventKey = "logevent"
)

func (d DinghyLogs) GetBytesBuffByLoggerKey(key string) (*bytes.Buffer, error) {
	if val, ok := d.Logs[key]; ok {
		return val.LogEventBuffer, nil
	}
	return nil, errors.New("Log does not exists")
}

func (d DinghyLogs) Debugf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Debugf(format, args)
	}
}

func (d DinghyLogs) Infof(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Infof(format, args)
	}
}

func (d DinghyLogs) Printf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Printf(format, args)
	}
}

func (d DinghyLogs) Warnf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Warnf(format, args)
	}
}

func (d DinghyLogs) Warningf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Warningf(format, args)
	}
}

func (d DinghyLogs) Errorf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Errorf(format, args)
	}
}

func (d DinghyLogs) Fatalf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Fatalf(format, args)
	}
}

func (d DinghyLogs) Panicf(format string, args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Panicf(format, args)
	}
}

func (d DinghyLogs) Debug(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Debug(args)
	}
}

func (d DinghyLogs) Info(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Info(args)
	}
}

func (d DinghyLogs) Print(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Print(args)
	}
}

func (d DinghyLogs) Warn(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Warn(args)
	}
}

func (d DinghyLogs) Warning(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Warning(args)
	}
}

func (d DinghyLogs) Error(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Error(args)
	}
}

func (d DinghyLogs) Fatal(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Fatal(args)
	}
}

func (d DinghyLogs) Panic(args ...interface{}) {
	for _, log := range d.Logs {
		log.Logger.Panic(args)
	}
}

type DinghyLogs struct {
	Logs	map[string]DinghyLogStruct
}

type DinghyLogStruct struct {
	Logger	 		log.FieldLogger
	LogEventBuffer	*bytes.Buffer
}

func NewDinghyLogs(systemLog log.FieldLogger) DinghyLog {
	logevent := log.New()
	var memLog bytes.Buffer
	lvl, _ := log.ParseLevel("info")
	logevent.SetLevel(lvl)
	logevent.SetOutput(&memLog)
	return DinghyLogs{Logs: map[string]DinghyLogStruct{
		SystemLogKey : {
			Logger:         systemLog,
			LogEventBuffer: &bytes.Buffer{},
		},
		LogEventKey : {
			Logger:			logevent,
			LogEventBuffer:	&memLog,
		},
	}}
}
