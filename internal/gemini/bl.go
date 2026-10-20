package gemini

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

var reBL = regexp.MustCompile(`boq_(?:assistant-bard-web-server|gemini-web-uiserver)_\d+\.\d+_p\d+`)

func FetchLatestBL(client Requester) (string, error) {
	return fetchLatestBLFrom(client, "https://gemini.google.com")
}

func fetchLatestBLFrom(client Requester, base string) (string, error) {
	if base == "" {
		base = "https://gemini.google.com"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", base+"/app", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("BL fetch request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("BL fetch read failed: %w", err)
	}

	m := reBL.FindString(string(bodyBytes))
	if m == "" {
		return "", fmt.Errorf("BL pattern not found in response")
	}
	return m, nil
}
