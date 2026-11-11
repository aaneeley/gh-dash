package prssection

import (
	"errors"
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func prUrl(n int) string {
	return fmt.Sprintf("https://github.com/dlvhdr/gh-dash/pull/%d", n)
}

func prData(n int) data.PullRequestData {
	pr := data.PullRequestData{Number: n, Url: prUrl(n), Title: fmt.Sprintf("PR %d", n)}
	pr.Repository.NameWithOwner = "dlvhdr/gh-dash"
	return pr
}

func prsData(numbers ...int) []data.PullRequestData {
	prs := make([]data.PullRequestData, 0, len(numbers))
	for _, n := range numbers {
		prs = append(prs, prData(n))
	}
	return prs
}

// newRefreshTestModel creates a section with the given PRs loaded and enough
// room to show all of them.
func newRefreshTestModel(t *testing.T, numbers ...int) *Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config:            &cfg,
		Theme:             thm,
		Styles:            context.InitStyles(thm),
		ScreenWidth:       120,
		ScreenHeight:      60,
		MainContentWidth:  120,
		MainContentHeight: 50,
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}

	m := NewModel(0, ctx, config.PrsSectionConfig{Title: "Test", Filters: "is:open"},
		time.Now(), time.Now())
	m.UpdateProgramContext(ctx)
	for _, pr := range prsData(numbers...) {
		m.Prs = append(m.Prs, prrow.Data{Primary: &pr})
	}
	m.Table.SetRows(m.BuildRows())
	return &m
}

func refreshMsg(m *Model, numbers ...int) BackgroundRefreshMsg {
	return BackgroundRefreshMsg{
		SectionId:  m.Id,
		Seq:        m.fetchSeq,
		Prs:        prsData(numbers...),
		TotalCount: len(numbers),
	}
}

func TestBackgroundRefresh_KeepsSelectedPRWhenItMoves(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2, 3, 4)
	m.NextRow()
	m.NextRow()
	require.Equal(t, prUrl(3), m.GetCurrRowUrl())

	// PR 3 was updated and is now first, a new PR 5 showed up as well
	_, _ = m.Update(refreshMsg(m, 3, 5, 1, 2, 4))

	require.Equal(t, 0, m.CurrRow())
	require.Equal(t, prUrl(3), m.GetCurrRowUrl())
	require.Len(t, m.Prs, 5)
	require.Equal(t, 5, m.TotalCount)
}

func TestBackgroundRefresh_KeepsPositionWhenSelectedPRDropsOff(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2, 3, 4)
	m.NextRow()
	require.Equal(t, prUrl(2), m.GetCurrRowUrl())

	_, _ = m.Update(refreshMsg(m, 1, 3, 4))

	require.Equal(t, 1, m.CurrRow())
	require.Equal(t, prUrl(3), m.GetCurrRowUrl())
}

func TestBackgroundRefresh_ClampsSelectionWhenListShrinks(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2, 3, 4)
	m.LastItem()
	require.Equal(t, prUrl(4), m.GetCurrRowUrl())

	_, _ = m.Update(refreshMsg(m, 1, 2))

	require.Equal(t, 1, m.CurrRow())
	require.Equal(t, prUrl(2), m.GetCurrRowUrl())

	_, _ = m.Update(refreshMsg(m))
	require.Equal(t, 0, m.CurrRow())
	require.Empty(t, m.GetCurrRowUrl())
}

func TestBackgroundRefresh_DoesNotShowLoadingState(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)

	cmd := m.RefreshInBackground()

	require.NotNil(t, cmd)
	require.False(t, m.GetIsLoading())
	require.False(t, m.Table.IsLoading())
	require.Len(t, m.Prs, 2, "rows should stay in place while refreshing")
}

func TestBackgroundRefresh_SkippedWhileLoading(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)
	m.SetIsLoading(true)

	require.Nil(t, m.RefreshInBackground())
}

func TestBackgroundRefresh_KeepsEnrichedData(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)
	m.EnrichPR(data.EnrichedPullRequestData{Url: prUrl(2), Number: 2, Body: "body"})
	require.True(t, m.Prs[1].IsEnriched)

	_, _ = m.Update(refreshMsg(m, 2, 1))

	require.True(t, m.Prs[0].IsEnriched)
	require.Equal(t, "body", m.Prs[0].Enriched.Body)
	require.False(t, m.Prs[1].IsEnriched)
}

