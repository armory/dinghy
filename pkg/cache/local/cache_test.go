package local

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCache(t *testing.T) {
	var c Cache
	c.Add("1", "one")
	c.Add("2", "two")
	one := c.Get("1")
	two := c.Get("2")
	assert.Equal(t, "one", one)
	assert.Equal(t, "two", two)
}
