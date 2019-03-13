package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const DINGHY_ENV_TEST = "DINGHY_ENV_TEST"

func TestGetenvOrDefault(t *testing.T) {
	err := os.Setenv(DINGHY_ENV_TEST, "test")
	if err != nil {
		t.Error(err)
	}

	test := GetenvOrDefault(DINGHY_ENV_TEST, "foo")
	assert.Equal(t, "test", test)

	notFound := GetenvOrDefault("DOES_NOT_EXIST", "baz")
	assert.Equal(t, "baz", notFound)
}