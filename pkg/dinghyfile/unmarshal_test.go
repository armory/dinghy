/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mystruct struct{}

func TestInvalidJSON(t *testing.T) {
	var d mystruct

	dmu := &DinghyJsonUnmarshaller{}

	noQuote := `{
		"key": noQuote"
	}`
	err := dmu.Unmarshal([]byte(noQuote), &d)
	assert.Error(t, err, "Missing quote JSON didn't generate correct erromessage")
	assert.Contains(t, err.Error(), `Error in line 2, char 10`)

	noComma := `{
		"foo": "bar"
		"baz": "foo"
	}`
	err = dmu.Unmarshal([]byte(noComma), &d)
	assert.Error(t, err, "Missing comma JSON didn't generate correct erromessage")
	assert.Contains(t, err.Error(), `Error in line 3, char 2: invalid character '"' after object key:value pair`)
}
