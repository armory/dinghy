package util

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type mock struct {
	S string `json:"s"`
}

func TestReadJSON(t *testing.T) {
	answer := &mock{}
	json := strings.NewReader(`{ "s" : "test" }`)
	ReadJSON(json, &answer)
	assert.NotNil(t, answer)
	assert.Equal(t, "test", answer.S)
}