package preprocessor

import (
	"strconv"
	"strings"
	"unicode"
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
