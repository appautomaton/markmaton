package quality

import "testing"

func TestAnalyzeQuality(t *testing.T) {
	result := Analyze(
		"# Title\n\nFirst paragraph.\n\nSecond paragraph.",
		"Title",
		1,
		1,
		true,
		false,
	)

	if result.TextLength == 0 {
		t.Fatalf("expected text length to be positive")
	}
	if !result.TitlePresent {
		t.Fatalf("expected title to be present")
	}
	if result.QualityScore <= 0 {
		t.Fatalf("expected quality score to be positive")
	}
}

func TestNeedsFallback(t *testing.T) {
	weak := Analyze("tiny", "", 0, 0, true, false)
	if !NeedsFallback(weak) {
		t.Fatalf("expected weak content to require fallback")
	}
}
