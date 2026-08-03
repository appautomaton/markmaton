package convert

import (
	"context"
	"errors"
	"strings"

	"github.com/appautomaton/markmaton/internal/convert/native"
)

var defaultConverter = native.New()

// ToMarkdown converts HTML with the Native converter.
func ToMarkdown(html string) (string, error) {
	return ToMarkdownContext(context.Background(), html)
}

// ToMarkdownContext converts HTML with the Native converter while honoring
// cancellation through ctx.
func ToMarkdownContext(ctx context.Context, html string) (string, error) {
	if ctx == nil {
		return "", errors.New("conversion context is nil")
	}
	if strings.TrimSpace(html) == "" {
		return "", nil
	}

	processedHTML, err := applyDefaultHTMLPolicies(ctx, html)
	if err != nil {
		return "", err
	}
	markdown, err := defaultConverter.Convert(ctx, processedHTML)
	if err != nil {
		return "", err
	}
	return applyDefaultMarkdownPolicies(ctx, markdown)
}
