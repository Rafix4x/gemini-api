package gemini

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func testPNG() []byte {
	data := make([]byte, 256)
	copy(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	for i := 8; i < len(data); i++ {
		data[i] = byte(i % 251)
	}
	return data
}

func buildWrbLine(inner string) string {
	return fmt.Sprintf(`[["wrb.fr",null,%q]]`, inner)
}

func TestExtractImagesFromResponseFindsInlineImage(t *testing.T) {
	png := testPNG()
	b64 := base64.StdEncoding.EncodeToString(png)
	inner := fmt.Sprintf(`[null,["c_123","r_456"],[["image/png","%s"]]]`, b64)
	raw := ")]}'\n\n177\n" + buildWrbLine(inner)

	images := ExtractImagesFromResponse(raw)
	if len(images) != 1 {
		t.Fatalf("got %d images, want 1", len(images))
	}
	if images[0].Mime != "image/png" {
		t.Errorf("mime = %q, want image/png", images[0].Mime)
	}
	if len(images[0].Data) != len(png) {
		t.Errorf("data length = %d, want %d", len(images[0].Data), len(png))
	}
}

func TestExtractImagesFromResponseFindsDataURL(t *testing.T) {
	png := testPNG()
	b64 := base64.StdEncoding.EncodeToString(png)
	inner := fmt.Sprintf(`[null,[["data:image/png;base64,%s"]]]`, b64)
	raw := ")]}'\n\n177\n" + buildWrbLine(inner)

	images := ExtractImagesFromResponse(raw)
	if len(images) != 1 {
		t.Fatalf("got %d images, want 1", len(images))
	}
	if images[0].Mime != "image/png" {
		t.Errorf("mime = %q, want image/png", images[0].Mime)
	}
}

func TestExtractImagesFromResponseIgnoresText(t *testing.T) {
	inner := `[null,["c_1","r_2"],[["here is a normal answer about images"]]]`
	raw := ")]}'\n\n177\n" + buildWrbLine(inner)

	if images := ExtractImagesFromResponse(raw); len(images) != 0 {
		t.Fatalf("got %d images from a text response, want 0", len(images))
	}
}

func TestExtractImagesFromResponseDeduplicates(t *testing.T) {
	png := testPNG()
	b64 := base64.StdEncoding.EncodeToString(png)
	inner := fmt.Sprintf(`[null,[["%s"]],[["%s"]]]`, b64, b64)
	raw := ")]}'\n\n177\n" + buildWrbLine(inner)

	if images := ExtractImagesFromResponse(raw); len(images) != 1 {
		t.Fatalf("got %d images, want 1 (duplicates removed)", len(images))
	}
}

func TestSniffImageMime(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, "image/png"},
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"webp", []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "image/webp"},
		{"gif", []byte("GIF89a"), "image/gif"},
		{"text", []byte("hello"), ""},
	}
	for _, c := range cases {
		if got := sniffImageMime(c.data); got != c.want {
			t.Errorf("%s: sniffImageMime = %q, want %q", c.name, got, c.want)
		}
	}
}
