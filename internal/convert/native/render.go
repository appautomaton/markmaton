package native

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func renderDocument(ctx context.Context, doc document) (string, error) {
	markdown, err := renderBlocks(ctx, doc.blocks)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(markdown), nil
}

func renderBlocks(ctx context.Context, blocks []block) (string, error) {
	parts := make([]string, 0, len(blocks))
	for _, value := range blocks {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		rendered, err := renderBlock(ctx, value)
		if err != nil {
			return "", err
		}
		rendered = strings.Trim(rendered, "\n")
		if strings.TrimSpace(rendered) != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func renderBlock(ctx context.Context, value block) (string, error) {
	switch value.kind {
	case blockParagraph:
		return renderInlines(ctx, value.inlines)
	case blockHeading:
		content, err := renderInlines(ctx, value.inlines)
		if err != nil {
			return "", err
		}
		level := value.level
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		return strings.Repeat("#", level) + " " + strings.TrimSpace(content), nil
	case blockQuote:
		content, err := renderBlocks(ctx, value.children)
		if err != nil {
			return "", err
		}
		if content == "" {
			return "", nil
		}
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if line == "" {
				lines[i] = ">"
			} else {
				lines[i] = "> " + line
			}
		}
		return strings.Join(lines, "\n"), nil
	case blockCode:
		fence, err := codeFence(ctx, value.code, '`', 3)
		if err != nil {
			return "", err
		}
		return fence + value.language + "\n" + value.code + "\n" + fence, nil
	case blockList:
		return renderList(ctx, value)
	case blockTable:
		return renderTable(ctx, value.table)
	case blockThematicBreak:
		return "---", nil
	default:
		return "", nil
	}
}

func renderInlines(ctx context.Context, values []inline) (string, error) {
	var builder strings.Builder
	for index, value := range values {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var rendered string
		var err error
		if value.kind == inlineEmphasis {
			rendered, err = renderEmphasis(ctx, values, index)
		} else {
			rendered, err = renderInline(ctx, value)
		}
		if err != nil {
			return "", err
		}
		builder.WriteString(rendered)
	}
	return builder.String(), nil
}

func renderEmphasis(ctx context.Context, siblings []inline, index int) (string, error) {
	content, err := renderInlines(ctx, siblings[index].children)
	if err != nil {
		return "", err
	}
	delimiter := "_"
	if hasWordBoundaryCollision(siblings, index) {
		delimiter = "*"
	}
	return renderDelimitedInline(content, delimiter), nil
}

func hasWordBoundaryCollision(siblings []inline, index int) bool {
	if index > 0 {
		previous := siblings[index-1]
		if previous.kind == inlineText {
			runes := []rune(previous.text)
			if len(runes) > 0 && (unicode.IsLetter(runes[len(runes)-1]) || unicode.IsDigit(runes[len(runes)-1])) {
				return true
			}
		}
	}
	if index+1 < len(siblings) {
		next := siblings[index+1]
		if next.kind == inlineText {
			runes := []rune(next.text)
			if len(runes) > 0 && (unicode.IsLetter(runes[0]) || unicode.IsDigit(runes[0])) {
				return true
			}
		}
	}
	return false
}

func renderInline(ctx context.Context, value inline) (string, error) {
	switch value.kind {
	case inlineText:
		return escapeText(ctx, value.text)
	case inlineCode:
		return renderInlineCode(ctx, value.text)
	case inlineImage:
		alt := escapeLabel(value.text)
		destination := renderDestination(value.url)
		title := renderTitle(value.title)
		return "![" + alt + "](" + destination + title + ")", nil
	case inlineHardBreak:
		return "\\\n", nil
	}

	content, err := renderInlines(ctx, value.children)
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", nil
	}

	switch value.kind {
	case inlineEmphasis:
		return renderDelimitedInline(content, "_"), nil
	case inlineStrong:
		return renderDelimitedInline(content, "**"), nil
	case inlineStrikethrough:
		return renderDelimitedInline(content, "~~"), nil
	case inlineLink:
		if value.title == "" && len(value.children) == 1 && value.children[0].kind == inlineImage && strings.TrimSpace(value.children[0].url) == strings.TrimSpace(value.url) {
			return content, nil
		}
		return "[" + content + "](" + renderDestination(value.url) + renderTitle(value.title) + ")", nil
	default:
		return content, nil
	}
}

