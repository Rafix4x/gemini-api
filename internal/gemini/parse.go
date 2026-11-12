package gemini

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	reCodeExec    = regexp.MustCompile(`(?s)\x60\x60\x60(?:python|javascript|text)\?code_(?:reference|stdout)&code_event_index=\d+\n.*?\x60\x60\x60\n?`)
	reCardContent = regexp.MustCompile(`http://googleusercontent\.com/card_content/\d+\n?`)
	reBardError   = regexp.MustCompile(`BardErrorInfo\s*\[(\d+)\]`)
)

func CleanText(text string, strip bool) string {
	text = reCodeExec.ReplaceAllString(text, "")
	text = reCardContent.ReplaceAllString(text, "")
	if strip {
		return strings.TrimSpace(text)
	}
	return text
}

func ExtractTextsFromLine(line string) []string {
	if !strings.Contains(line, `"wrb.fr"`) || len(line) < 200 {
		return nil
	}

	var arr []any
	if err := json.Unmarshal([]byte(line), &arr); err != nil || len(arr) == 0 {
		return nil
	}

	firstElem, ok := arr[0].([]any)
	if !ok || len(firstElem) < 3 {
		return nil
	}

	innerStr, ok := firstElem[2].(string)
	if !ok || len(innerStr) < 50 {
		return nil
	}

	var inner []any
	if err := json.Unmarshal([]byte(innerStr), &inner); err != nil {
		return nil
	}

	if len(inner) <= 4 || inner[4] == nil {
		return nil
	}

	parts, ok := inner[4].([]any)
	if !ok {
		return nil
	}

	var texts []string
	for _, part := range parts {
		partList, ok := part.([]any)
		if ok && len(partList) > 1 && partList[1] != nil {
			tList, ok := partList[1].([]any)
			if ok {
				for _, t := range tList {
					if str, ok := t.(string); ok && str != "" {
						texts = append(texts, str)
					}
				}
			}
		}
	}

	return texts
}

func IsBardError(raw string) (string, bool) {
	match := reBardError.FindStringSubmatch(raw)
	if len(match) > 1 {
		return match[1], true
	}
	return "", false
}

func ExtractResponseText(raw string) (string, error) {
	if code, ok := IsBardError(raw); ok {
		return "", fmt.Errorf("Gemini upstream rejected request: BardErrorInfo [%s]", code)
	}

	var lastText string
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		texts := ExtractTextsFromLine(line)
		for _, t := range texts {
			if len(t) > len(lastText) {
				lastText = t
			}
		}
	}

	return CleanText(lastText, true), nil
}

// ExtractResponseModel reads the upstream UI model label at payload index 42.
// It does not inspect generated text or infer identity from a requested selector.
func ExtractResponseModel(raw string) string {
	var label string
	for _, line := range strings.Split(raw, "\n") {
		var frames [][]json.RawMessage
		if json.Unmarshal([]byte(line), &frames) != nil {
			continue
		}
		for _, frame := range frames {
			if len(frame) < 3 {
				continue
			}
			var kind, innerString string
			if json.Unmarshal(frame[0], &kind) != nil || kind != "wrb.fr" || json.Unmarshal(frame[2], &innerString) != nil {
				continue
			}
			var inner []json.RawMessage
			if json.Unmarshal([]byte(innerString), &inner) != nil || len(inner) <= 42 {
				continue
			}
			var current string
			if json.Unmarshal(inner[42], &current) == nil && strings.TrimSpace(current) != "" {
				label = current
			}
		}
	}
	return label
}
