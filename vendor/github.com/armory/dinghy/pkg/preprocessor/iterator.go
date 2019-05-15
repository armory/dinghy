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

package preprocessor

type iterator struct {
	length int
	text   string
	pos    int
}

func newIterator(s string) *iterator {
	return &iterator{
		text:   s,
		length: len(s),
		pos:    0,
	}
}

func (it *iterator) get() rune {
	if it.end() {
		return 0
	}
	return rune(it.text[it.pos])
}

func (it *iterator) end() bool {
	return it.pos >= it.length
}

func (it *iterator) slice(begin int) string {
	return it.text[begin:it.pos]
}
