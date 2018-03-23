package preprocessor

import (
	"strings"
)

func preprocess(rawText string) string {
	for i := 0; i < len(rawText)-1 && len(rawText) >= 2; i++ {
		if rawText[i:i+2] == "{{" {
			start := i + 2
			end := start
			for j := start; j < len(rawText)-1; j++ {
				if rawText[j:j+2] == "}}" {
					end = j
					i = j + 2
					innerText := rawText[start:end]
					replacementText := stringifyArgs(innerText)
					outputText := rawText[0:start] + replacementText + preprocess(rawText[end:])
					return outputText
				}
			}
		}
	}
	return rawText
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
