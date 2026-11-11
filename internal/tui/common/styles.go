package common

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

var (
	HeaderHeight       = 2
	SearchHeight       = 3
	FooterHeight       = 1
	ExpandedHelpHeight = 17
	InputBoxHeight     = 8
	SingleRuneWidth    = 4
	MainContentPadding = 1
	TabsBorderHeight   = 1
	TabsContentHeight  = 1
	TabsHeight         = TabsBorderHeight + TabsContentHeight
	ViewSwitcherMargin = 1
	TableHeaderHeight  = 2
)

type CommonStyles struct {
	MainTextStyle       lipgloss.Style
	FaintTextStyle      lipgloss.Style
	FooterStyle         lipgloss.Style
	ErrorStyle          lipgloss.Style
	SuccessStyle        lipgloss.Style
	DraftGlyph          string
	PersonGlyph         string
	WaitingGlyph        string
	WaitingDotGlyph     string
	FailureGlyph        string
	SuccessGlyph        string
	MergedGlyph         string
	CommentGlyph        string
	ActionRequiredGlyph string
}

func BuildStyles(theme theme.Theme) CommonStyles {
	var s CommonStyles

	s.MainTextStyle = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Bold(true)
	s.FaintTextStyle = lipgloss.NewStyle().
		Foreground(theme.FaintText)
	s.FooterStyle = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Height(FooterHeight)

	s.ErrorStyle = lipgloss.NewStyle().Foreground(theme.ErrorText)
	s.SuccessStyle = lipgloss.NewStyle().Foreground(theme.SuccessText)

	s.PersonGlyph = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Render(constants.PersonIcon)
	s.WaitingGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Render(constants.WaitingIcon)
	s.WaitingDotGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Render(constants.DotIcon)
	s.FailureGlyph = lipgloss.NewStyle().
		Foreground(theme.ErrorText).
		Render(constants.FailureIcon)
	s.SuccessGlyph = lipgloss.NewStyle().
		Foreground(theme.SuccessText).
		Render(constants.SuccessIcon)
	s.CommentGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Render(constants.CommentIcon)
	s.DraftGlyph = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Render(constants.DraftIcon)
	s.MergedGlyph = lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Render(constants.MergedIcon)
	s.ActionRequiredGlyph = lipgloss.NewStyle().
		Foreground(theme.WarningText).
		Render(constants.ActionRequiredIcon)
	return s
}

// PreviewHeaderPadding is the horizontal padding of the header of preview
// panes, it lines the header up with the content below it.
const PreviewHeaderPadding = 2

// RenderPreviewHeader renders the repo/number line at the top of preview panes
// (e.g., "owner/repo · #123" or "#123 · owner/repo")
func RenderPreviewHeader(theme theme.Theme, width int, text string) string {
	return lipgloss.NewStyle().
		Padding(1, PreviewHeaderPadding, 0).
		Width(width).
		Foreground(theme.FaintText).
		Render(text)
}

// RenderPreviewTitle renders the title of preview panes
func RenderPreviewTitle(theme theme.Theme, styles CommonStyles, width int, title string) string {
	return styles.MainTextStyle.
		Padding(0, PreviewHeaderPadding).
		Width(width).
		Render(title)
}

// BarBorder is the thick vertical bar that marks blocks such as inputs,
// comments and the selected row.
var BarBorder = lipgloss.Border{Left: "┃"}

// RenderBlock renders content as a block: it is set off from its surroundings
// by a darker background and a bar in the given color on its left. The width
// includes the bar.
func RenderBlock(theme theme.Theme, content string, bar color.Color, width int) string {
	return lipgloss.NewStyle().
		Border(BarBorder, false, false, false, true).
		BorderForeground(bar).
		Render(PaintBackground(content, theme.ElementBackground, max(0, width-1)))
}

// RenderQuote marks content with a bar on its left, without changing its
// background.
func RenderQuote(content string, bar color.Color) string {
	return lipgloss.NewStyle().
		Border(BarBorder, false, false, false, true).
		BorderForeground(bar).
		PaddingLeft(1).
		Render(content)
}
