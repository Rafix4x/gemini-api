package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Rafix4x/gemini-api/internal/format"
)

type storedImage struct {
	Data      []byte
	Mime      string
	CreatedAt time.Time
}

type imageStore struct {
	mu    sync.Mutex
	items map[string]storedImage
	ttl   time.Duration
	max   int
}

func newImageStore(ttlMinutes, max int) *imageStore {
	if ttlMinutes <= 0 {
		ttlMinutes = 30
	}
	if max <= 0 {
		max = 100
	}
	return &imageStore{items: map[string]storedImage{}, ttl: time.Duration(ttlMinutes) * time.Minute, max: max}
}

func (s *imageStore) put(data []byte, mime string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, item := range s.items {
		if now.Sub(item.CreatedAt) > s.ttl {
			delete(s.items, id)
		}
	}
	for len(s.items) >= s.max {
		oldestID := ""
		var oldest time.Time
		for id, item := range s.items {
			if oldestID == "" || item.CreatedAt.Before(oldest) {
				oldestID = id
				oldest = item.CreatedAt
			}
		}
		delete(s.items, oldestID)
	}

	buf := make([]byte, 10)
	_, _ = rand.Read(buf)
	id := "img_" + hex.EncodeToString(buf)
	s.items[id] = storedImage{Data: data, Mime: mime, CreatedAt: now}
	return id
}

func (s *imageStore) get(id string) (storedImage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return storedImage{}, false
	}
	if time.Since(item.CreatedAt) > s.ttl {
		delete(s.items, id)
		return storedImage{}, false
	}
	return item, true
}

type imageGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format"`
	Quality        string `json:"quality"`
	Image          string `json:"image"`
	ImageURL       string `json:"image_url"`
}

func (a *App) handleImageGenerations(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "invalid JSON",
				"type":    "invalid_request_error",
				"code":    "invalid_json",
			},
		})
		return
	}

	var req imageGenerationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "invalid_json",
			},
		})
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if len(prompt) < 3 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "prompt is required (min 3 characters)",
				"type":    "invalid_request_error",
				"code":    "invalid_prompt",
			},
		})
		return
	}

	if req.Model != "" && req.Model != a.Cfg.ImageModel {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": fmt.Sprintf("unknown image model %q (use %q)", req.Model, a.Cfg.ImageModel),
				"type":    "invalid_request_error",
				"code":    "model_not_found",
			},
		})
		return
	}

	n := req.N
	if n <= 0 {
		n = 1
	}
	if n > 4 {
		n = 4
	}

	var refs []string
	if src := req.Image; src != "" {
		refs = []string{src}
	} else if src := req.ImageURL; src != "" {
		refs = []string{src}
	}
	if len(refs) > 0 {
		ref := format.Image{URL: refs[0]}
		if data, mime, ok := decodeDataURL(refs[0]); ok {
			ref = format.Image{Data: data, MIME: mime}
		}
		uploaded, err := a.uploadImages([]format.Image{ref})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"message": "reference image could not be read: " + err.Error(),
					"type":    "invalid_request_error",
					"code":    "invalid_image",
				},
			})
			return
		}
		refs = uploaded
	}

	images, err := a.generateImage(prompt, refs, req.Size, req.Quality)
	if err != nil {
		a.Logf("Image generation failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": map[string]any{
				"message": err.Error(),
				"type":    "upstream_error",
				"code":    "image_generation_failed",
			},
		})
		return
	}

	created := time.Now().Unix()
	data := make([]map[string]any, 0, len(images))
	for _, img := range images {
		entry := map[string]any{}
		if req.ResponseFormat == "b64_json" {
			entry["b64_json"] = base64.StdEncoding.EncodeToString(img.Data)
		} else {
			id := a.Images.put(img.Data, img.Mime)
			entry["url"] = a.publicURL(r, "/v1/images/"+id+imageExtension(img.Mime))
		}
		entry["revised_prompt"] = prompt
		data = append(data, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"created": created,
		"data":    data,
	})
}

// generateImage dispatches to the configured backend.
func (a *App) generateImage(prompt string, refs []string, size, quality string) ([]imageBytes, error) {
	if a.Cfg.ImageBackend == "local" {
		cli := newLocalImageClient(a.Cfg.LocalImageURL, a.Cfg.LocalImageProvider, a.Cfg.LocalImageModel, a.Cfg.LocalImageKey, a.HTTPClient)
		img, err := cli.Generate(prompt, refs, aspectFromSize(size), resolutionFromSize(size), quality)
		if err != nil {
			return nil, err
		}
		return []imageBytes{*img}, nil
	}

	raw, err := a.Gem.GenerateImage(prompt, a.Cfg.ImageModelID, 0, refs, nil)
	if err != nil {
		return nil, err
	}
	out := make([]imageBytes, 0, len(raw))
	for _, img := range raw {
		out = append(out, imageBytes{Mime: img.Mime, Data: img.Data})
	}
	return out, nil
}

func (a *App) handleImageFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if idx := strings.LastIndex(id, "."); idx > 0 {
		id = id[:idx]
	}
	item, ok := a.Images.get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"message": "image not found or expired",
				"type":    "invalid_request_error",
				"code":    "image_not_found",
			},
		})
		return
	}
	w.Header().Set("Content-Type", item.Mime)
	w.Header().Set("Cache-Control", "public, max-age=1800")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(item.Data)
}

func imageExtension(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func (a *App) publicURL(r *http.Request, path string) string {
	scheme := "https"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.Split(proto, ",")[0]
	} else if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

func decodeDataURL(src string) ([]byte, string, bool) {
	src = strings.TrimSpace(src)
	if !strings.HasPrefix(src, "data:") {
		return nil, "", false
	}
	idx := strings.Index(src, ",")
	if idx < 0 {
		return nil, "", false
	}
	header := src[5:idx]
	payload := src[idx+1:]
	mime := strings.Split(header, ";")[0]
	if !strings.HasPrefix(header, "base64") {
		return nil, "", false
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	if err != nil {
		return nil, "", false
	}
	return data, mime, true
}