func renderList(ctx context.Context, value block) (string, error) {
	var builder strings.Builder
	currentNumber := value.start
	for _, item := range value.items {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		content, err := renderListItemBlocks(ctx, item.blocks)
		if err != nil {
			return "", err
		}
		content = strings.TrimSpace(content)
		if content == "" && !item.hasTask {
			continue
		}

		prefix := "- "
		if value.ordered {
			if item.hasExplicitValue {
				currentNumber = item.explicitValue
			}
			prefix = strconv.Itoa(currentNumber) + ". "
			if value.reversed {
				currentNumber--
			} else {
				currentNumber++
			}
		}
		continuation := strings.Repeat(" ", len([]rune(prefix)))
		if !item.hasTask && len(item.blocks) == 1 && item.blocks[0].kind == blockList {
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			lines := strings.Split(content, "\n")
			for lineIndex, line := range lines {
				if lineIndex > 0 {
					builder.WriteByte('\n')
				}
				if line != "" {
					builder.WriteString(continuation)
					builder.WriteString(line)
				}
			}
			continue
		}
		lines := strings.Split(content, "\n")
		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(prefix)
		if item.hasTask {
			if item.checked {
				builder.WriteString("[x] ")
			} else {
				builder.WriteString("[ ] ")
			}
		}
		builder.WriteString(lines[0])
		for _, line := range lines[1:] {
			builder.WriteByte('\n')
			if line != "" {
				builder.WriteString(continuation)
				builder.WriteString(line)
			}
		}
	}
	return builder.String(), nil
}

func renderListItemBlocks(ctx context.Context, blocks []block) (string, error) {
	parts := make([]string, 0, len(blocks))
	for index, value := range blocks {
		rendered, err := renderBlock(ctx, value)
		if err != nil {
			return "", err
		}
		rendered = strings.Trim(rendered, "\n")
		if strings.TrimSpace(rendered) == "" {
			continue
		}
		if index > 0 && value.kind == blockList && len(parts) > 0 {
			parts[len(parts)-1] += "\n" + rendered
			continue
		}
		parts = append(parts, rendered)
	}
	return strings.Join(parts, "\n\n"), nil
}

func renderTable(ctx context.Context, value table) (string, error) {
	if value.complex {
		return renderTableFallback(ctx, value)
	}
	if len(value.rows) == 0 {
		if hasInlineContent(value.caption) {
			return renderInlines(ctx, value.caption)
		}
		return "", nil
	}

	columnCount := 0
	for _, row := range value.rows {
		if len(row.cells) > columnCount {
			columnCount = len(row.cells)
		}
	}
	if columnCount == 0 {
		return "", nil
	}

	headerIndex := -1
	for i, row := range value.rows {
		if row.header {
			headerIndex = i
			break
		}
	}

	alignments := tableAlignments(value.rows, headerIndex, columnCount)
	var lines []string
	if headerIndex >= 0 {
		line, err := renderTableRow(ctx, value.rows[headerIndex], columnCount)
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	} else {
		lines = append(lines, renderEmptyTableRow(columnCount))
	}
	lines = append(lines, renderTableDivider(alignments))

	for i, row := range value.rows {
		if i == headerIndex {
			continue
		}
		line, err := renderTableRow(ctx, row, columnCount)
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	}
	tableMarkdown := strings.Join(lines, "\n")
	if !hasInlineContent(value.caption) {
		return tableMarkdown, nil
	}
	caption, err := renderInlines(ctx, value.caption)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(caption) + "\n\n" + tableMarkdown, nil
}

