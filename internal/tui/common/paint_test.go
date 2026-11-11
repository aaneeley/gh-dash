package common

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestPaintBackground(t *testing.T) {
	bg := lipgloss.Color("#141414")
	bgSeq := "\x1b[48;2;20;20;20m"

	t.Run("pads every line to the given width", func(t *testing.T) {
		out := PaintBackground("ab\n\nabcd", bg, 4)
		lines := strings.Split(out, "\n")
		require.Len(t, lines, 3)
		for _, line := range lines {
			require.Equal(t, 4, lipgloss.Width(line))
			require.True(t, strings.HasPrefix(line, bgSeq))
			require.True(t, strings.HasSuffix(line, "\x1b[m"))
		}
		require.Equal(t, "ab  \n\u00a0   \nabcd", ansi.Strip(out))
	})

	t.Run("restores the background after resets", func(t *testing.T) {
		for _, reset := range []string{"\x1b[m", "\x1b[0m", "\x1b[49m", "\x1b[0;1m", "\x1b[1;49m"} {
			out := PaintBackground("a"+reset+"b", bg, 2)
			require.Equal(t, bgSeq+"a"+reset+bgSeq+"b\x1b[m", out, "reset %q", reset)
		}
	})

	t.Run("keeps explicit colors", func(t *testing.T) {
		for _, seq := range []string{
			"\x1b[38;2;0;0;0m",      // 0 is a color component, not a reset
			"\x1b[38;5;0m",          // same for indexed colors
			"\x1b[48;2;30;30;30m",   // background set by the content itself
			"\x1b[0;48;2;30;30;30m", // reset followed by a background
			"\x1b[41m",
			"\x1b[1m",
		} {
			out := PaintBackground("a"+seq+"b", bg, 2)
			require.Equal(t, bgSeq+"a"+seq+"b\x1b[m", out, "sequence %q", seq)
		}
	})

	t.Run("leaves other escape sequences alone", func(t *testing.T) {
		out := PaintBackground("a\x1b[2Kb", bg, 2)
		require.Equal(t, bgSeq+"a\x1b[2Kb\x1b[m", out)
	})
}

func TestPaintBackground_CutsLinesThatAreTooWide(t *testing.T) {
	out := PaintBackground("abcdef\nab", lipgloss.Color("#141414"), 4)
	require.Equal(t, "abcd\nab  ", ansi.Strip(out))
}

func TestFitWidth(t *testing.T) {
	styled := lipgloss.NewStyle().Background(lipgloss.Color("#141414")).Render("   ")
	out := FitWidth("ab\nabcdef\n"+styled, 4)

	lines := strings.Split(out, "\n")
	require.Len(t, lines, 3)
	for _, line := range lines {
		require.Equal(t, 4, lipgloss.Width(line))
	}
	require.Equal(t, "ab  \nabcd\n    ", ansi.Strip(out))
	require.True(t, strings.HasPrefix(lines[2], styled),
		"whitespace that carries a style should be kept as it is")
}
