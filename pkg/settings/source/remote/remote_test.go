package remote

import (
	"github.com/armory/dinghy/pkg/settings/source"
	logr "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRemoteSource(t *testing.T) {
	i := NewRemoteSource()

	var s source.SourceConfiguration = i
	assert.NotNil(t, i)
	assert.NotNil(t, s)
}

func TestRemoteSourceIsEnabledWhenRemoteIsActive(t *testing.T) {

	i := NewRemoteSource()
	log := logr.New()
	s, err := i.LoadSetupSettings(log)

	assert.Nil(t, err)
	assert.Equal(t, true, s.SQL.Enabled)
	assert.Equal(t, false, s.SQL.EventLogsOnly)
}
