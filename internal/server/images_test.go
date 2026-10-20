package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rafix4x/gemini-api/internal/config"
)

func testPNG() []byte {
	data := make([]byte, 256)
	copy(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	for i := 8; i < len(data); i++ {
		data[i] = byte(i % 251)
	}
	return data
}

func fakeUpstream(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b64 := base64.StdEncoding.EncodeToString(testPNG())
		inner := fmt.Sprintf(`[null,["c_1","r_2"],[["image/png","%s"]]]`, b64)
		line := fmt.Sprintf(`[["wrb.fr",null,%q]]`, inner)
		w.Header().Set("Content-Type", "application/json+protobuf")
		_, _ = w.Write([]byte(")]}'\n\n177\n" + line + "\n"))
	}))
}

func newTestApp(t *testing.T, upstream string) *App {
	t.Helper()
	cfg := config.Default()
	cfg.UpstreamBase = upstream
	cfg.GeminiBL = "boq_test_p0"
	cfg.RetryAttempts = 1
	cfg.RequestTimeoutSec = 20
	return New(cfg, "test")
}

func TestImageGenerationsEndToEnd(t *testing.T) {
	upstream := fakeUpstream(t)
	defer upstream.Close()

	app := newTestApp(t, upstream.URL)
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()

	body, _ := json.Marshal(map[string]any{
		"model":  "gemini-image",
		"prompt": "a red bicycle leaning on a white wall",
		"n":      1,
	})
	resp, err := http.Post(ts.URL+"/v1/images/generations", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL           string `json:"url"`
			B64JSON       string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("bad json: %v (%s)", err, string(raw))
	}
	if len(out.Data) != 1 {
		t.Fatalf("got %d images, want 1", len(out.Data))
	}
	if out.Data[0].URL == "" {
		t.Fatal("image url is empty")
	}

	imgResp, err := http.Get(out.Data[0].URL)
	if err != nil {
		t.Fatalf("image fetch failed: %v", err)
	}
	defer imgResp.Body.Close()
	imgBytes, _ := io.ReadAll(imgResp.Body)
	if imgResp.StatusCode != http.StatusOK {
		t.Fatalf("image status = %d", imgResp.StatusCode)
	}
	if !bytes.HasPrefix(imgBytes, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("image is not a png: %x", imgBytes[:4])
	}
	if len(imgBytes) != 256 {
		t.Fatalf("image length = %d, want 256", len(imgBytes))
	}
}

func TestImageGenerationsB64(t *testing.T) {
	upstream := fakeUpstream(t)
	defer upstream.Close()

	app := newTestApp(t, upstream.URL)
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()

	body, _ := json.Marshal(map[string]any{
		"prompt":          "a red bicycle",
		"response_format": "b64_json",
	})
	resp, err := http.Post(ts.URL+"/v1/images/generations", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var out struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("bad json: %v (%s)", err, string(raw))
	}
	if len(out.Data) != 1 || out.Data[0].B64JSON == "" {
		t.Fatalf("expected b64_json, got %s", string(raw))
	}
	decoded, err := base64.StdEncoding.DecodeString(out.Data[0].B64JSON)
	if err != nil || !bytes.HasPrefix(decoded, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("b64_json is not a png: %v", err)
	}
}

func TestImageGenerationsRejectsEmptyPrompt(t *testing.T) {
	upstream := fakeUpstream(t)
	defer upstream.Close()

	app := newTestApp(t, upstream.URL)
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/images/generations", "application/json", strings.NewReader(`{"prompt":""}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestImageGenerationsRejectsUnknownModel(t *testing.T) {
	upstream := fakeUpstream(t)
	defer upstream.Close()

	app := newTestApp(t, upstream.URL)
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/v1/images/generations", "application/json", strings.NewReader(`{"model":"nope","prompt":"a cat"}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func fakeLocalAPI(t *testing.T, seen *map[string]any) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var self *httptest.Server
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if seen != nil {
			*seen = body
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  true,
			"imageId":  "img_1",
			"imageUrl": self.URL + "/img.png",
		})
	})
	mux.HandleFunc("/img.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(testPNG())
	})
	ts := httptest.NewServer(mux)
	self = ts
	t.Cleanup(ts.Close)
	return ts
}

func TestImageGenerationsLocalBackend(t *testing.T) {
	var seen map[string]any
	upstream := fakeLocalAPI(t, &seen)

	cfg := config.Default()
	cfg.UpstreamBase = "http://127.0.0.1:1"
	cfg.ImageBackend = "local"
	cfg.LocalImageURL = upstream.URL
	cfg.LocalImageProvider = "alpha"
	cfg.LocalImageModel = "flare"
	cfg.RetryAttempts = 1
	app := New(cfg, "test")
	ts := httptest.NewServer(app.Handler())
	defer ts.Close()

	body, _ := json.Marshal(map[string]any{
		"model":  "gemini-image",
		"prompt": "a red bicycle",
		"size":   "1024x1024",
	})
	resp, err := http.Post(ts.URL+"/v1/images/generations", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, string(raw))
	}
	var out struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(out.Data) != 1 || out.Data[0].URL == "" {
		t.Fatalf("no image url: %s", string(raw))
	}
	if seen["provider"] != "alpha" || seen["model"] != "flare" {
		t.Fatalf("local api got %v", seen)
	}
	if seen["aspect_ratio"] != "1:1" {
		t.Fatalf("aspect = %v, want 1:1", seen["aspect_ratio"])
	}
	imgResp, err := http.Get(out.Data[0].URL)
	if err != nil {
		t.Fatalf("image fetch failed: %v", err)
	}
	defer imgResp.Body.Close()
	imgBytes, _ := io.ReadAll(imgResp.Body)
	if !bytes.HasPrefix(imgBytes, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("not a png: %x", imgBytes[:4])
	}
}

func TestAspectFromSize(t *testing.T) {
	cases := map[string]string{
		"1024x1024": "1:1",
		"1920x1080": "16:9",
		"1080x1920": "9:16",
		"800x600":   "4:3",
		"1200x800":  "3:2",
		"":          "1:1",
		"junk":      "1:1",
	}
	for in, want := range cases {
		if got := aspectFromSize(in); got != want {
			t.Errorf("aspectFromSize(%q) = %q, want %q", in, got, want)
		}
	}
}
