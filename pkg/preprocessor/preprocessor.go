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
					replacementText := stringifyJSON(innerText, '[', ']')
					replacementText = stringifyJSON(replacementText, '{', '}')
					outputText := rawText[0:start] + replacementText + Preprocess(rawText[end:])
					return outputText
				}
			}
		}
	}
	return rawText
}

func stringifyJSON(text string, lparen, rparen byte) string {
	stack := ""
	var start, end int
	for i := 0; i < len(text); i++ {
		if text[i] == lparen {
			if len(stack) == 0 {
				start = i
			}
			stack += string(lparen) // push
		} else if text[i] == rparen && len(stack) > 0 {
			stack = stack[:len(stack)-1] // pop
			if len(stack) == 0 {
				end = i
				stringified := `"` + strings.Replace(text[start:end+1], `"`, `\"`, -1) + `"`
				return text[:start] + stringified + stringifyJSON(text[end+1:], lparen, rparen)
			}
		}
	}
	return text
}
