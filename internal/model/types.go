package model

import (
	"errors"
	"strings"
)

type Options struct {
	OnlyMainContent  *bool    `json:"only_main_content,omitempty"`
	IncludeSelectors []string `json:"include_selectors,omitempty"`
	ExcludeSelectors []string `json:"exclude_selectors,omitempty"`
}

type Request struct {
	URL         string  `json:"url,omitempty"`
	FinalURL    string  `json:"final_url,omitempty"`
	ContentType string  `json:"content_type,omitempty"`
	HTML        string  `json:"html"`
	Options     Options `json:"options,omitempty"`
}

type Metadata struct {
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	CanonicalURL  string            `json:"canonical_url,omitempty"`
	Language      string            `json:"language,omitempty"`
	Author        string            `json:"author,omitempty"`
	OGTitle       string            `json:"og_title,omitempty"`
	OGDescription string            `json:"og_description,omitempty"`
	Extras        map[string]string `json:"extras,omitempty"`
}

type Quality struct {
	TextLength      int     `json:"text_length"`
	ParagraphCount  int     `json:"paragraph_count"`
	LinkCount       int     `json:"link_count"`
	ImageCount      int     `json:"image_count"`
	TitlePresent    bool    `json:"title_present"`
	LinkDensity     float64 `json:"link_density"`
	QualityScore    float64 `json:"quality_score"`
	UsedMainContent bool    `json:"used_main_content"`
	FallbackUsed    bool    `json:"fallback_used"`
}

type Response struct {
	Markdown  string   `json:"markdown"`
	HTMLClean string   `json:"html_clean"`
	Metadata  Metadata `json:"metadata"`
	Links     []string `json:"links"`
	Images    []string `json:"images"`
	Quality   Quality  `json:"quality"`
}

func (r *Request) ApplyDefaults() {
	r.Options.ApplyDefaults()

	r.URL = strings.TrimSpace(r.URL)
	r.FinalURL = strings.TrimSpace(r.FinalURL)
	r.ContentType = strings.TrimSpace(r.ContentType)
}

func (o *Options) ApplyDefaults() {
	if o.OnlyMainContent == nil {
		o.OnlyMainContent = Bool(true)
	}
}

func (o Options) UseOnlyMainContent() bool {
	if o.OnlyMainContent == nil {
		return true
	}
	return *o.OnlyMainContent
}

func (r Request) Validate() error {
	if strings.TrimSpace(r.HTML) == "" {
		return errors.New("html is required")
	}

	return nil
}

func (r Request) EffectiveURL() string {
	if r.FinalURL != "" {
		return r.FinalURL
	}

	return r.URL
}

func Bool(value bool) *bool {
	return &value
}
