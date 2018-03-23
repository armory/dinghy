package preprocessor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPreprocess(t *testing.T) {
	input := `{ "a": {{ module "myMod" "key" {"my": "value"} }} }`
	expected := `{ "a": {{ module "myMod" "key" "{\"my\": \"value\"}" }} }`
	actual := preprocess(input)
	t.Log(actual)
	assert.Equal(t, expected, actual)
}
