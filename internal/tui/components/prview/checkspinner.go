package prview

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// ChecksSpinnerTickMsg keeps check animations separate from loading/task
// spinners. Its command only schedules a frame, never a GitHub request.
type ChecksSpinnerTickMsg struct {
	Tick spinner.TickMsg
}

func wrapChecksTick(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		return ChecksSpinnerTickMsg{Tick: cmd().(spinner.TickMsg)}
	}
}

// SyncChecksSpinner starts one animation loop while pending checks are shown.
// Stopping replaces the spinner so any outstanding frames become stale.
func (m *Model) SyncChecksSpinner(previewVisible bool) tea.Cmd {
	shouldAnimate := previewVisible && m.hasData() && m.pr.Data.Primary != nil &&
		m.pr.Data.Primary.State == "OPEN" &&
		(m.SelectedTab() == tabs[0] || m.SelectedTab() == tabs[3]) &&
		len(m.inProgressCheckNames()) > 0
	if !shouldAnimate {
		if m.checksAnimating {
			m.checksAnimating = false
			m.checkSpinner = spinner.New(spinner.WithSpinner(spinner.MiniDot))
		}
		return nil
	}
	if m.checksAnimating {
		return nil
	}
	m.checksAnimating = true
	return wrapChecksTick(m.checkSpinner.Tick)
}

// TickChecksSpinner advances a valid animation frame and schedules the next.
func (m *Model) TickChecksSpinner(msg ChecksSpinnerTickMsg) tea.Cmd {
	if !m.checksAnimating {
		return nil
	}
	var cmd tea.Cmd
	m.checkSpinner, cmd = m.checkSpinner.Update(msg.Tick)
	return wrapChecksTick(cmd)
}
