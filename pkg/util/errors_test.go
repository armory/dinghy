package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGitHubFileNotFoundErr(t *testing.T) {
	assert.True(t, IsGitHubFileNotFoundErr("No file named stuff found in other stuff"))
	assert.False(t, IsGitHubFileNotFoundErr("meh"))
}
