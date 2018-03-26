package preprocessor

import (
	"strings"
)

// Preprocess makes a first pass at the dinghyfile and stringifies the JSON args to a module
func Preprocess(rawText string) string {
	for i := 0; i < len(rawText)-1 && len(rawText) >= 2; i++ {
		if rawText[i:i+2] == "{{" {
			start := i + 2
			end := start
			for j := start; j < len(rawText)-1; j++ {
				if rawText[j:j+2] == "}}" {
					end = j
					i = j + 2
					innerText := rawText[start:end]
					replacementText := quoteArray(innerText)
					replacementText = stringifyArgs(replacementText)
					outputText := rawText[0:start] + replacementText + Preprocess(rawText[end:])
					return outputText
				}
			}
		}
	}
	return rawText
}

// this is a hack, there are complexities to escaping [] inside a json object
// for our purposes since a stringified [] is still an []
func quoteArray(args string) string {
	args = strings.Replace(args, `[`, `"[`, 1)
	args = strings.Replace(args, `]`, `]"`, 1)
	return args
}

func stringifyArgs(args string) string {
	stack := ""
	var start, end int
	for i := 0; i < len(args); i++ {
		if args[i] == '{' {
			if len(stack) == 0 {
				start = i
			}
			stack += "{" // push
		} else if args[i] == '}' && len(stack) > 0 {
			stack = stack[:len(stack)-1] // pop
			if len(stack) == 0 {
				end = i
				stringified := `"` + strings.Replace(args[start:end+1], `"`, `\"`, -1) + `"`
				return args[:start] + stringified + stringifyArgs(args[end+1:])
			}
		}
	}
	return args
}
