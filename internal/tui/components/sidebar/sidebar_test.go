package sidebar

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestSidebar(t *testing.T, position string) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config:               &cfg,
		Theme:                thm,
		Styles:               context.InitStyles(thm),
		PreviewPosition:      position,
		MainContentHeight:    12,
		DynamicPreviewWidth:  40,
		DynamicPreviewHeight: 8,
	}

	m := NewModel()
	m.IsOpen = true
	m.UpdateProgramContext(ctx)
	return m
}

// The sidebar has to fill exactly the space the layout reserves for it,
// otherwise it pushes the rest of the screen around.
func TestView_FillsItsSpace(t *testing.T) {
	content := "a line that is much wider than the sidebar and has to be cut off\nshort"

	tests := []struct {
		position string
		height   int
	}{
		{position: "right", height: 12},
		// the gap above the sidebar takes up one more row
		{position: "bottom", height: 9},
	}
	for _, tt := range tests {
		for name, data := range map[string]string{"empty": "", "with content": content} {
			t.Run(tt.position+" "+name, func(t *testing.T) {
				m := newTestSidebar(t, tt.position)
				m.SetContent(data)

				lines := strings.Split(m.View(), "\n")
				require.Len(t, lines, tt.height)
				for i, line := range lines {
					if tt.position == "bottom" && i == 0 {
						require.Empty(t, line)
						continue
					}
					require.Equal(t, 40, lipgloss.Width(line), "line %d", i)
				}
			})
		}
	}
}

func TestView_ClosedSidebarRendersNothing(t *testing.T) {
	m := newTestSidebar(t, "right")
	m.IsOpen = false
	require.Empty(t, m.View())
}
