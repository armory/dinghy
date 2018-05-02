package dinghyfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mystruct struct{}

func TestInvalidJSON(t *testing.T) {
	var d mystruct

	noQuote := `{
		"key": noQuote"
	}`
	err := Unmarshal([]byte(noQuote), &d)
	assert.Error(t, err, "Missing quote JSON didn't generate correct erromessage")
	assert.Contains(t, err.Error(), `Error in line 2, char 10`)

	noComma := `{
		"foo": "bar"
		"baz": "foo"
	}`
	err = Unmarshal([]byte(noComma), &d)
	assert.Error(t, err, "Missing comma JSON didn't generate correct erromessage")
	assert.Contains(t, err.Error(), `Error in line 3, char 2: invalid character '"' after object key:value pair`)
}
