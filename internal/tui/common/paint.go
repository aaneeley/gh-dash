package common

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const nbsp = "\u00a0"

// PaintBackground fills s with the given background color. Unlike setting a
// background on a lipgloss style, this also covers content that was already
// rendered: every place where a nested style resets its attributes falls back
// to bg instead of the terminal's default background. Parts that set their
// own background keep it.
//
// Every line is padded or cut to width, so the result is a solid block.
func PaintBackground(s string, bg color.Color, width int) string {
	bgSeq := ansi.Style{}.BackgroundColor(bg).String()
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			line = ansi.Truncate(line, width, "")
		}
		var b strings.Builder
		b.WriteString(bgSeq)
		b.WriteString(restoreBackground(line, bgSeq))
		if pad := width - lipgloss.Width(line); pad > 0 {
			if strings.TrimSpace(ansi.Strip(line)) == "" {
				// Lines that only consist of whitespace lose their styling
				// when they are wrapped by lipgloss later on, a non-breaking
				// space prevents that.
				b.WriteString(nbsp)
				pad--
			}
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString(ansi.ResetStyle)
		lines[i] = b.String()
	}
	return strings.Join(lines, "\n")
}

// restoreBackground re-applies bgSeq after every SGR sequence in line that
// resets the background color.
func restoreBackground(line string, bgSeq string) string {
	if !strings.Contains(line, "\x1b[") {
		return line
	}

	var b strings.Builder
	for {
		start := strings.Index(line, "\x1b[")
		if start == -1 {
			break
		}
		end := start + 2
		for end < len(line) && (line[end] == ';' || line[end] == ':' ||
			(line[end] >= '0' && line[end] <= '9')) {
			end++
		}
		if end >= len(line) {
			break
		}

		b.WriteString(line[:end+1])
		if line[end] == 'm' && resetsBackground(line[start+2:end]) {
			b.WriteString(bgSeq)
		}
		line = line[end+1:]
	}
	b.WriteString(line)
	return b.String()
}

// resetsBackground reports whether the parameters of an SGR sequence leave
// the background color at the terminal's default.
func resetsBackground(params string) bool {
	if params == "" {
		return true
	}

	resets := false
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "", "0", "49":
			resets = true
		case "48":
			// an explicit background color wins over an earlier reset
			resets = false
			i += extendedColorParams(parts[i+1:])
		case "38", "58":
			i += extendedColorParams(parts[i+1:])
		default:
			if isBackgroundColor(parts[i]) {
				resets = false
			}
		}
	}
	return resets
}

// extendedColorParams returns how many parameters following a 38/48/58 belong
// to that color, e.g. "5;n" or "2;r;g;b".
func extendedColorParams(rest []string) int {
	if len(rest) == 0 {
		return 0
	}
	switch rest[0] {
	case "5":
		return min(2, len(rest))
	case "2":
		return min(4, len(rest))
	}
	return 0
}

func isBackgroundColor(param string) bool {
	if len(param) < 2 || len(param) > 3 {
		return false
	}
	n := 0
	for _, c := range param {
		if c < '0' || c > '9' {
			return false
		}
		n = n*10 + int(c-'0')
	}
	return (n >= 40 && n <= 47) || (n >= 100 && n <= 107)
}

// FitWidth cuts or pads every line of s so it is exactly width cells wide.
// Unlike setting a width on a lipgloss style it never wraps lines, and it
// leaves whitespace that carries a background color as it is.
func FitWidth(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		w := lipgloss.Width(line)
		switch {
		case w > width:
			lines[i] = ansi.Truncate(line, width, "")
		case w < width:
			lines[i] = line + strings.Repeat(" ", width-w)
		}
	}
	return strings.Join(lines, "\n")
}
