package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlahBlah(t *testing.T) {
	c := NewRedisCacheStore()
	c.Clear()

	c.SetDeps("df1", "mod1", "mod2")
	c.SetDeps("mod1", "mod3", "mod4")

	upstreams, roots := c.UpstreamURLs("mod4")
	assert.EqualValuesf(t, []string{"mod1", "df1"}, upstreams, "mod4 should have parents mod1, df1")
	assert.EqualValuesf(t, []string{"df1"}, roots, "mod4 should have roots df1")

	c.SetDeps("mod1", "mod3")

	_, roots = c.UpstreamURLs("mod4")
	assert.EqualValuesf(t, []string{}, roots, "mod4 should have no roots")
}
