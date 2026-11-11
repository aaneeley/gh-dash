package tui

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/recovery"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
)

func attachTestRecovery(t *testing.T, m *Model, state json.RawMessage) (*recovery.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	store, err := recovery.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	require.NoError(t, m.EnableRecovery(store, recovery.Checkpoint{Cwd: t.TempDir(), UI: state}))
	return store, path
}

func readTestUI(t *testing.T, path string) (recovery.Checkpoint, uiCheckpoint) {
	t.Helper()
	saved, err := recovery.Load(path)
	require.NoError(t, err)
	var ui uiCheckpoint
	require.NoError(t, json.Unmarshal(saved.UI, &ui))
	return saved, ui
}

func TestRecovery_RestoresSelectedPRAndPreviewAfterReordering(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	_, path := attachTestRecovery(t, m, nil)
	s.NextRow()
	s.EnrichPR(data.EnrichedPullRequestData{
		Url: testPrUrl(2), Number: 2, State: "OPEN", Title: "saved PR",
		Comments: data.CommentsWithBody{Nodes: []data.Comment{{Body: strings.Repeat("a comment line\n", 200)}}},
	})
	m.prView.GoToActivityTab()
	m.prView.SetSummaryViewMore()
	m.syncSidebar()
	m.sidebar.SetYOffset(10)
	require.Equal(t, 10, m.sidebar.YOffset())
	m.writeCheckpoint()
	saved, ui := readTestUI(t, path)
	require.Equal(t, testPrUrl(2), ui.Sections[0].Selected)
	require.Equal(t, 10, ui.PreviewOffset)
	require.True(t, saved.Active)
	require.NotContains(t, string(saved.UI), "a comment line", "do not save GitHub response bodies")

	restored, _ := newBackgroundRefreshTestModel(t)
	attachTestRecovery(t, restored, saved.UI)
	restored.initRecovery()
	prSection := restored.prs[0].(*prssection.Model)
	// API results changed order while the program was stopped.
	for _, n := range []int{3, 1, 2} {
		prSection.Prs = append(prSection.Prs, prrow.Data{Primary: &data.PullRequestData{
			Number: n, Url: testPrUrl(n), State: "OPEN", Title: "fresh PR",
		}})
	}
	prSection.PageInfo = &data.PageInfo{}
	prSection.Table.SetRows(prSection.BuildRows())
	_, handled := restored.restoreSection(prSection)
	require.True(t, handled)
	require.Equal(t, 2, prSection.CurrRow())
	require.Equal(t, testPrUrl(2), prSection.GetCurrRowUrl())

	// Fetching the sidebar must complete before its scroll offset is applied.
	restored.finishRecoveryPreview()
	require.NotNil(t, restored.recovery.preview)
	prSection.EnrichPR(s.Prs[1].Enriched)
	restored.finishRecoveryPreview()
	require.Nil(t, restored.recovery.preview)
	require.Equal(t, m.prView.SelectedTab(), restored.prView.SelectedTab())
	require.True(t, restored.prView.SummaryExpanded())
	require.Equal(t, 10, restored.sidebar.YOffset())
}

func TestRecovery_PaginatesForSavedSelectionAndHandlesMissingPR(t *testing.T) {
	for _, found := range []bool{true, false} {
		t.Run(map[bool]string{true: "later page", false: "removed from list"}[found], func(t *testing.T) {
			m, s := newBackgroundRefreshTestModel(t, 1, 2)
			attachTestRecovery(t, m, nil)
			m.recovery.pending[recoveryKey(s)] = sectionCheckpoint{Selected: testPrUrl(3), Row: 1, Loaded: 3}
			s.PageInfo = &data.PageInfo{HasNextPage: true, EndCursor: "page-1"}
			cmd, handled := m.restoreSection(s)
			require.True(t, handled)
			require.NotNil(t, cmd)
			require.True(t, s.GetIsLoading())
			s.SetIsLoading(false)
			s.PageInfo.HasNextPage = false
			if found {
				s.Prs = append(s.Prs, prrow.Data{Primary: &data.PullRequestData{Number: 3, Url: testPrUrl(3)}})
				s.Table.SetRows(s.BuildRows())
			}
			m.restoreSection(s)
			require.Empty(t, m.recovery.pending)
			if found {
				require.Equal(t, testPrUrl(3), s.GetCurrRowUrl())
			} else {
				require.Equal(t, testPrUrl(2), s.GetCurrRowUrl())
			}
		})
	}
}

func TestRecovery_PreservesCheckpointDuringLoadingAndDisarmsOnQuit(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	_, path := attachTestRecovery(t, m, nil)
	s.NextRow()
	m.writeCheckpoint()
	s.ResetRows()
	s.SetIsLoading(true)
	m.writeCheckpoint()
	_, saved := readTestUI(t, path)
	require.Equal(t, testPrUrl(2), saved.Sections[0].Selected)
	m.ctx.Config.ConfirmQuit = false
	m.Update(tea.KeyPressMsg{Text: "q"})
	checkpoint, _ := readTestUI(t, path)
	require.False(t, checkpoint.Active, "intentional quits must not auto-reopen")
}

func TestRecovery_ClosedPreviewRetainsStateUntilReopened(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1)
	_, path := attachTestRecovery(t, m, nil)
	s.EnrichPR(data.EnrichedPullRequestData{
		Url: testPrUrl(1), Number: 1, State: "OPEN", Title: "saved PR",
		Comments: data.CommentsWithBody{Nodes: []data.Comment{{Body: strings.Repeat("comment\n\n", 200)}}},
	})
	m.sidebar.IsOpen = false
	m.recovery.preview = &uiCheckpoint{PreviewTab: " Activity", PreviewOffset: 10, Expanded: true}
	m.finishRecoveryPreview()
	require.NotNil(t, m.recovery.preview)
	m.writeCheckpoint()
	_, before := readTestUI(t, path)
	require.False(t, before.PreviewOpen)
	require.Equal(t, 10, before.PreviewOffset)

	m.Update(tea.KeyPressMsg{Text: "p"})
	require.True(t, m.sidebar.IsOpen)
	require.Nil(t, m.recovery.preview)
	require.Equal(t, " Activity", m.prView.SelectedTab())
	require.Equal(t, 10, m.sidebar.YOffset())
}
