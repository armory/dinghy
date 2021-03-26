package remote

import (
	"github.com/armory/dinghy/pkg/settings/source"
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
	s, err := i.LoadSetupSettings()

	assert.Nil(t, err)
	assert.Equal(t, true, s.SQL.Enabled)
	assert.Equal(t, false, s.SQL.EventLogsOnly)
}
