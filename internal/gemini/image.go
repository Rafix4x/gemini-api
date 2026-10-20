package gemini

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
)

type ImageData struct {
	Mime string
	Data []byte
}

var (
	reDataURL = regexp.MustCompile(`^data:(image/[a-z0-9.+-]+);base64,`)
	reLongB64 = regexp.MustCompile(`^[A-Za-z0-9+/=\r\n]{64,}$`)
)

func SniffImageMime(b []byte) string {
	return sniffImageMime(b)
}

func sniffImageMime(b []byte) string {
	switch {
	case len(b) >= 8 && b[0] == 0x89 && b[1] == 'P' && b[2] == 'N' && b[3] == 'G':
		return "image/png"
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return "image/jpeg"
	case len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP":
		return "image/webp"
	case len(b) >= 3 && string(b[0:3]) == "GIF":
		return "image/gif"
	}
	return ""
}

func decodeImageCandidate(s string) (ImageData, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ImageData{}, false
	}

	mime := ""
	if m := reDataURL.FindStringSubmatch(s); m != nil {
		mime = m[1]
		s = s[len(m[0]):]
	} else if !reLongB64.MatchString(s) {
		return ImageData{}, false
	}

	clean := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, s)

	raw, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(clean)
		if err != nil {
			return ImageData{}, false
		}
	}

	if sniffed := sniffImageMime(raw); sniffed != "" {
		if mime == "" {
			mime = sniffed
		}
		return ImageData{Mime: mime, Data: raw}, true
	}

	return ImageData{}, false
}

func walkStrings(v any, fn func(string)) {
	switch t := v.(type) {
	case string:
		fn(t)
	case []any:
		for _, item := range t {
			walkStrings(item, fn)
		}
	case map[string]any:
		for _, item := range t {
			walkStrings(item, fn)
		}
	}
}

func ExtractImagesFromResponse(raw string) []ImageData {
	var out []ImageData
	seen := map[string]bool{}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if !strings.Contains(line, `"wrb.fr"`) {
			continue
		}

		var arr []any
		if err := json.Unmarshal([]byte(line), &arr); err != nil || len(arr) == 0 {
			continue
		}

		first, ok := arr[0].([]any)
		if !ok || len(first) < 3 {
			continue
		}
		innerStr, ok := first[2].(string)
		if !ok || len(innerStr) < 20 {
			continue
		}

		var inner any
		if err := json.Unmarshal([]byte(innerStr), &inner); err != nil {
			continue
		}

		walkStrings(inner, func(s string) {
			img, ok := decodeImageCandidate(s)
			if !ok {
				return
			}
			key := img.Mime + ":" + string(img.Data[:min(64, len(img.Data))])
			if seen[key] {
				return
			}
			seen[key] = true
			out = append(out, img)
		})
	}

	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
