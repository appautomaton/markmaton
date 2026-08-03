package native

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/appautomaton/markmaton/internal/safeurl"
	"golang.org/x/net/html"
)

type parser struct {
	ctx context.Context
}

func parseDocument(ctx context.Context, source string, limits Limits, transforms []DocumentTransform) (document, error) {
	root, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return document{}, fmt.Errorf("parse HTML: %w", err)
	}
	if err := applyDocumentTransforms(ctx, root, transforms); err != nil {
		return document{}, err
	}
	if err := validateDocumentLimits(ctx, root, limits); err != nil {
		return document{}, err
	}

	body := findElement(root, "body")
	if body == nil {
		body = root
	}

	p := &parser{ctx: ctx}
	blocks, err := p.parseBlocks(body, 0)
	if err != nil {
		return document{}, err
	}
	return document{blocks: blocks}, nil
}

func validateDocumentLimits(ctx context.Context, root *html.Node, limits Limits) error {
	type pendingNode struct {
		node  *html.Node
		depth int
	}
	stack := []pendingNode{{node: root, depth: 0}}
	nodeCount := 0
	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		nodeCount++
		if nodeCount > limits.MaxNodes {
			return fmt.Errorf("%w: got at least %d nodes, maximum is %d", ErrNodeLimit, nodeCount, limits.MaxNodes)
		}
		if current.depth > limits.MaxDepth {
			return fmt.Errorf("%w: got depth %d, maximum is %d", ErrDepthLimit, current.depth, limits.MaxDepth)
		}
		for child := current.node.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, pendingNode{node: child, depth: current.depth + 1})
		}
	}
	return nil
}

func (p *parser) visit(_ int) error {
	return p.ctx.Err()
}

func (p *parser) parseBlocks(parent *html.Node, depth int) ([]block, error) {
	if err := p.visit(depth); err != nil {
		return nil, err
	}

	var blocks []block
	var inlineRun []*html.Node

	flushInlineRun := func() error {
		if len(inlineRun) == 0 {
			return nil
		}
		inlines, err := p.parseInlineNodes(inlineRun, depth+1)
		if err != nil {
			return err
		}
		inlines = compactInlines(inlines, true)
		if hasInlineContent(inlines) {
			blocks = append(blocks, block{kind: blockParagraph, inlines: inlines})
		}
		inlineRun = nil
		return nil
	}

	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.CommentNode || isIgnoredElement(child) {
			continue
		}
		if isLinkedBlockContainer(child) {
			if err := flushInlineRun(); err != nil {
				return nil, err
			}
			converted, err := p.parseLinkedBlocks(child, depth+1)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, converted...)
			continue
		}
		if isBlockNode(child) || hasBlockDescendant(child) {
			if err := flushInlineRun(); err != nil {
				return nil, err
			}
			converted, err := p.parseBlock(child, depth+1)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, converted...)
			continue
		}
		inlineRun = append(inlineRun, child)
	}

	if err := flushInlineRun(); err != nil {
		return nil, err
	}
	return blocks, nil
}

