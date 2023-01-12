package test

import (
	"github.com/armory/dinghy/cmd"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestThatGetAddressWorks(t *testing.T) {
	var options = dinghy.NewRedisOptions(global.Redis{
		BaseURL: "rediss://something",
		Password: "bob",
	})
	assert.Equal(t, "something", options.Addr)
	assert.NotNilf(t, options.TLSConfig, "TLS config should not be nil!")

	options = dinghy.NewRedisOptions(global.Redis{
		BaseURL: "redis://no-ssl-redis",
		Password: "bob",
	})
	assert.Equal(t, "no-ssl-redis", options.Addr)
	assert.Nilf(t, options.TLSConfig, "TLS config should be nil for redis:// requests!")
}
