package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	gh "github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/branchsidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/footer"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issueview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/sidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tabs"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func testPrUrl(n int) string {
	return fmt.Sprintf("https://github.com/dlvhdr/gh-dash/pull/%d", n)
}

// mockGitHub serves canned GraphQL responses: searches return the PRs in
// searchResult and single PR fetches return a PR in the given state.
type mockGitHub struct {
	searchResult []int
	prState      string
	requests     []string
}

func (g *mockGitHub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	w.Header().Set("Content-Type", "application/json")

	var res map[string]any
	switch {
	case strings.Contains(string(body), "query SearchPullRequests"):
		g.requests = append(g.requests, "search")
		nodes := make([]map[string]any, 0, len(g.searchResult))
		for _, n := range g.searchResult {
			nodes = append(nodes, map[string]any{
				"number": n,
				"title":  fmt.Sprintf("PR %d", n),
				"url":    testPrUrl(n),
				"state":  "OPEN",
			})
		}
		res = map[string]any{"data": map[string]any{"search": map[string]any{
			"nodes":      nodes,
			"issueCount": len(nodes),
			"pageInfo":   map[string]any{"hasNextPage": false},
		}}}
	case strings.Contains(string(body), "query FetchPullRequest"):
		g.requests = append(g.requests, "pr")
		var parsed struct {
			Variables struct {
				Url string
			}
		}
		_ = json.Unmarshal(body, &parsed)
		res = map[string]any{"data": map[string]any{"resource": map[string]any{
			"url":            parsed.Variables.Url,
			"title":          "refreshed title",
			"state":          g.prState,
			"reviewDecision": "APPROVED",
			"comments":       map[string]any{"totalCount": 7},
			"commits": map[string]any{"nodes": []map[string]any{{
				"commit": map[string]any{
					"statusCheckRollup": map[string]any{"state": "FAILURE"},
				},
			}}},
		}}}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(res)
}

func setMockGitHub(t *testing.T, g *mockGitHub) {
	t.Helper()
	client, err := gh.NewGraphQLClient(gh.ClientOptions{
		Transport: localRoundTripper{handler: g},
		Host:      "localhost:3000",
		AuthToken: "fake-token",
	})
	require.NoError(t, err)
	data.SetClient(client)
	t.Cleanup(func() { data.SetClient(nil) })
}

// newBackgroundRefreshTestModel creates a model showing a single PRs section
// that has the given PRs loaded, with the preview open.
func newBackgroundRefreshTestModel(t *testing.T, numbers ...int) (*Model, *prssection.Model) {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config:       &cfg,
		ScreenWidth:  200,
		ScreenHeight: 60,
		View:         config.PRsView,
		StartTask:    func(task context.Task) tea.Cmd { return nil },
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	prSection := prssection.NewModel(0, ctx,
		config.PrsSectionConfig{Title: "Test", Filters: "is:open"},
		time.Now(), time.Now())
	for _, n := range numbers {
		prSection.Prs = append(prSection.Prs, prrow.Data{Primary: &data.PullRequestData{
			Number: n,
			Title:  fmt.Sprintf("PR %d", n),
			Url:    testPrUrl(n),
			State:  "OPEN",
		}})
	}

	m := &Model{
		ctx:              ctx,
		keys:             keys.Keys,
		prs:              []section.Section{&prSection},
		sidebar:          sidebar.NewModel(),
		footer:           footer.NewModel(ctx),
		tabs:             tabs.NewModel(ctx),
		prView:           prview.NewModel(ctx),
		issueSidebar:     issueview.NewModel(ctx),
		branchSidebar:    branchsidebar.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
		tasks:            map[string]context.Task{},
	}
	m.sidebar.IsOpen = true
	m.syncMainContentDimensions()
	m.syncProgramContext()
	prSection.Table.SetRows(prSection.BuildRows())
	m.updateTabs()
	m.syncSidebar()

	return m, &prSection
}

// runBackgroundRefresh starts a background refresh of the section and feeds
// its result back into the model, like the program's event loop would.
func runBackgroundRefresh(t *testing.T, m *Model, s *prssection.Model) {
	t.Helper()
	cmd := s.RefreshInBackground()
	require.NotNil(t, cmd)
	msg := cmd()
	require.IsType(t, prssection.BackgroundRefreshMsg{}, msg)
	m.Update(msg)
}

func TestIntervalRefresh_DoesNotResetThePRsView(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	s.NextRow()

	m.Update(intervalRefresh(time.Now()))

	require.False(t, s.GetIsLoading(), "the section should not show its loading state")
	require.False(t, s.Table.IsLoading(), "the table should not be replaced by a spinner")
	require.Len(t, s.Prs, 3)
	require.Equal(t, testPrUrl(2), s.GetCurrRowUrl())
}

func TestBackgroundRefresh_KeepsViewedPRAndSidebarState(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	setMockGitHub(t, &mockGitHub{searchResult: []int{4, 3, 1, 2}})

	s.NextRow()
	s.NextRow()
	require.Equal(t, testPrUrl(3), s.GetCurrRowUrl())
	m.Update(tea.KeyPressMsg{Text: "]"})
	selectedTab := m.prView.SelectedTab()

	runBackgroundRefresh(t, m, s)

	require.Len(t, s.Prs, 4)
	require.Equal(t, testPrUrl(3), s.GetCurrRowUrl(), "the viewed PR should stay selected")
	require.Equal(t, 1, s.CurrRow())
	require.Equal(t, selectedTab, m.prView.SelectedTab(),
		"the sidebar should stay on the same tab")
	require.False(t, s.Table.IsLoading())
}

func TestBackgroundRefresh_ResetsSidebarWhenViewedPRDropsOff(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	setMockGitHub(t, &mockGitHub{searchResult: []int{1, 3}})

	s.NextRow()
	require.Equal(t, testPrUrl(2), s.GetCurrRowUrl())
	initialTab := m.prView.SelectedTab()
	m.Update(tea.KeyPressMsg{Text: "]"})
	require.NotEqual(t, initialTab, m.prView.SelectedTab())

	runBackgroundRefresh(t, m, s)

	require.Len(t, s.Prs, 2)
	require.Equal(t, 1, s.CurrRow(), "the selection should stay at the same position")
	require.Equal(t, testPrUrl(3), s.GetCurrRowUrl())
	require.Equal(t, initialTab, m.prView.SelectedTab(),
		"a different PR is viewed now, so the sidebar starts over")
}

func TestRefreshSelectedPr_OnlyFetchesAndUpdatesTheViewedPR(t *testing.T) {
	m, s := newBackgroundRefreshTestModel(t, 1, 2, 3)
	github := &mockGitHub{prState: "MERGED"}
	setMockGitHub(t, github)

	s.NextRow()
	m.Update(tea.KeyPressMsg{Text: "]"})
	selectedTab := m.prView.SelectedTab()

	cmd := m.refreshSelectedPr()
	require.NotNil(t, cmd)
	msg := cmd()
	require.IsType(t, prview.EnrichedPrMsg{}, msg)
	m.Update(msg)

	require.Equal(t, []string{"pr"}, github.requests, "only the viewed PR should be fetched")
	require.Equal(t, testPrUrl(2), s.GetCurrRowUrl())
	require.Equal(t, selectedTab, m.prView.SelectedTab())
	require.False(t, s.Table.IsLoading())

	refreshed := s.Prs[1]
	require.True(t, refreshed.IsEnriched)
	require.Equal(t, "refreshed title", refreshed.Primary.Title)
	require.Equal(t, "MERGED", refreshed.Primary.State)
	require.Equal(t, "APPROVED", refreshed.Primary.ReviewDecision)
	require.Equal(t, 7, refreshed.Primary.Comments.TotalCount)
	require.Len(t, refreshed.Primary.Commits.Nodes, 1)
	require.EqualValues(t, "FAILURE",
		refreshed.Primary.Commits.Nodes[0].Commit.StatusCheckRollup.State)

	for _, i := range []int{0, 2} {
		require.False(t, s.Prs[i].IsEnriched)
		require.Equal(t, "OPEN", s.Prs[i].Primary.State)
	}
}

func TestRefreshSelectedPr_NoOpOutsideOfThePRsView(t *testing.T) {
	m, _ := newBackgroundRefreshTestModel(t, 1, 2, 3)
	m.ctx.View = config.IssuesView

	require.Nil(t, m.refreshSelectedPr())
}

func TestDoRefreshSelectedPrAtInterval_CanBeDisabled(t *testing.T) {
	m, _ := newBackgroundRefreshTestModel(t, 1)
	require.Equal(t, 15, m.ctx.Config.Defaults.SelectedPrRefetchIntervalSeconds)
	require.NotNil(t, m.doRefreshSelectedPrAtInterval())

	m.ctx.Config.Defaults.SelectedPrRefetchIntervalSeconds = 0
	require.Nil(t, m.doRefreshSelectedPrAtInterval())
}
