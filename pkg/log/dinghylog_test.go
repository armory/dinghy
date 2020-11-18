package log

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewDinghyLogs(t *testing.T) {
	logger := log.New()
	dinghylog := NewDinghyLogs(logger)
	val, err := dinghylog.GetBytesBuffByLoggerKey(SystemLogKey)
	if err != nil {
		t.Errorf("System log does not exists in dinghylog")
	}
	if "" != val.String() {
		t.Errorf("System log buffer should be empty")
	}
}