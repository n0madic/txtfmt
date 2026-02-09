package ast

import "github.com/n0madic/txtfmt/internal/config"

type Pos struct {
	Off  int
	Line int
	Col  int
}

type Diag struct {
	Pos     Pos
	Code    string
	Message string
}

type Document struct {
	Lang   config.Lang
	Style  config.Style
	Blocks []Block
	Diags  []Diag
}

type Block interface{ isBlock() }

type Paragraph struct{ In []Inline }

func (Paragraph) isBlock() {}

type DialogueBlock struct{ Turns []DialogueTurn }

func (DialogueBlock) isBlock() {}

type DialogueTurn struct {
	In []Inline
}

type Heading struct {
	Level int
	In    []Inline
}

func (Heading) isBlock() {}

type TitleBlock struct {
	In []Inline
}

func (TitleBlock) isBlock() {}

type ContentsBlock struct {
	In      []Inline
	Entries []ContentsEntry
}

func (ContentsBlock) isBlock() {}

type ContentsEntry struct {
	Level int
	In    []Inline
}

type MetaLineBlock struct {
	Key string
	In  []Inline
}

func (MetaLineBlock) isBlock() {}

type SceneBreak struct{ Marker string }

func (SceneBreak) isBlock() {}

type Inline interface{ isInline() }

type SpaceKind int

const (
	SpaceNormal SpaceKind = iota
	SpaceNBSP
	SpaceThin
)

type DashKind int

const (
	DashHyphen DashKind = iota
	DashNDash
	DashEmDash
)

type QuoteLevel int

const (
	QuotePrimary QuoteLevel = iota + 1
	QuoteSecondary
)

type Word struct{ S string }

func (Word) isInline() {}

type Space struct{ Kind SpaceKind }

func (Space) isInline() {}

type Punct struct{ Ch rune }

func (Punct) isInline() {}

type Dash struct{ Kind DashKind }

func (Dash) isInline() {}

type Ellipsis struct{}

func (Ellipsis) isInline() {}

type QuoteSpan struct {
	Level QuoteLevel
	In    []Inline
}

func (QuoteSpan) isInline() {}

type ParenSpan struct {
	Open  rune
	Close rune
	In    []Inline
}

func (ParenSpan) isInline() {}
