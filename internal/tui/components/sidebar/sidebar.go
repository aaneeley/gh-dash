package sidebar

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

type Model struct {
	IsOpen     bool
	data       string
	viewport   viewport.Model
	ctx        *context.ProgramContext
	emptyState string
}

func NewModel() Model {
	vp := viewport.New(
		viewport.WithWidth(0),
		viewport.WithHeight(0),
	)

	return Model{
		IsOpen:     false,
		data:       "",
		viewport:   vp,
		ctx:        nil,
		emptyState: "Nothing selected...",
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.PageDown):
			m.viewport.HalfPageDown()

		case key.Matches(msg, keys.Keys.PageUp):
			m.viewport.HalfPageUp()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.IsOpen {
		return ""
	}

	// The sidebar is a panel that is raised from the rest of the screen by
	// its background. A gap of BorderWidth separates it from the main content.
	height := m.ctx.MainContentHeight
	width := m.ctx.DynamicPreviewWidth - m.ctx.Styles.Sidebar.BorderWidth
	isBottom := m.ctx.PreviewPosition == "bottom"
	if isBottom {
		height = m.ctx.DynamicPreviewHeight
		width = m.ctx.DynamicPreviewWidth
	}
	width = max(0, width)

	var content string
	if m.data == "" {
		content = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			m.ctx.Styles.Common.FaintTextStyle.Render(m.emptyState))
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Top,
			m.viewport.View(),
			m.ctx.Styles.Sidebar.PagerStyle.
				Render(fmt.Sprintf("%d%%", int(m.viewport.ScrollPercent()*100))),
		)
	}

	// The background is painted last: passing painted content through a
	// lipgloss style with a width would strip parts of it again.
	panel := common.PaintBackground(
		lipgloss.NewStyle().Height(height).MaxHeight(height).Render(content),
		m.ctx.Theme.PanelBackground,
		width,
	)

	gap := strings.Repeat(" ", m.ctx.Styles.Sidebar.BorderWidth)
	if isBottom {
		return strings.Repeat("\n", m.ctx.Styles.Sidebar.BorderWidth) + panel
	}
	lines := strings.Split(panel, "\n")
	for i := range lines {
		lines[i] = gap + lines[i]
	}
	return strings.Join(lines, "\n")
}

func (m *Model) SetContent(data string) {
	m.data = data
	m.viewport.SetContent(data)
}

func (m *Model) GetSidebarContentWidth() int {
	if m.ctx == nil || m.ctx.Config == nil {
		return 0
	}
	if m.ctx.PreviewPosition == "bottom" {
		return max(0, m.ctx.DynamicPreviewWidth)
	}
	return max(0, m.ctx.DynamicPreviewWidth-m.ctx.Styles.Sidebar.BorderWidth)
}

func (m *Model) ScrollToTop() {
	m.viewport.GotoTop()
}

func (m *Model) ScrollToBottom() {
	m.viewport.GotoBottom()
}

func (m *Model) YOffset() int {
	return m.viewport.YOffset()
}

func (m *Model) SetYOffset(offset int) {
	m.viewport.SetYOffset(max(0, offset))
}

func (m *Model) ScrollToPercent(percent float64) {
	totalLines := m.viewport.TotalLineCount()
	targetLine := int(float64(totalLines) * percent)
	m.viewport.SetYOffset(targetLine)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
	if m.ctx.PreviewPosition == "bottom" {
		m.viewport.SetHeight(m.ctx.DynamicPreviewHeight - m.ctx.Styles.Sidebar.PagerHeight)
	} else {
		m.viewport.SetHeight(m.ctx.MainContentHeight - m.ctx.Styles.Sidebar.PagerHeight)
	}
	m.viewport.SetWidth(m.GetSidebarContentWidth())
}
