package convert

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type htmlPolicy func(root *goquery.Selection)

type markdownPolicy func(markdown string) string

type htmlPolicyRegistration struct {
	Name  string
	Apply htmlPolicy
}

type markdownPolicyRegistration struct {
	Name  string
	Apply markdownPolicy
}

func defaultHTMLPolicyRegistrations() []htmlPolicyRegistration {
	return []htmlPolicyRegistration{
		{Name: "drop_button_like_elements", Apply: removeButtonLikeElements},
	}
}

func defaultMarkdownPolicyRegistrations() []markdownPolicyRegistration {
	return []markdownPolicyRegistration{
		{Name: "trim_opening_shell_controls", Apply: trimOpeningShellControls()},
		{Name: "drop_standalone_control_lines", Apply: dropStandaloneControlLines()},
		{Name: "drop_standalone_metric_lines", Apply: dropStandaloneMetricLines()},
		{Name: "drop_redundant_opening_heading_echoes", Apply: dropRedundantOpeningHeadingEchoes()},
		{Name: "collapse_adjacent_duplicate_lines", Apply: collapseAdjacentDuplicateLines()},
	}
}

func applyDefaultHTMLPolicies(ctx context.Context, source string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	document, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("parse HTML for conversion policies: %w", err)
	}
	for _, registration := range defaultHTMLPolicyRegistrations() {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if registration.Apply == nil {
			return "", fmt.Errorf("HTML policy %q is nil", registration.Name)
		}
		registration.Apply(document.Selection)
	}

	if body := document.Find("body").First(); body.Length() > 0 {
		processed, err := body.Html()
		if err != nil {
			return "", fmt.Errorf("render HTML after conversion policies: %w", err)
		}
		return processed, nil
	}
	processed, err := document.Html()
	if err != nil {
		return "", fmt.Errorf("render HTML after conversion policies: %w", err)
	}
	return processed, nil
}

func applyDefaultMarkdownPolicies(ctx context.Context, markdown string) (string, error) {
	for _, registration := range defaultMarkdownPolicyRegistrations() {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if registration.Apply == nil {
			return "", fmt.Errorf("Markdown policy %q is nil", registration.Name)
		}
		markdown = registration.Apply(markdown)
	}
	return markdown, nil
}
