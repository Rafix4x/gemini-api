package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Rafix4x/gemini-api/internal/gemini"
)

type localImageClient struct {
	base     string
	provider string
	model    string
	key      string
	client   *http.Client
}

func newLocalImageClient(base, provider, model, key string, c *http.Client) *localImageClient {
	return &localImageClient{
		base:     strings.TrimRight(base, "/"),
		provider: provider,
		model:    model,
		key:      key,
		client:   c,
	}
}

type localImageRequest struct {
	Provider    string   `json:"provider"`
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Quality     string   `json:"quality,omitempty"`
	AspectRatio string   `json:"aspect_ratio,omitempty"`
	Resolution  string   `json:"resolution,omitempty"`
	Images      []string `json:"images,omitempty"`
}

type localImageResponse struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"imageUrl"`
	ImageID  string `json:"imageId"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Error    string `json:"error"`
	Code     string `json:"code"`
}

type imageBytes struct {
	Mime string
	Data []byte
}

func (c *localImageClient) Generate(prompt string, refs []string, aspect, resolution, quality string) (*imageBytes, error) {
	payload := localImageRequest{
		Provider:    c.provider,
		Model:       c.model,
		Prompt:      prompt,
		Quality:     quality,
		AspectRatio: aspect,
		Resolution:  resolution,
		Images:      refs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.base+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("local image api: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out localImageResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("local image api: bad response (%d): %s", resp.StatusCode, truncate(string(raw), 160))
	}
	if !out.Success {
		msg := out.Error
		if msg == "" {
			msg = "unknown error"
		}
		if out.Code != "" {
			msg = out.Code + ": " + msg
		}
		return nil, fmt.Errorf("local image api: %s", msg)
	}
	if out.ImageURL == "" {
		return nil, fmt.Errorf("local image api: no image url in the response")
	}

	imgReq, err := http.NewRequest("GET", out.ImageURL, nil)
	if err != nil {
		return nil, err
	}
	if c.key != "" {
		imgReq.Header.Set("Authorization", "Bearer "+c.key)
	}
	imgResp, err := c.client.Do(imgReq)
	if err != nil {
		return nil, fmt.Errorf("local image api: image download: %w", err)
	}
	defer imgResp.Body.Close()

	data, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return nil, err
	}
	if imgResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local image api: image download failed (%d)", imgResp.StatusCode)
	}

	mime := imgResp.Header.Get("Content-Type")
	if mime == "" || !strings.HasPrefix(mime, "image/") {
		mime = gemini.SniffImageMime(data)
	}
	if mime == "" {
		mime = "image/png"
	}
	return &imageBytes{Mime: mime, Data: data}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func aspectFromSize(size string) string {
	size = strings.ToLower(strings.TrimSpace(size))
	if size == "" {
		return "1:1"
	}
	parts := strings.SplitN(size, "x", 2)
	if len(parts) != 2 {
		return "1:1"
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return "1:1"
	}
	r := float64(w) / float64(h)
	switch {
	case r > 0.9 && r < 1.1:
		return "1:1"
	case r >= 1.4 && r <= 1.6:
		return "3:2"
	case r > 1.6:
		return "16:9"
	case r > 1.1:
		return "4:3"
	case r < 0.45:
		return "9:16"
	case r <= 0.6:
		return "9:16"
	case r <= 0.72:
		return "2:3"
	default:
		return "3:4"
	}
}

func resolutionFromSize(size string) string {
	size = strings.ToLower(strings.TrimSpace(size))
	parts := strings.SplitN(size, "x", 2)
	if len(parts) != 2 {
		return "1k"
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return "1k"
	}
	longest := w
	if h > longest {
		longest = h
	}
	switch {
	case longest >= 3800:
		return "4k"
	case longest >= 2048:
		return "2k"
	default:
		return "1k"
	}
}
