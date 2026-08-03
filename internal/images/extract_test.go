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

func TestExtractDropsUnsafeImageSources(t *testing.T) {
	html := `<img src="javascript:alert(1)"><img src="data:image/svg+xml,bad"><img src="https://example.com/safe.jpg">`
	images, err := Extract(html)
	if err != nil {
		t.Fatalf("extract images: %v", err)
	}
	if len(images) != 1 || images[0] != "https://example.com/safe.jpg" {
		t.Fatalf("unexpected safe images: %v", images)
	}
}

func TestExtractIncludesPictureSourcesButNotAudioOrVideoSources(t *testing.T) {
	html := `<picture><source src="https://example.com/cover.webp"><img src="https://example.com/cover.jpg"></picture>
	<video><source src="https://example.com/demo.mp4"></video>
	<audio><source src="https://example.com/episode.mp3"></audio>`
	images, err := Extract(html)
	if err != nil {
		t.Fatalf("extract images: %v", err)
	}
	if len(images) != 2 || images[0] != "https://example.com/cover.webp" || images[1] != "https://example.com/cover.jpg" {
		t.Fatalf("unexpected images: %v", images)
	}
}
