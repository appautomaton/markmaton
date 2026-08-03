package native

type blockKind uint8

const (
	blockParagraph blockKind = iota + 1
	blockHeading
	blockQuote
	blockCode
	blockList
	blockTable
	blockThematicBreak
)

type inlineKind uint8

const (
	inlineText inlineKind = iota + 1
	inlineEmphasis
	inlineStrong
	inlineCode
	inlineLink
	inlineImage
	inlineStrikethrough
	inlineHardBreak
)

type document struct {
	blocks []block
}

type block struct {
	kind     blockKind
	level    int
	inlines  []inline
	children []block
	items    []listItem
	ordered  bool
	reversed bool
	start    int
	code     string
	language string
	table    table
}

type listItem struct {
	blocks           []block
	hasTask          bool
	checked          bool
	hasExplicitValue bool
	explicitValue    int
}

type table struct {
	caption []inline
	rows    []tableRow
	complex bool
}

type tableRow struct {
	header bool
	cells  []tableCell
}

type tableCell struct {
	inlines   []inline
	alignment tableAlignment
}

type tableAlignment uint8

const (
	tableAlignDefault tableAlignment = iota
	tableAlignLeft
	tableAlignCenter
	tableAlignRight
)

type inline struct {
	kind     inlineKind
	text     string
	url      string
	title    string
	children []inline
}
