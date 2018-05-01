package preprocessor

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	log "github.com/sirupsen/logrus"
)

func parseWhitespace(it *iterator) string {
	begin := it.pos
	for !it.end() && unicode.IsSpace(it.get()) {
		it.pos++
	}
	return it.slice(begin)
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
	for !it.end() && !unicode.IsSpace(it.get()) {
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
func Preprocess(text string) string {
	length := len(text)

	for i := 0; i < length-1 && length >= 2; i++ {
		if text[i:i+2] != "{{" {
			continue
		}

		it := newIterator(text)
		it.pos = i + 2
		parts := []string{"{{"}

		for !it.end() {
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

		return text[:i] + strings.Join(parts, "") + Preprocess(text[it.pos:])
	}

	return text
}

// ParseGlobalVars returns the map of global variables in the dinghyfile
func ParseGlobalVars(input string) interface{} {

	d := make(map[string]interface{})
	input = removeModules(input)
	err := json.Unmarshal([]byte(input), &d)
	if err != nil {
		log.Error(err)
		return nil
	}
	if val, ok := d["globals"]; ok {
		return val
	}
	return make(map[string]interface{})
}

func dummySubstitute(args ...interface{}) string {
	return `{ "a": "b" }`
}

// since {{ var ... }} can be a string or an int!
func dummyVar(args ...interface{}) string {
	return "1"
}

// removeModules replaces all template function calls ({{ ... }}) in the dinghyfile with
// the JSON: { "a": "b" } so that we can extract the global vars using JSON.Unmarshal
func removeModules(input string) string {

	funcMap := template.FuncMap{
		"module": dummySubstitute,
		"var": dummyVar,
		"pipelineID": dummySubstitute,
	}

	tmpl, err := template.New("blank-out").Funcs(funcMap).Parse(input)
	if err != nil {
		log.Errorf("Blank out modules parse error: %s", err)
		return input
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		log.Errorf("Blank out modules execute error: %s", err)
		return input
	}

	return buf.String()
}