func (p *parser) parseBlock(node *html.Node, depth int) ([]block, error) {
	if err := p.visit(depth); err != nil {
		return nil, err
	}
	if node.Type != html.ElementNode {
		inlines, err := p.parseInlineNodes([]*html.Node{node}, depth+1)
		if err != nil {
			return nil, err
		}
		inlines = compactInlines(inlines, true)
		if !hasInlineContent(inlines) {
			return nil, nil
		}
		return []block{{kind: blockParagraph, inlines: inlines}}, nil
	}

	name := strings.ToLower(node.Data)
	switch name {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level, _ := strconv.Atoi(name[1:])
		inlines, err := p.parseInlineChildren(node, depth+1)
		if err != nil {
			return nil, err
		}
		inlines = normalizeHeadingInlines(compactInlines(inlines, true))
		if !hasInlineContent(inlines) {
			return nil, nil
		}
		return []block{{kind: blockHeading, level: level, inlines: inlines}}, nil

	case "p", "figcaption", "summary", "dt", "dd":
		inlines, err := p.parseInlineChildren(node, depth+1)
		if err != nil {
			return nil, err
		}
		inlines = compactInlines(inlines, true)
		if !hasInlineContent(inlines) {
			return nil, nil
		}
		return []block{{kind: blockParagraph, inlines: inlines}}, nil

	case "iframe":
		return inlineBlock(iframeLink(node)), nil

	case "video":
		return inlineBlock(mediaLink(node, "Video")), nil

	case "audio":
		return inlineBlock(mediaLink(node, "Audio")), nil

	case "pre":
		codeNode := findElement(node, "code")
		if codeNode == nil {
			codeNode = node
		}
		code, err := p.collectCodeText(codeNode, depth+1)
		if err != nil {
			return nil, err
		}
		language := codeLanguage(codeNode)
		if language == "" {
			language = codeLanguage(node)
		}
		return []block{{kind: blockCode, code: strings.TrimRight(code, "\n"), language: language}}, nil

	case "blockquote":
		children, err := p.parseBlocks(node, depth+1)
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			return nil, nil
		}
		return []block{{kind: blockQuote, children: children}}, nil

	case "ul", "ol":
		converted, err := p.parseList(node, depth+1)
		if err != nil {
			return nil, err
		}
		if len(converted.items) == 0 {
			return nil, nil
		}
		return []block{converted}, nil

	case "table":
		converted, err := p.parseTable(node, depth+1)
		if err != nil {
			return nil, err
		}
		if len(converted.table.rows) == 0 {
			return p.parseBlocks(node, depth+1)
		}
		return []block{converted}, nil

	case "hr":
		return []block{{kind: blockThematicBreak}}, nil

	default:
		return p.parseBlocks(node, depth+1)
	}
}

func isLinkedBlockContainer(node *html.Node) bool {
	return node != nil && node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") && hasBlockDescendant(node)
}

func hasBlockDescendant(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	stack := make([]*html.Node, 0, 8)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		stack = append(stack, child)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if isBlockNode(current) {
			return true
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			stack = append(stack, child)
		}
	}
	return false
}

func (p *parser) parseLinkedBlocks(node *html.Node, depth int) ([]block, error) {
	blocks, err := p.parseBlocks(node, depth+1)
	if err != nil {
		return nil, err
	}
	href, _ := attribute(node, "href")
	href = strings.TrimSpace(href)
	if href == "" || href == "#" || !safeurl.Link(href) {
		return blocks, nil
	}
	title, _ := attribute(node, "title")
	link := inline{kind: inlineLink, url: href, title: collapseWhitespace(title)}
	for index := range blocks {
		wrapBlockInLink(&blocks[index], link)
	}
	return blocks, nil
}

func inlineBlock(inlines []inline) []block {
	inlines = compactInlines(inlines, true)
	if !hasInlineContent(inlines) {
		return nil
	}
	return []block{{kind: blockParagraph, inlines: inlines}}
}

func wrapBlockInLink(value *block, link inline) {
	switch value.kind {
	case blockParagraph, blockHeading:
		value.inlines = wrapInlineNode(link, value.inlines)
	case blockQuote:
		for index := range value.children {
			wrapBlockInLink(&value.children[index], link)
		}
	case blockList:
		for itemIndex := range value.items {
			for blockIndex := range value.items[itemIndex].blocks {
				wrapBlockInLink(&value.items[itemIndex].blocks[blockIndex], link)
			}
		}
	case blockTable:
		for rowIndex := range value.table.rows {
			for cellIndex := range value.table.rows[rowIndex].cells {
				cell := &value.table.rows[rowIndex].cells[cellIndex]
				cell.inlines = wrapInlineNode(link, cell.inlines)
			}
		}
	}
}

func normalizeHeadingInlines(values []inline) []inline {
	normalized := make([]inline, 0, len(values))
	for _, value := range values {
		if value.kind == inlineHardBreak {
			normalized = append(normalized, inline{kind: inlineText, text: " "})
			continue
		}
		if len(value.children) > 0 {
			value.children = normalizeHeadingInlines(value.children)
		}
		normalized = append(normalized, value)
	}
	return compactInlines(normalized, true)
}

