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

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Masterminds/sprig/v3"
	"github.com/armory/dinghy/pkg/git"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

func parseWhitespace(it *iterator) string {
	for !it.end() && unicode.IsSpace(it.get()) {
		it.pos++
	}
	return " "
}

func parseString(it *iterator) string {
	begin := it.pos
	it.pos++
	for !it.end() && it.get() != '"' {
		if it.get() == '\\' {
			it.pos++
		}
		it.pos++
	}
	it.pos++
	return it.slice(begin)
}

func parseToken(it *iterator) string {
	begin := it.pos
	var prevstr string
	for !it.end() && !unicode.IsSpace(it.get()) {
		if prevstr+string(it.get()) == "}}" {
			it.pos--
			break
		}
		prevstr = string(it.get())
		it.pos++
	}
	return it.slice(begin)
}

func parseJSONObject(it *iterator) string {
	begin := it.pos
	stack := []rune{it.get()}
	it.pos++

	for !it.end() && len(stack) > 0 {
		switch it.get() {
		case '"':
			parseString(it)

		case '[', '{':
			stack = append(stack, it.get())
			it.pos++

		case ']':
			if stack[len(stack)-1] == '[' {
				stack = stack[:len(stack)-1]
			}
			it.pos++

		case '}':
			if stack[len(stack)-1] == '{' {
				stack = stack[:len(stack)-1]
			}
			it.pos++

		default:
			it.pos++
		}
	}
	return strconv.Quote(it.slice(begin))
}

func isElvisOperator(it *iterator) bool {
	if it.pos+2 < it.length {
		if it.text[it.pos:it.pos+2] == "?:" {
			if unicode.IsSpace(rune(it.text[it.pos+2])) {
				return true
			}
		}
	}
	return false
}

func parseElvisOperator(it *iterator) string {
	it.pos += 2
	for !it.end() && unicode.IsSpace(it.get()) {
		it.pos++
	}

	// ignore the elvis operator -- it's just for improved readability
	return ""
}

// Preprocess makes a first pass at the dinghyfile and stringifies the JSON args to a module
func Preprocess(text string) (string, error) {
	length := len(text)

	for i := 0; i < length-1 && length >= 2; i++ {
		if text[i:i+2] != "{{" {
			continue
		}

		it := newIterator(text)
		it.pos = i + 2
		parts := []string{"{{"}

		for !it.end() {
			if it.pos+2 > length {
				return text, errors.New("Index out of bounds while pre-processing template action, possibly a missing '}}'")
			}

			if it.text[it.pos:it.pos+2] == "}}" {
				parts = append(parts, "}}")
				it.pos += 2
				break
			}

			ch := it.get()
			var part string

			if unicode.IsSpace(ch) {
				part = parseWhitespace(it)
			} else if ch == '"' {
				part = parseString(it)
			} else if ch == '{' || ch == '[' {
				part = parseJSONObject(it)
			} else if isElvisOperator(it) {
				part = parseElvisOperator(it)
			} else {
				part = parseToken(it)
			}

			parts = append(parts, part)
		}

		remaining, err := Preprocess(text[it.pos:])
		if err != nil {
			return text, err
		}
		return text[:i] + strings.Join(parts, "") + remaining, nil
	}

	return text, nil
}

// ParseGlobalVars returns the map of global variables in the dinghyfile
func ParseGlobalVars(input string, gitInfo git.GitInfo) (interface{}, error) {

	d := make(map[string]interface{})
	input = removeModules(input, gitInfo)
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		return nil, err
	}
	if val, ok := d["globals"]; ok {
		return val, nil
	}
	return make(map[string]interface{}), nil
}

func dummySubstitute(args ...interface{}) string {
	return `{ "a": "b" }`
}

func dummyKV(args ...interface{}) string {
	return `"a": "b"`
}

// since {{ var ... }} can be a string or an int!
func dummyVar(args ...interface{}) string {
	return "1"
}

func dummySlice(args ...interface{}) []string {
	return make([]string, 0)
}

// removeModules replaces all template function calls ({{ ... }}) in the dinghyfile with
// the JSON: { "a": "b" } so that we can extract the global vars using JSON.Unmarshal
func removeModules(input string, gitInfo git.GitInfo) string {

	funcMap := template.FuncMap{
		"module":       dummySubstitute,
		"local_module": dummySubstitute,
		"appModule":    dummyKV,
		"var":          dummyVar,
		"pipelineID":   dummyVar,
		"makeSlice":    dummySlice,
		"if":           dummySlice,
	}

	// All sprig functions will be changed for a dummy slice
	for key,_ := range sprig.GenericFuncMap() {
		funcMap[key] = dummySlice
	}

	tmpl, err := template.New("blank-out").Funcs(funcMap).Parse(input)
	if err != nil {
		return input
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, gitInfo)
	if err != nil {
		return input
	}

	return buf.String()
}
