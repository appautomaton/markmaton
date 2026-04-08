package images

import "testing"

func TestExtractDeduplicatesImages(t *testing.T) {
	html := `<img src="https://example.com/a.jpg"><img src="https://example.com/a.jpg">`
	images, err := Extract(html)
	if err != nil {
		t.Fatalf("extract images: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected one unique image, got %d", len(images))
	}
}
