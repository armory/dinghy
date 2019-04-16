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
