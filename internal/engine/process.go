package engine

import (
	"context"
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
	return ProcessContext(context.Background(), request)
}

func ProcessContext(ctx context.Context, request model.Request) (model.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}

	request.ApplyDefaults()
	if err := request.Validate(); err != nil {
		return model.Response{}, err
	}
	onlyMainContent := request.Options.UseOnlyMainContent()

	meta, err := metadata.Extract(request.HTML)
	if err != nil {
		return model.Response{}, err
	}

	response, err := runPipeline(ctx, request, meta, onlyMainContent, false)
	if err != nil {
		return model.Response{}, err
	}

	if onlyMainContent && quality.NeedsFallback(response.Quality) {
		fallback, err := runPipeline(ctx, request, meta, false, true)
		if err != nil {
			return model.Response{}, err
		}
		response = fallback
	}

	return response, nil
}

func runPipeline(ctx context.Context, request model.Request, meta model.Metadata, onlyMainContent bool, fallbackUsed bool) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}

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

	markdown, err := convert.ToMarkdownContext(ctx, resolved)
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

	meta = normalizeMetadata(request, meta)
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

func normalizeMetadata(request model.Request, meta model.Metadata) model.Metadata {
	if meta.CanonicalURL == "" {
		meta.CanonicalURL = request.EffectiveURL()
	}
	return meta
}
