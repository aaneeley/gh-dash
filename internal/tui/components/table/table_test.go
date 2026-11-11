package table

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestTable(t *testing.T, width int) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := context.ProgramContext{Config: &cfg, Theme: thm, Styles: context.InitStyles(thm)}
	grow := true
	stateWidth := 3
	m := NewModel(
		ctx,
		constants.Dimensions{Width: width, Height: 20},
		time.Now(),
		time.Now(),
		[]Column{{Title: "S", Width: &stateWidth}, {Title: "Title", Grow: &grow}},
		nil,
		"PR",
		nil,
		"Loading...",
		false,
	)
	m.SetRows([]Row{{"o", "first"}, {"o", "second"}})
	return m
}

// The gutter that marks the selected row is part of the table's width.
func TestRowsAndHeaderFitTheTableWidth(t *testing.T) {
	const width = 60
	m := newTestTable(t, width)

	for i, line := range strings.Split(m.View(), "\n") {
		require.LessOrEqual(t, lipgloss.Width(line), width, "line %d", i)
	}
	for rowId := range m.Rows {
		row := m.renderRow(rowId, m.renderHeaderColumns())
		for _, line := range strings.Split(row, "\n") {
			require.Equal(t, width, lipgloss.Width(line))
		}
	}
}

func TestOnlyTheSelectedRowHasABar(t *testing.T) {
	m := newTestTable(t, 60)
	headerColumns := m.renderHeaderColumns()

	selected := ansi.Strip(m.renderRow(0, headerColumns))
	other := ansi.Strip(m.renderRow(1, headerColumns))
	require.True(t, strings.HasPrefix(selected, "┃"), "got %q", selected)
	require.True(t, strings.HasPrefix(other, " "), "got %q", other)

	m.NextItem()
	require.True(t, strings.HasPrefix(ansi.Strip(m.renderRow(1, m.renderHeaderColumns())), "┃"))
	require.True(t, strings.HasPrefix(ansi.Strip(m.renderRow(0, m.renderHeaderColumns())), " "))
}