func renderTableFallback(ctx context.Context, value table) (string, error) {
	parts := make([]string, 0, len(value.rows)+1)
	if hasInlineContent(value.caption) {
		caption, err := renderInlines(ctx, value.caption)
		if err != nil {
			return "", err
		}
		if caption = strings.TrimSpace(caption); caption != "" {
			parts = append(parts, caption)
		}
	}
	for _, row := range value.rows {
		cells := make([]string, 0, len(row.cells))
		for _, cell := range row.cells {
			content, err := renderInlines(ctx, cell.inlines)
			if err != nil {
				return "", err
			}
			if content = strings.TrimSpace(strings.ReplaceAll(content, "\\\n", " ")); content != "" {
				cells = append(cells, content)
			}
		}
		if len(cells) > 0 {
			parts = append(parts, strings.Join(cells, " · "))
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func renderTableRow(ctx context.Context, row tableRow, columns int) (string, error) {
	cells := make([]string, columns)
	for i := 0; i < columns; i++ {
		if i >= len(row.cells) {
			continue
		}
		content, err := renderInlines(ctx, row.cells[i].inlines)
		if err != nil {
			return "", err
		}
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\\\n", "<br>")
		content = strings.ReplaceAll(content, "\n", "<br>")
		content = strings.ReplaceAll(content, "|", "\\|")
		cells[i] = strings.TrimSpace(content)
	}
	return "| " + strings.Join(cells, " | ") + " |", nil
}

func renderEmptyTableRow(columns int) string {
	return "| " + strings.Join(make([]string, columns), " | ") + " |"
}

func tableAlignments(rows []tableRow, headerIndex, columns int) []tableAlignment {
	alignments := make([]tableAlignment, columns)
	if headerIndex >= 0 {
		for index, cell := range rows[headerIndex].cells {
			if index < columns {
				alignments[index] = cell.alignment
			}
		}
		return alignments
	}
	for _, row := range rows {
		for index, cell := range row.cells {
			if index < columns && alignments[index] == tableAlignDefault {
				alignments[index] = cell.alignment
			}
		}
	}
	return alignments
}

func renderTableDivider(alignments []tableAlignment) string {
	cells := make([]string, len(alignments))
	for index, alignment := range alignments {
		switch alignment {
		case tableAlignLeft:
			cells[index] = ":--"
		case tableAlignCenter:
			cells[index] = ":-:"
		case tableAlignRight:
			cells[index] = "--:"
		default:
			cells[index] = "---"
		}
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

func renderDelimitedInline(content, delimiter string) string {
	if strings.TrimSpace(content) == "" {
		return content
	}
	leadingLength := len(content) - len(strings.TrimLeftFunc(content, unicode.IsSpace))
	trailingStart := len(strings.TrimRightFunc(content, unicode.IsSpace))
	leading := content[:leadingLength]
	trailing := content[trailingStart:]
	core := content[leadingLength:trailingStart]
	return leading + delimiter + core + delimiter + trailing
}

func renderInlineCode(ctx context.Context, value string) (string, error) {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", " ")
	fence, err := codeFence(ctx, value, '`', 1)
	if err != nil {
		return "", err
	}
	isAllSpaces := value != "" && strings.Trim(value, " ") == ""
	if !isAllSpaces && (strings.HasPrefix(value, "`") || strings.HasSuffix(value, "`") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ")) {
		value = " " + value + " "
	}
	return fence + value + fence, nil
}

func codeFence(ctx context.Context, content string, marker rune, minimum int) (string, error) {
	maxRun := 0
	current := 0
	processed := 0
	for _, r := range content {
		processed++
		if processed%4_096 == 0 {
			if err := ctx.Err(); err != nil {
				return "", err
			}
		}
		if r == marker {
			current++
			if current > maxRun {
				maxRun = current
			}
			continue
		}
		current = 0
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	length := maxRun + 1
	if length < minimum {
		length = minimum
	}
	return strings.Repeat(string(marker), length), nil
}

var (
	orderedLineMarker   = regexp.MustCompile(`(?m)^( {0,3}\d+)\. `)
	unorderedLineMarker = regexp.MustCompile(`(?m)^( {0,3})([-+]) `)
	blockquoteMarker    = regexp.MustCompile(`(?m)^( {0,3})> `)
	hyphenThematicLine  = regexp.MustCompile(`(?m)^ {0,3}(?:- *){3,}$`)
	equalsSetextLine    = regexp.MustCompile(`(?m)^ {0,3}=+ *$`)
	strikethroughRun    = regexp.MustCompile(`~{2,}`)
)

func escapeText(ctx context.Context, value string) (string, error) {
	var builder strings.Builder
	processed := 0
	for _, r := range value {
		processed++
		if processed%4_096 == 0 {
			if err := ctx.Err(); err != nil {
				return "", err
			}
		}
		switch r {
		case '\\', '`', '*', '_', '[', ']', '#', '<', '>':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		default:
			builder.WriteRune(r)
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	escaped := builder.String()
	escaped = orderedLineMarker.ReplaceAllString(escaped, `${1}\. `)
	escaped = unorderedLineMarker.ReplaceAllString(escaped, `${1}\${2} `)
	escaped = blockquoteMarker.ReplaceAllString(escaped, `${1}\> `)
	escaped = hyphenThematicLine.ReplaceAllStringFunc(escaped, escapeFirstStructuralRune)
	escaped = equalsSetextLine.ReplaceAllStringFunc(escaped, escapeFirstStructuralRune)
	escaped = strikethroughRun.ReplaceAllStringFunc(escaped, func(run string) string {
		return strings.ReplaceAll(run, "~", `\~`)
	})
	return escaped, nil
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "[", "\\[")
	value = strings.ReplaceAll(value, "]", "\\]")
	return value
}

func escapeFirstStructuralRune(line string) string {
	for index, r := range line {
		if r == '-' || r == '_' || r == '*' || r == '=' {
			return line[:index] + `\` + line[index:]
		}
	}
	return line
}

func renderDestination(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		value = strings.Join(strings.Fields(value), "%20")
	}
	value = strings.ReplaceAll(value, "<", "%3C")
	value = strings.ReplaceAll(value, ">", "%3E")
	if strings.ContainsAny(value, "()") {
		return "<" + value + ">"
	}
	return strings.ReplaceAll(value, "\\", "\\\\")
}

func renderTitle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return fmt.Sprintf(" \"%s\"", value)
}
