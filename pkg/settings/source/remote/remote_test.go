package remote

import (
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRemoteSource(t *testing.T) {
	i := NewRemoteSource()

	var s source.Source = i
	assert.NotNil(t, i)
	assert.NotNil(t, s)
}
