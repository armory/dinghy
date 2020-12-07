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

// Credits to hashicorp:
// https://github.com/hashicorp/packer/blob/master/common/json/unmarshal.go

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Unmarshaller interface {
	Unmarshal(data []byte, i interface{}) error
}

type DinghyJsonUnmarshaller struct{}

// Unmarshal is wrapper around json.Unmarshal that returns user-friendly
// errors when there are syntax errors.
func (d DinghyJsonUnmarshaller) Unmarshal(data []byte, i interface{}) error {
	err := json.Unmarshal(data, i)
	if err != nil {
		syntaxErr, ok := err.(*json.SyntaxError)
		if !ok {
			return err
		}

		// We have a syntax error. Extract out the line number and friends.
		// https://groups.google.com/forum/#!topic/golang-nuts/fizimmXtVfc
		newline := []byte{'\x0a'}

		// Calculate the start/end position of the line where the error is
		start := bytes.LastIndex(data[:syntaxErr.Offset], newline) + 1
		end := len(data)
		if idx := bytes.Index(data[start:], newline); idx >= 0 {
			end = start + idx
		}

		// Count the line number we're on plus the offset in the line
		line := bytes.Count(data[:start], newline) + 1
		pos := int(syntaxErr.Offset) - start - 1

		err = fmt.Errorf("Error in line %d, char %d: %s\n%s",
			line, pos, syntaxErr, data[start:end])
		return err
	}

	return nil
}
