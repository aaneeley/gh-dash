package context

import (
	bbHelp "charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

type SelectStyles struct {
	PopupStyle    lipgloss.Style
	HelpStyle     lipgloss.Style
	Pointer       lipgloss.Style
	ItemStyle     lipgloss.Style
	Status        lipgloss.Style
	SelectedStyle lipgloss.Style
	NoResults     lipgloss.Style
}

type Styles struct {
	Colors struct {
		OpenIssue   compat.AdaptiveColor
		ClosedIssue compat.AdaptiveColor
		SuccessText compat.AdaptiveColor
		OpenPR      compat.AdaptiveColor
		ClosedPR    compat.AdaptiveColor
		MergedPR    compat.AdaptiveColor
	}

	Common common.CommonStyles

	Search struct {
		Root lipgloss.Style
	}

	PrView struct {
		PillStyle lipgloss.Style
	}
	Help struct {
		Text         lipgloss.Style
		KeyText      lipgloss.Style
		BubbleStyles bbHelp.Styles
	}
	CommentBox struct {
		Text lipgloss.Style
	}
	Pager struct {
		Height int
		Root   lipgloss.Style
	}
	Section struct {
		ContainerPadding int
		ContainerStyle   lipgloss.Style
		SpinnerStyle     lipgloss.Style
		EmptyStateStyle  lipgloss.Style
		KeyStyle         lipgloss.Style
	}
	PrSection struct {
		CiCellWidth        int
		LinesCellWidth     int
		UpdatedAtCellWidth int
		PrRepoCellWidth    int
		PrAuthorCellWidth  int
	}
	Sidebar struct {
		BorderWidth    int
		PagerHeight    int
		ContentPadding int
		PagerStyle     lipgloss.Style
		InputBox       lipgloss.Style
	}
	ListViewPort struct {
		PagerStyle lipgloss.Style
	}
	Table struct {
		CellStyle                lipgloss.Style
		SelectedCellStyle        lipgloss.Style
		TitleCellStyle           lipgloss.Style
		Gutter                   lipgloss.Style
		SelectedGutter           lipgloss.Style
		SingleRuneTitleCellStyle lipgloss.Style
		HeaderStyle              lipgloss.Style
		RowStyle                 lipgloss.Style
	}
	Tabs struct {
		Tab               lipgloss.Style
		ActiveTab         lipgloss.Style
		OverflowIndicator lipgloss.Style
		TabSeparator      lipgloss.Style
		TabsRow           lipgloss.Style
	}
	ViewSwitcher struct {
		ActiveView     lipgloss.Style
		InactiveView   lipgloss.Style
		Root           lipgloss.Style
		ViewsSeparator lipgloss.Style
	}
	Select  SelectStyles
	KeyHint lipgloss.Style
}

var DefaultStyles = Styles{
	Colors: struct {
		OpenIssue   compat.AdaptiveColor
		ClosedIssue compat.AdaptiveColor
		SuccessText compat.AdaptiveColor
		OpenPR      compat.AdaptiveColor
		ClosedPR    compat.AdaptiveColor
		MergedPR    compat.AdaptiveColor
	}{
		OpenIssue: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
		ClosedIssue: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
		SuccessText: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
		OpenPR: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
		ClosedPR: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
		MergedPR: compat.AdaptiveColor{
			Light: lipgloss.NoColor{},
			Dark:  lipgloss.NoColor{},
		},
	},
}

// BarBorder is the thick vertical bar that marks blocks such as inputs,
// comments and the selected row.
var BarBorder = common.BarBorder

var LogoColor = lipgloss.Color("#89b4fa")

func InitStyles(theme theme.Theme) Styles {
	var s Styles

	s.Colors.OpenIssue = theme.SuccessText
	s.Colors.ClosedIssue = theme.ErrorText
	s.Colors.SuccessText = theme.SuccessText
	s.Colors.OpenPR = theme.SuccessText
	s.Colors.ClosedPR = theme.ErrorText
	s.Colors.MergedPR = theme.Secondary

	s.Common = common.BuildStyles(theme)

	// the search input is a block with a bar on its left, like a prompt
	s.Search.Root = lipgloss.NewStyle().
		Border(BarBorder, false, false, false, true).
		BorderForeground(theme.PrimaryBorder)

	s.PrView.PillStyle = s.Common.MainTextStyle.
		Border(lipgloss.Border{Left: "", Right: ""}, false, true, false, true).
		Foreground(theme.InvertedText)

	s.Help.Text = lipgloss.NewStyle().Foreground(theme.SecondaryText)
	s.Help.KeyText = lipgloss.NewStyle().Foreground(theme.PrimaryText)
	s.Help.BubbleStyles = bbHelp.Styles{
		ShortDesc:      s.Help.Text.Foreground(theme.FaintText),
		FullDesc:       s.Help.Text.Foreground(theme.FaintText),
		ShortSeparator: s.Help.Text.Foreground(theme.FaintBorder),
		FullSeparator:  s.Help.Text,
		FullKey:        s.Help.KeyText,
		ShortKey:       s.Help.KeyText,
		Ellipsis:       s.Help.Text,
	}

	s.CommentBox.Text = s.Help.Text

	s.Pager.Height = 2
	s.Pager.Root = lipgloss.NewStyle().
		Height(s.Pager.Height).
		MaxHeight(s.Pager.Height).
		PaddingTop(1).
		Foreground(theme.FaintText)

	s.Section.ContainerPadding = 1
	s.Section.ContainerStyle = lipgloss.NewStyle().
		Padding(0, s.Section.ContainerPadding)
	s.Section.SpinnerStyle = lipgloss.NewStyle().Padding(0, 1)
	s.Section.EmptyStateStyle = lipgloss.NewStyle().
		Faint(true).
		PaddingLeft(1).
		MarginBottom(1)
	s.Section.KeyStyle = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Background(theme.SelectedBackground).
		Padding(0, 1)

	s.PrSection.CiCellWidth = lipgloss.Width(" CI ")
	s.PrSection.LinesCellWidth = lipgloss.Width(" 123450 / -123450 ")
	s.PrSection.UpdatedAtCellWidth = lipgloss.Width("2mo  ")
	s.PrSection.PrRepoCellWidth = 15
	s.PrSection.PrAuthorCellWidth = 15

	s.Sidebar.PagerHeight = 1
	s.Sidebar.BorderWidth = 1
	s.Sidebar.ContentPadding = 2
	// The sidebar is a raised panel instead of a bordered box. It paints the
	// panel itself, see sidebar.View. BorderWidth is the gap around it.
	s.Sidebar.PagerStyle = lipgloss.NewStyle().
		Height(s.Sidebar.PagerHeight).
		PaddingLeft(s.Sidebar.ContentPadding).
		Foreground(theme.FaintText)

	s.Sidebar.InputBox = lipgloss.NewStyle().
		MarginTop(1).
		PaddingLeft(1).
		Border(BarBorder, false, false, false, true).
		BorderForeground(theme.Accent)

	s.ListViewPort.PagerStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(theme.FaintText)

	s.Table.CellStyle = lipgloss.NewStyle().PaddingLeft(1).
		PaddingRight(1).
		MaxHeight(1)
	s.Table.SelectedCellStyle = s.Table.CellStyle.
		Background(theme.SelectedBackground)
	s.Table.TitleCellStyle = s.Table.CellStyle.
		Foreground(theme.FaintText)
	s.Table.SingleRuneTitleCellStyle = s.Table.TitleCellStyle.
		Width(common.SingleRuneWidth)
	s.Table.HeaderStyle = lipgloss.NewStyle()
	// rows are separated by whitespace instead of rules
	s.Table.RowStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.Border{Bottom: " "})
	s.Table.Gutter = lipgloss.NewStyle()
	s.Table.SelectedGutter = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Background(theme.SelectedBackground)

	s.Tabs.Tab = lipgloss.NewStyle().
		Foreground(theme.FaintText).
		Padding(0, 2)
	s.Tabs.ActiveTab = s.Tabs.Tab.
		Bold(true).
		Background(theme.SelectedBackground).
		Foreground(theme.PrimaryText)
	s.Tabs.OverflowIndicator = s.Common.FaintTextStyle.Padding(0, 1)
	s.Tabs.TabSeparator = lipgloss.NewStyle().
		Foreground(theme.FaintBorder)
	s.Tabs.TabsRow = lipgloss.NewStyle().
		Height(common.TabsContentHeight).
		BorderBottom(true).
		BorderStyle(lipgloss.Border{Bottom: " "})
	s.ViewSwitcher.Root = lipgloss.NewStyle()

	s.ViewSwitcher.ActiveView = lipgloss.NewStyle().
		Foreground(theme.PrimaryText).
		Bold(true)
	s.ViewSwitcher.ViewsSeparator = lipgloss.NewStyle().
		Foreground(theme.FaintBorder)
	s.ViewSwitcher.InactiveView = lipgloss.NewStyle().
		Foreground(theme.FaintText)
	s.Select.PopupStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryBorder).
		Foreground(theme.PrimaryText)
	s.Select.HelpStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderTopForeground(theme.FaintBorder)
	s.Select.Pointer = lipgloss.NewStyle().
		Foreground(theme.Accent)
	s.Select.ItemStyle = lipgloss.NewStyle().
		PaddingRight(1)
	s.Select.SelectedStyle = s.Select.ItemStyle.
		Foreground(theme.PrimaryText)
	s.Select.NoResults = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Foreground(theme.FaintText).
		Italic(true)
	s.Select.Status = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Foreground(theme.FaintText).
		Italic(true)

	s.KeyHint = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		Background(theme.SelectedBackground).
		Foreground(theme.PrimaryText)

	return s
}