func TestBackgroundRefresh_DropsStaleResults(t *testing.T) {
	t.Run("a manual refresh was started in the meantime", func(t *testing.T) {
		m := newRefreshTestModel(t, 1, 2)
		msg := refreshMsg(m, 7, 8, 9)
		m.ResetRows()

		_, _ = m.Update(msg)

		require.Empty(t, m.Prs)
	})

	t.Run("the same result is only applied once", func(t *testing.T) {
		m := newRefreshTestModel(t, 1, 2)
		msg := refreshMsg(m, 1, 2, 3)

		_, _ = m.Update(msg)
		require.Len(t, m.Prs, 3)

		m.Prs = m.Prs[:1]
		_, _ = m.Update(msg)
		require.Len(t, m.Prs, 1)
	})

	t.Run("result for another section", func(t *testing.T) {
		m := newRefreshTestModel(t, 1, 2)
		msg := refreshMsg(m, 7)
		msg.SectionId = m.Id + 1

		_, _ = m.Update(msg)

		require.Len(t, m.Prs, 2)
	})
}

func TestBackgroundRefresh_ErrorKeepsRows(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)
	m.NextRow()
	msg := refreshMsg(m)
	msg.Err = errors.New("boom")

	_, _ = m.Update(msg)

	require.Len(t, m.Prs, 2)
	require.Equal(t, prUrl(2), m.GetCurrRowUrl())
}

func TestEnrichPR_UpdatesListRow(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)
	// a PR with the same number in another repo must not be touched
	other := prData(2)
	other.Url = "https://github.com/other/repo/pull/2"
	m.Prs = append(m.Prs, prrow.Data{Primary: &other})

	enriched := data.EnrichedPullRequestData{
		Url:            prUrl(2),
		Number:         2,
		Title:          "new title",
		State:          "MERGED",
		ReviewDecision: "APPROVED",
		IsInMergeQueue: true,
	}
	enriched.Comments.TotalCount = 3
	enriched.ReviewThreads.TotalCount = 4
	m.EnrichPR(enriched)

	require.Equal(t, "PR 1", m.Prs[0].Primary.Title)
	require.False(t, m.Prs[0].IsEnriched)

	updated := m.Prs[1]
	require.True(t, updated.IsEnriched)
	require.Equal(t, "new title", updated.Primary.Title)
	require.Equal(t, "MERGED", updated.Primary.State)
	require.Equal(t, "APPROVED", updated.Primary.ReviewDecision)
	require.True(t, updated.Primary.IsInMergeQueue)
	require.Equal(t, 3, updated.Primary.Comments.TotalCount)
	require.Equal(t, 4, updated.Primary.ReviewThreads.TotalCount)

	require.Equal(t, "PR 2", m.Prs[2].Primary.Title)
	require.False(t, m.Prs[2].IsEnriched)
}

func TestPagerContent_ReflectsSelectedPRRefresh(t *testing.T) {
	m := newRefreshTestModel(t, 1, 2)
	m.TotalCount = 2
	m.SetLastUpdated(time.Now().Add(-10 * time.Minute))
	require.Contains(t, m.GetPagerContent(), "Updated ~10m ago")

	// the selected PR was just refetched on its own
	m.EnrichPR(data.EnrichedPullRequestData{Url: prUrl(1), Number: 1})
	require.Contains(t, m.GetPagerContent(), "Updated just now")

	// the other PR still has the data from when the section was fetched
	m.NextRow()
	require.Contains(t, m.GetPagerContent(), "Updated ~10m ago")

	// a refresh of the whole section is newer than the PR's own refresh
	m.Prs[0].RefreshedAt = time.Now().Add(-5 * time.Minute)
	m.PrevRow()
	require.Contains(t, m.GetPagerContent(), "Updated ~5m ago")
	_, _ = m.Update(refreshMsg(m, 1, 2))
	require.Contains(t, m.GetPagerContent(), "Updated just now")
}

var _ section.Section = (*Model)(nil)