func (p *parser) parseList(node *html.Node, depth int) (block, error) {
	result := block{
		kind:     blockList,
		ordered:  strings.EqualFold(node.Data, "ol"),
		reversed: false,
		start:    1,
	}
	hasExplicitStart := false
	if result.ordered {
		_, result.reversed = attribute(node, "reversed")
		if value, ok := attribute(node, "start"); ok {
			if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed >= 0 {
				result.start = parsed
				hasExplicitStart = true
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if strings.EqualFold(child.Data, "li") {
			checked, hasTask := listItemTaskState(child)
			blocks, err := p.parseBlocks(child, depth+1)
			if err != nil {
				return block{}, err
			}
			if len(blocks) == 0 && !hasTask {
				continue
			}
			item := listItem{blocks: blocks, hasTask: hasTask, checked: checked}
			if result.ordered {
				if value, ok := attribute(child, "value"); ok {
					if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed >= 0 {
						item.hasExplicitValue = true
						item.explicitValue = parsed
					}
				}
			}
			result.items = append(result.items, item)
			continue
		}
		if !strings.EqualFold(child.Data, "ul") && !strings.EqualFold(child.Data, "ol") {
			continue
		}
		nested, err := p.parseList(child, depth+1)
		if err != nil {
			return block{}, err
		}
		if len(nested.items) == 0 {
			continue
		}
		if len(result.items) == 0 {
			result.items = append(result.items, listItem{blocks: []block{nested}})
		} else {
			last := len(result.items) - 1
			result.items[last].blocks = append(result.items[last].blocks, nested)
		}
	}
	if result.ordered && result.reversed && !hasExplicitStart {
		result.start = len(result.items)
	}
	return result, nil
}

func listItemTaskState(item *html.Node) (bool, bool) {
	stack := make([]*html.Node, 0, 8)
	for child := item.LastChild; child != nil; child = child.PrevSibling {
		stack = append(stack, child)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "li") {
			continue
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "input") {
			typeValue, _ := attribute(current, "type")
			if strings.EqualFold(strings.TrimSpace(typeValue), "checkbox") {
				_, checked := attribute(current, "checked")
				return checked, true
			}
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return false, false
}

func (p *parser) parseTable(node *html.Node, depth int) (block, error) {
	result := block{kind: blockTable}
	if caption := firstDescendantBeforeNestedTable(node, "caption"); caption != nil {
		inlines, err := p.parseInlineChildren(caption, depth+1)
		if err != nil {
			return block{}, err
		}
		result.table.caption = compactInlines(inlines, true)
	}

	var walk func(*html.Node, int) error
	walk = func(current *html.Node, currentDepth int) error {
		if err := p.visit(currentDepth); err != nil {
			return err
		}
		if current != node && current.Type == html.ElementNode && strings.EqualFold(current.Data, "table") {
			result.table.complex = true
			return nil
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "tr") {
			row := tableRow{header: hasAncestorBefore(current, "thead", "table")}
			allHeaders := true
			for cellNode := current.FirstChild; cellNode != nil; cellNode = cellNode.NextSibling {
				if cellNode.Type != html.ElementNode || (!strings.EqualFold(cellNode.Data, "th") && !strings.EqualFold(cellNode.Data, "td")) {
					continue
				}
				if !strings.EqualFold(cellNode.Data, "th") {
					allHeaders = false
				}
				if tableSpan(cellNode, "colspan") > 1 || tableSpan(cellNode, "rowspan") > 1 || hasComplexTableCellContent(cellNode) {
					result.table.complex = true
				}
				inlines, err := p.parseInlineChildren(cellNode, currentDepth+1)
				if err != nil {
					return err
				}
				row.cells = append(row.cells, tableCell{
					inlines:   compactInlines(inlines, true),
					alignment: tableCellAlignment(cellNode),
				})
			}
			if len(row.cells) > 0 {
				row.header = row.header || allHeaders
				result.table.rows = append(result.table.rows, row)
			}
			return nil
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child, currentDepth+1); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(node, depth); err != nil {
		return block{}, err
	}
	return result, nil
}

func firstDescendantBeforeNestedTable(root *html.Node, name string) *html.Node {
	stack := make([]*html.Node, 0, 8)
	for child := root.LastChild; child != nil; child = child.PrevSibling {
		stack = append(stack, child)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, name) {
			return current
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "table") {
			continue
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return nil
}

func tableSpan(node *html.Node, name string) int {
	value, ok := attribute(node, name)
	if !ok {
		return 1
	}
	span, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || span < 1 {
		return 1
	}
	return span
}

func tableCellAlignment(node *html.Node) tableAlignment {
	alignment, _ := attribute(node, "align")
	alignment = strings.ToLower(strings.TrimSpace(alignment))
	if alignment == "" {
		style, _ := attribute(node, "style")
		style = strings.ToLower(strings.ReplaceAll(style, " ", ""))
		for _, candidate := range []string{"left", "center", "right"} {
			if strings.Contains(style, "text-align:"+candidate) {
				alignment = candidate
				break
			}
		}
	}
	switch alignment {
	case "left":
		return tableAlignLeft
	case "center":
		return tableAlignCenter
	case "right":
		return tableAlignRight
	default:
		return tableAlignDefault
	}
}

func hasComplexTableCellContent(cell *html.Node) bool {
	stack := make([]*html.Node, 0, 8)
	for child := cell.FirstChild; child != nil; child = child.NextSibling {
		stack = append(stack, child)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.ElementNode {
			switch strings.ToLower(current.Data) {
			case "blockquote", "div", "ol", "p", "pre", "table", "ul":
				return true
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			stack = append(stack, child)
		}
	}
	return false
}

func (p *parser) parseInlineChildren(parent *html.Node, depth int) ([]inline, error) {
	var nodes []*html.Node
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		nodes = append(nodes, child)
	}
	return p.parseInlineNodes(nodes, depth)
}

func (p *parser) parseInlineNodes(nodes []*html.Node, depth int) ([]inline, error) {
	var result []inline
	for _, node := range nodes {
		converted, err := p.parseInline(node, depth)
		if err != nil {
			return nil, err
		}
		result = append(result, converted...)
	}
	return compactInlines(result, false), nil
}

func (p *parser) parseInline(node *html.Node, depth int) ([]inline, error) {
	if err := p.visit(depth); err != nil {
		return nil, err
	}

	switch node.Type {
	case html.TextNode:
		text := normalizeHTMLText(node.Data)
		if text == "" {
			return nil, nil
		}
		return []inline{{kind: inlineText, text: text}}, nil
	case html.CommentNode:
		return nil, nil
	case html.ElementNode:
		if isIgnoredElement(node) {
			return nil, nil
		}
	default:
		return nil, nil
	}

	name := strings.ToLower(node.Data)
	switch name {
	case "br":
		return []inline{{kind: inlineHardBreak}}, nil
	case "img":
		src, _ := attribute(node, "src")
		src = strings.TrimSpace(src)
		alt, _ := attribute(node, "alt")
		alt = collapseWhitespace(alt)
		if src == "" {
			return nil, nil
		}
		if !safeurl.Resource(src) {
			if alt == "" {
				return nil, nil
			}
			return []inline{{kind: inlineText, text: alt}}, nil
		}
		title, _ := attribute(node, "title")
		return []inline{{kind: inlineImage, text: alt, url: src, title: collapseWhitespace(title)}}, nil
	case "iframe":
		return iframeLink(node), nil
	case "video":
		return mediaLink(node, "Video"), nil
	case "audio":
		return mediaLink(node, "Audio"), nil
	case "input":
		return nil, nil
	case "code", "kbd", "samp", "tt":
		code, err := p.collectCodeText(node, depth+1)
		if err != nil {
			return nil, err
		}
		return []inline{{kind: inlineCode, text: code}}, nil
	}

	children, err := p.parseInlineChildren(node, depth+1)
	if err != nil {
		return nil, err
	}
	children = compactInlines(children, false)

	switch name {
	case "strong", "b":
		return wrapInlineNode(inline{kind: inlineStrong}, children), nil
	case "em", "i":
		return wrapInlineNode(inline{kind: inlineEmphasis}, children), nil
	case "del", "s", "strike":
		return wrapInlineNode(inline{kind: inlineStrikethrough}, children), nil
	case "a":
		href, _ := attribute(node, "href")
		href = strings.TrimSpace(href)
		if !hasInlineContent(children) {
			label, _ := attribute(node, "aria-label")
			if label == "" {
				label, _ = attribute(node, "title")
			}
			if label != "" {
				children = []inline{{kind: inlineText, text: collapseWhitespace(label)}}
			}
		}
		if href == "" || href == "#" || !safeurl.Link(href) || !hasInlineContent(children) {
			return children, nil
		}
		title, _ := attribute(node, "title")
		return wrapInlineNode(inline{kind: inlineLink, url: href, title: collapseWhitespace(title)}, children), nil
	default:
		return children, nil
	}
}

func wrapInlineNode(wrapper inline, children []inline) []inline {
	children = flattenInlineKind(children, wrapper.kind)
	if wrapper.kind != inlineEmphasis && wrapper.kind != inlineStrong && wrapper.kind != inlineStrikethrough {
		return wrapSingleInlineNode(wrapper, children)
	}

	segments := splitInlinesAtHardBreaks(children)
	result := make([]inline, 0, len(segments))
	for _, segment := range segments {
		if len(segment) == 1 && segment[0].kind == inlineHardBreak {
			result = append(result, segment[0])
			continue
		}
		result = append(result, wrapSingleInlineNode(wrapper, segment)...)
	}
	return compactInlines(result, false)
}

func wrapSingleInlineNode(wrapper inline, children []inline) []inline {
	children, leading := trimLeadingInlineWhitespace(children)
	children, trailing := trimTrailingInlineWhitespace(children)
	children = compactInlines(children, false)

	result := make([]inline, 0, 3)
	if leading != "" {
		result = append(result, inline{kind: inlineText, text: leading})
	}
	if hasInlineContent(children) {
		wrapped := wrapper
		wrapped.children = children
		result = append(result, wrapped)
	}
	if trailing != "" {
		result = append(result, inline{kind: inlineText, text: trailing})
	}
	return compactInlines(result, false)
}

func splitInlinesAtHardBreaks(values []inline) [][]inline {
	segments := make([][]inline, 0, 1)
	current := make([]inline, 0, len(values))
	flush := func() {
		if len(current) > 0 {
			segments = append(segments, current)
			current = nil
		}
	}
	for _, value := range values {
		if value.kind == inlineHardBreak {
			flush()
			segments = append(segments, []inline{value})
			continue
		}
		current = append(current, value)
	}
	flush()
	return segments
}

func flattenInlineKind(values []inline, kind inlineKind) []inline {
	flattened := make([]inline, 0, len(values))
	for _, value := range values {
		if value.kind == kind {
			flattened = append(flattened, flattenInlineKind(value.children, kind)...)
			continue
		}
		flattened = append(flattened, value)
	}
	return flattened
}

func trimLeadingInlineWhitespace(values []inline) ([]inline, string) {
	for index := range values {
		value := &values[index]
		switch value.kind {
		case inlineText:
			trimmed := strings.TrimLeftFunc(value.text, unicode.IsSpace)
			leading := value.text[:len(value.text)-len(trimmed)]
			value.text = trimmed
			return removeEmptyInlineText(values), leading
		case inlineCode, inlineImage, inlineHardBreak:
			return values, ""
		default:
			children, leading := trimLeadingInlineWhitespace(value.children)
			value.children = children
			if leading != "" {
				return values, leading
			}
			if hasInlineContent(value.children) {
				return values, ""
			}
		}
	}
	return removeEmptyInlineText(values), ""
}

func trimTrailingInlineWhitespace(values []inline) ([]inline, string) {
	for index := len(values) - 1; index >= 0; index-- {
		value := &values[index]
		switch value.kind {
		case inlineText:
			trimmed := strings.TrimRightFunc(value.text, unicode.IsSpace)
			trailing := value.text[len(trimmed):]
			value.text = trimmed
			return removeEmptyInlineText(values), trailing
		case inlineCode, inlineImage, inlineHardBreak:
			return values, ""
		default:
			children, trailing := trimTrailingInlineWhitespace(value.children)
			value.children = children
			if trailing != "" {
				return values, trailing
			}
			if hasInlineContent(value.children) {
				return values, ""
			}
		}
	}
	return removeEmptyInlineText(values), ""
}

func removeEmptyInlineText(values []inline) []inline {
	out := values[:0]
	for _, value := range values {
		if value.kind == inlineText && value.text == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func compactInlines(values []inline, trimEdges bool) []inline {
	compacted := make([]inline, 0, len(values))
	for _, value := range values {
		if value.kind == inlineText && value.text == "" {
			continue
		}
		if value.kind == inlineText && len(compacted) > 0 && compacted[len(compacted)-1].kind == inlineText {
			compacted[len(compacted)-1].text += value.text
			continue
		}
		compacted = append(compacted, value)
	}

	for i := range compacted {
		if compacted[i].kind == inlineText {
			compacted[i].text = collapseInlineSpaces(compacted[i].text)
		}
	}
	if trimEdges && len(compacted) > 0 {
		if compacted[0].kind == inlineText {
			compacted[0].text = strings.TrimLeftFunc(compacted[0].text, unicode.IsSpace)
		}
		last := len(compacted) - 1
		if compacted[last].kind == inlineText {
			compacted[last].text = strings.TrimRightFunc(compacted[last].text, unicode.IsSpace)
		}
	}

	out := compacted[:0]
	for _, value := range compacted {
		if value.kind != inlineText || value.text != "" {
			out = append(out, value)
		}
	}
	return out
}

func normalizeHTMLText(value string) string {
	return collapseInlineSpaces(value)
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func collapseInlineSpaces(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	leading := unicode.IsSpace(runes[0])
	trailing := unicode.IsSpace(runes[len(runes)-1])
	core := strings.Join(strings.Fields(value), " ")
	if core == "" {
		return " "
	}
	if leading {
		core = " " + core
	}
	if trailing {
		core += " "
	}
	return core
}

func hasInlineContent(values []inline) bool {
	for _, value := range values {
		switch value.kind {
		case inlineText, inlineCode:
			if strings.TrimSpace(value.text) != "" {
				return true
			}
		case inlineImage, inlineHardBreak:
			return true
		default:
			if hasInlineContent(value.children) {
				return true
			}
		}
	}
	return false
}

func isBlockNode(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "address", "article", "aside", "audio", "blockquote", "dd", "details", "div", "dl", "dt", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "iframe", "li", "main", "nav", "ol", "p", "pre", "section", "summary", "table", "ul", "video":
		return true
	default:
		return false
	}
}

func isIgnoredElement(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "head", "script", "style", "template", "noscript", "textarea", "svg", "canvas":
		return true
	default:
		return false
	}
}

func iframeLink(node *html.Node) []inline {
	src, _ := attribute(node, "src")
	src = strings.TrimSpace(src)
	if !isMeaningfulIframeSource(src) {
		return nil
	}

	label, _ := attribute(node, "title")
	if strings.TrimSpace(label) == "" {
		label, _ = attribute(node, "aria-label")
	}
	label = collapseWhitespace(label)
	if label == "" && isVideoEmbedSource(src) {
		label = "Embedded video"
	}
	if label == "" {
		return nil
	}

	return wrapSingleInlineNode(
		inline{kind: inlineLink, url: src},
		[]inline{{kind: inlineText, text: label}},
	)
}

func isMeaningfulIframeSource(src string) bool {
	return safeurl.Resource(src)
}

func mediaLink(node *html.Node, fallbackLabel string) []inline {
	src, _ := attribute(node, "src")
	src = strings.TrimSpace(src)
	if !safeurl.Resource(src) {
		src = ""
		stack := make([]*html.Node, 0, 4)
		for child := node.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
		for len(stack) > 0 && src == "" {
			last := len(stack) - 1
			current := stack[last]
			stack = stack[:last]
			if current.Type == html.ElementNode && strings.EqualFold(current.Data, "source") {
				candidate, _ := attribute(current, "src")
				candidate = strings.TrimSpace(candidate)
				if safeurl.Resource(candidate) {
					src = candidate
					break
				}
			}
			for child := current.LastChild; child != nil; child = child.PrevSibling {
				stack = append(stack, child)
			}
		}
	}

	label, _ := attribute(node, "title")
	if strings.TrimSpace(label) == "" {
		label, _ = attribute(node, "aria-label")
	}
	label = collapseWhitespace(label)
	if label == "" {
		label = fallbackLabel
	}
	if src == "" {
		return []inline{{kind: inlineText, text: label}}
	}
	return wrapSingleInlineNode(
		inline{kind: inlineLink, url: src},
		[]inline{{kind: inlineText, text: label}},
	)
}

func isVideoEmbedSource(src string) bool {
	parsed, err := url.Parse(src)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "youtube.com" || strings.HasSuffix(host, ".youtube.com") ||
		host == "youtube-nocookie.com" || strings.HasSuffix(host, ".youtube-nocookie.com") ||
		host == "youtu.be" || strings.HasSuffix(host, ".youtu.be") ||
		host == "vimeo.com" || strings.HasSuffix(host, ".vimeo.com")
}

func findElement(root *html.Node, name string) *html.Node {
	if root == nil {
		return nil
	}
	stack := []*html.Node{root}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, name) {
			return current
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return nil
}

func attribute(node *html.Node, name string) (string, bool) {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val, true
		}
	}
	return "", false
}

func hasAncestorBefore(node *html.Node, wanted, stop string) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type != html.ElementNode {
			continue
		}
		if strings.EqualFold(parent.Data, wanted) {
			return true
		}
		if strings.EqualFold(parent.Data, stop) {
			return false
		}
	}
	return false
}

func codeLanguage(node *html.Node) string {
	for current := node; current != nil; current = current.Parent {
		for _, attributeName := range []string{"data-language", "data-lang"} {
			if value, ok := attribute(current, attributeName); ok {
				if language := sanitizeLanguage(strings.ToLower(strings.TrimSpace(value))); language != "" {
					return language
				}
			}
		}

		className, _ := attribute(current, "class")
		tokens := strings.Fields(strings.ToLower(className))
		for index, token := range tokens {
			for _, prefix := range []string{"language-", "lang-"} {
				if strings.HasPrefix(token, prefix) {
					return sanitizeLanguage(strings.TrimPrefix(token, prefix))
				}
			}
			if strings.HasPrefix(token, "brush:") {
				language := strings.TrimPrefix(token, "brush:")
				if language == "" && index+1 < len(tokens) {
					language = tokens[index+1]
				}
				if language = sanitizeLanguage(language); language != "" {
					return language
				}
			}
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "body") {
			break
		}
	}
	return ""
}

func sanitizeLanguage(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func (p *parser) collectCodeText(node *html.Node, depth int) (string, error) {
	var builder strings.Builder
	lastWasNewline := false
	writeString := func(value string) {
		if value == "" {
			return
		}
		builder.WriteString(value)
		lastWasNewline = strings.HasSuffix(value, "\n") || strings.HasSuffix(value, "\r")
	}
	writeNewline := func() {
		builder.WriteByte('\n')
		lastWasNewline = true
	}

	var walk func(*html.Node, int) error
	walk = func(current *html.Node, currentDepth int) error {
		if current == nil {
			return nil
		}
		if err := p.visit(currentDepth); err != nil {
			return err
		}
		switch current.Type {
		case html.TextNode:
			writeString(current.Data)
			return nil
		case html.ElementNode:
			if current != node && isCodeGutter(current) {
				return nil
			}
			if strings.EqualFold(current.Data, "br") {
				writeNewline()
				return nil
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child, currentDepth+1); err != nil {
				return err
			}
		}
		if current != node && current.Type == html.ElementNode && isCodeLineContainer(current.Data) && builder.Len() > 0 && !lastWasNewline {
			writeNewline()
		}
		return nil
	}
	if err := walk(node, depth); err != nil {
		return "", err
	}
	return strings.ReplaceAll(builder.String(), "\r\n", "\n"), nil
}

func isCodeGutter(node *html.Node) bool {
	className, _ := attribute(node, "class")
	className = strings.ToLower(className)
	return strings.Contains(className, "gutter") || strings.Contains(className, "line-numbers")
}

func isCodeLineContainer(name string) bool {
	switch strings.ToLower(name) {
	case "div", "p", "li", "tr":
		return true
	default:
		return false
	}
}
