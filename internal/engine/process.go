package engine

import (
	"strings"

	"github.com/appautomaton/markmaton/internal/cleanhtml"
	"github.com/appautomaton/markmaton/internal/convert"
	"github.com/appautomaton/markmaton/internal/images"
	"github.com/appautomaton/markmaton/internal/links"
	"github.com/appautomaton/markmaton/internal/metadata"
	"github.com/appautomaton/markmaton/internal/model"
	"github.com/appautomaton/markmaton/internal/postprocess"
	"github.com/appautomaton/markmaton/internal/quality"
	"github.com/appautomaton/markmaton/internal/resolve"
)

func Process(request model.Request) (model.Response, error) {
	request.ApplyDefaults()

	meta, err := metadata.Extract(request.HTML)
	if err != nil {
		return model.Response{}, err
	}

	response, err := runPipeline(request, meta, request.Options.OnlyMainContent, false)
	if err != nil {
		return model.Response{}, err
	}

	if request.Options.OnlyMainContent && quality.NeedsFallback(response.Quality) {
		fallback, err := runPipeline(request, meta, false, true)
		if err != nil {
			return model.Response{}, err
		}
		response = fallback
	}

	return response, nil
}

func runPipeline(request model.Request, meta model.Metadata, onlyMainContent bool, fallbackUsed bool) (model.Response, error) {
	cleaned, err := cleanhtml.Clean(
		request.HTML,
		onlyMainContent,
		request.Options.IncludeSelectors,
		request.Options.ExcludeSelectors,
	)
	if err != nil {
		return model.Response{}, err
	}

	resolved, err := resolve.ResolveHTML(cleaned, request.EffectiveURL())
	if err != nil {
		return model.Response{}, err
	}

	markdown, err := convert.ToMarkdown(resolved)
	if err != nil {
		return model.Response{}, err
	}
	markdown = postprocess.Markdown(markdown)

	extractedLinks, err := links.Extract(resolved)
	if err != nil {
		return model.Response{}, err
	}

	extractedImages, err := images.Extract(resolved)
	if err != nil {
		return model.Response{}, err
	}

	meta = normalizeMetadata(meta, extractedLinks)
	qualityResult := quality.Analyze(markdown, meta.Title, len(extractedLinks), len(extractedImages), onlyMainContent, fallbackUsed)

	return model.Response{
		Markdown:  markdown,
		HTMLClean: strings.TrimSpace(resolved),
		Metadata:  meta,
		Links:     extractedLinks,
		Images:    extractedImages,
		Quality:   qualityResult,
	}, nil
}

func normalizeMetadata(meta model.Metadata, links []string) model.Metadata {
	if meta.CanonicalURL == "" && len(links) > 0 {
		meta.CanonicalURL = links[0]
	}
	return meta
}
