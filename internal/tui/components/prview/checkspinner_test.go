package prview

import (
	"testing"

	"charm.land/bubbles/v2/spinner"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestChecksSpinner_AnimatesAndStops(t *testing.T) {
	m := newTestModelForChecks(t, checksTestOptions{
		checkRuns: []data.CheckRun{makeCheckRun("running tests", "IN_PROGRESS", "")},
	})
	m.pr.Data.Primary.State = "OPEN"
	before := ansi.Strip(m.viewInProgressChecks())
	start := m.SyncChecksSpinner(true)
	require.NotNil(t, start)
	require.Nil(t, m.SyncChecksSpinner(true), "do not start duplicate timer loops")

	tick := start().(ChecksSpinnerTickMsg)
	next := m.TickChecksSpinner(tick)
	require.NotNil(t, next)
	after := ansi.Strip(m.viewInProgressChecks())
	require.NotEqual(t, before, after, "the spinner should advance while idle")
	require.Contains(t, after, "running tests")

	// Other loading/task spinners must not drive this animation.
	other := spinner.New()
	require.Nil(t, m.TickChecksSpinner(ChecksSpinnerTickMsg{Tick: other.Tick().(spinner.TickMsg)}))
	require.Equal(t, after, ansi.Strip(m.viewInProgressChecks()))

	stale := next().(ChecksSpinnerTickMsg)
	require.Nil(t, m.SyncChecksSpinner(false))
	require.False(t, m.checksAnimating)
	require.Nil(t, m.TickChecksSpinner(stale))

	// Starting again invalidates outstanding frames from the previous loop.
	require.NotNil(t, m.SyncChecksSpinner(true))
	require.Nil(t, m.TickChecksSpinner(stale))
}

func TestChecksSpinner_OnlyRunsForVisiblePendingChecks(t *testing.T) {
	m := newTestModelForChecks(t, checksTestOptions{
		checkRuns: []data.CheckRun{makeCheckRun("running tests", "IN_PROGRESS", "")},
	})
	m.pr.Data.Primary.State = "OPEN"

	for _, tab := range []int{0, 3} {
		m.carousel.SetCursor(tab)
		require.NotNil(t, m.SyncChecksSpinner(true))
		m.SyncChecksSpinner(false)
	}
	for _, tab := range []int{1, 2, 4} {
		m.carousel.SetCursor(tab)
		require.Nil(t, m.SyncChecksSpinner(true))
	}
	m.carousel.SetCursor(0)
	m.pr.Data.Enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes[0].CheckRun.Status = "COMPLETED"
	require.Nil(t, m.SyncChecksSpinner(true), "completed checks must stop animating")
	m.pr.Data.Primary = nil
	require.NotPanics(t, func() { require.Nil(t, m.SyncChecksSpinner(true)) })
}
