package prssection

import (
	"fmt"
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "pr"

type Model struct {
	section.BaseModel
	Prs []prrow.Data
	cfg config.PrsSectionConfig
	// fetchSeq is bumped whenever the section's rows are (re)fetched or
	// replaced, so that results of background refreshes that were started
	// before that point can be recognized as stale and dropped.
	fetchSeq int
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.PrsSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
) Model {
	m := Model{}
	m.cfg = cfg
	m.Prs = []prrow.Data{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)

	return m
}

func (m *Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd
	var err error

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return m, blinkCmd

			case "enter":
				m.SearchValue = m.SearchBar.Value()
				m.SyncSmartFilterWithSearchValue()
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.PromptConfirmationBox.Reset()
				cmd = m.SetIsPromptConfirmationShown(false)
				return m, cmd

			case "enter":
				input := m.PromptConfirmationBox.Value()
				action := m.GetPromptConfirmationAction()
				pr := m.GetCurrRow()
				sid := tasks.SectionIdentifier{Id: m.Id, Type: SectionType}
				if input == "Y" || input == "y" {
					switch action {
					case "close":
						cmd = tasks.ClosePR(m.Ctx, sid, pr)
					case "reopen":
						cmd = tasks.ReopenPR(m.Ctx, sid, pr)
					case "ready":
						cmd = tasks.PRReady(m.Ctx, sid, pr)
					case "merge":
						cmd = tasks.MergePR(m.Ctx, sid, pr)
					case "update":
						cmd = tasks.UpdatePR(m.Ctx, sid, pr)
					case "approveWorkflows":
						cmd = tasks.ApproveWorkflows(m.Ctx, sid, pr)
					}
				}

				m.PromptConfirmationBox.Reset()
				blinkCmd := m.SetIsPromptConfirmationShown(false)

				return m, tea.Batch(cmd, blinkCmd)
			}

			break
		}

		switch {
		case key.Matches(msg, keys.PRKeys.Diff):
			cmd = m.diff()

		case key.Matches(msg, keys.PRKeys.ToggleSmartFiltering):
			before := m.IsFilteredByCurrentRemote

			// If we're filtering by the current repo - we want to remove it
			// If there's no repo filter we want to add the current repo filter.
			if m.HasCurrentRepoNameInConfiguredFilter() || !m.HasRepoNameInConfiguredFilter() {
				m.IsFilteredByCurrentRemote = !before
			}
			log.Debug(
				"toggled smart filtering",
				"before",
				before,
				"after",
				m.IsFilteredByCurrentRemote,
			)
			searchValue := m.GetSearchValue()
			if m.SearchValue != searchValue {
				m.SearchValue = searchValue
				m.SearchBar.SetValue(searchValue)
				m.SetIsSearching(false)
				m.ResetRows()
				return m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

		case key.Matches(msg, keys.PRKeys.Checkout):
			cmd, err = m.checkout()
			if err != nil {
				m.Ctx.Error = err
			}

		case key.Matches(msg, keys.PRKeys.WatchChecks):
			cmd = m.watchChecks()
		}

	case tasks.UpdatePRMsg:
		for i, currPr := range m.Prs {
			if currPr.Primary.Number != msg.PrNumber {
				continue
			}

			if msg.IsClosed != nil {
				if *msg.IsClosed {
					currPr.Primary.State = "CLOSED"
				} else {
					currPr.Primary.State = "OPEN"
				}
			}
			if msg.NewComment != nil {
				currPr.Enriched.Comments.Nodes = append(
					currPr.Enriched.Comments.Nodes, *msg.NewComment)
			}
			if msg.AddedAssignees != nil {
				currPr.Primary.Assignees.Nodes = addAssignees(
					currPr.Primary.Assignees.Nodes, msg.AddedAssignees.Nodes)
			}
			if msg.RemovedAssignees != nil {
				currPr.Primary.Assignees.Nodes = removeAssignees(
					currPr.Primary.Assignees.Nodes, msg.RemovedAssignees.Nodes)
			}
			if msg.Labels != nil {
				currPr.Primary.Labels.Nodes = msg.Labels.Nodes
			}
			if msg.ReadyForReview != nil && *msg.ReadyForReview {
				currPr.Primary.IsDraft = false
			}
			if msg.IsMerged != nil && *msg.IsMerged {
				currPr.Primary.State = "MERGED"
				currPr.Primary.Mergeable = ""
			}
			m.Prs[i] = currPr
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			break
		}

	case BackgroundRefreshMsg:
		m.applyBackgroundRefresh(msg)

	case SectionPullRequestsFetchedMsg:
		if m.LastFetch.TaskId == msg.TaskId {
			m.fetchSeq++
			if m.PageInfo != nil {
				m.Prs = append(m.Prs, msg.Prs...)
			} else {
				m.Prs = msg.Prs
			}
			m.TotalCount = msg.TotalCount
			m.PageInfo = &msg.PageInfo
			m.SetIsLoading(false)
			m.Table.SetRows(m.BuildRows())
			m.Table.SetLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

// EnrichPR stores the result of a single PR fetch on the matching row. The
// row's list data (state, checks, review status, etc.) is updated as well so
// the table reflects the fresh data without refetching the whole section.
func (m *Model) EnrichPR(data data.EnrichedPullRequestData) {
	found := false
	for i, currPr := range m.Prs {
		if currPr.Primary == nil || currPr.Primary.Url != data.Url {
			continue
		}

		found = true
		primary := data.ToListPullRequestData()
		m.Prs[i].Primary = &primary
		m.Prs[i].IsEnriched = true
		m.Prs[i].Enriched = data
		m.Prs[i].RefreshedAt = time.Now()
	}

	if found {
		m.Table.SetRows(m.BuildRows())
	}
}

// BackgroundRefreshMsg carries the result of a refresh that was done without
// putting the section into its loading state.
type BackgroundRefreshMsg struct {
	SectionId  int
	Seq        int
	Prs        []data.PullRequestData
	TotalCount int
	PageInfo   data.PageInfo
	Err        error
}

// maxSearchPageSize is the maximum number of results GitHub's search returns
// in a single page.
const maxSearchPageSize = 100

// RefreshInBackground refetches the rows that are currently loaded without
// showing the loading spinner or resetting the table. The selection is
// restored once the results arrive, see applyBackgroundRefresh.
func (m *Model) RefreshInBackground() tea.Cmd {
	if m == nil || m.IsLoading || m.IsSearching {
		return nil
	}

	limit := m.Ctx.Config.Defaults.PrsLimit
	if m.Config.Limit != nil {
		limit = *m.Config.Limit
	}
	// refetch everything that was already paged in, so the list doesn't shrink
	target := max(limit, len(m.Prs))
	if target <= 0 {
		return nil
	}

	m.fetchSeq++
	seq := m.fetchSeq
	id := m.Id
	filters := m.GetFilters()

	return func() tea.Msg {
		msg := BackgroundRefreshMsg{SectionId: id, Seq: seq}
		var pageInfo *data.PageInfo
		for len(msg.Prs) < target {
			pageSize := min(target-len(msg.Prs), maxSearchPageSize)
			res, err := data.FetchPullRequests(filters, pageSize, pageInfo)
			if err != nil {
				msg.Err = err
				return msg
			}

			msg.Prs = append(msg.Prs, res.Prs...)
			msg.TotalCount = res.TotalCount
			msg.PageInfo = res.PageInfo
			pageInfo = &res.PageInfo
			if !res.PageInfo.HasNextPage || len(res.Prs) == 0 {
				break
			}
		}
		return msg
	}
}

// applyBackgroundRefresh swaps in the refreshed rows while keeping the
// currently selected PR selected. If that PR is no longer part of the results
// the selection stays at the same position in the list.
func (m *Model) applyBackgroundRefresh(msg BackgroundRefreshMsg) {
	if msg.SectionId != m.Id || msg.Seq != m.fetchSeq {
		// a newer fetch was started (or finished) in the meantime
		return
	}
	// make sure the same result is never applied twice
	m.fetchSeq++

	if msg.Err != nil {
		log.Error("failed refreshing PRs in the background", "section", m.Id, "err", msg.Err)
		return
	}

	currIdx := m.Table.GetCurrItem()
	selectedUrl := ""
	if currIdx >= 0 && currIdx < len(m.Prs) && m.Prs[currIdx].Primary != nil {
		selectedUrl = m.Prs[currIdx].Primary.Url
	}

	// keep the data that was already fetched for the sidebar, otherwise it
	// would flash back to its loading state on every refresh
	enriched := make(map[string]prrow.Data, len(m.Prs))
	for _, pr := range m.Prs {
		if pr.IsEnriched && pr.Primary != nil {
			enriched[pr.Primary.Url] = pr
		}
	}

	prs := make([]prrow.Data, 0, len(msg.Prs))
	newIdx := -1
	for i := range msg.Prs {
		pr := msg.Prs[i]
		row := prrow.Data{Primary: &pr}
		if prev, ok := enriched[pr.Url]; ok {
			row.Enriched = prev.Enriched
			row.IsEnriched = true
			row.RefreshedAt = prev.RefreshedAt
		}
		if newIdx == -1 && selectedUrl != "" && pr.Url == selectedUrl {
			newIdx = i
		}
		prs = append(prs, row)
	}
	if newIdx == -1 {
		newIdx = currIdx
	}

	m.Prs = prs
	m.TotalCount = msg.TotalCount
	m.PageInfo = &msg.PageInfo
	m.LastFetch.At = time.Now()
	m.Table.SetRows(m.BuildRows())
	m.Table.SetCurrItem(newIdx)
	m.Table.SetLastUpdated(time.Now())
	m.UpdateTotalItemsCount(m.TotalCount)
}

// GetCurrRowUrl returns the URL of the selected PR, or an empty string when
// nothing is selected.
func (m *Model) GetCurrRowUrl() string {
	idx := m.Table.GetCurrItem()
	if idx < 0 || idx >= len(m.Prs) || m.Prs[idx].Primary == nil {
		return ""
	}
	return m.Prs[idx].Primary.Url
}

// RefreshAllSectionsInBackground refreshes all the given sections without
// showing their loading state.
func RefreshAllSectionsInBackground(sections []section.Section) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(sections))
	for _, prSection := range section.ToImplSections[*Model](sections) {
		cmds = append(cmds, prSection.RefreshInBackground())
	}
	return tea.Batch(cmds...)
}

func GetSectionColumns(
	cfg config.PrsSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Prs
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(
		dLayout.UpdatedAt,
		sLayout.UpdatedAt,
	)
	createdAtLayout := config.MergeColumnConfigs(
		dLayout.CreatedAt,
		sLayout.CreatedAt,
	)
	repoLayout := config.MergeColumnConfigs(dLayout.Repo, sLayout.Repo)
	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	authorLayout := config.MergeColumnConfigs(dLayout.Author, sLayout.Author)
	assigneesLayout := config.MergeColumnConfigs(
		dLayout.Assignees,
		sLayout.Assignees,
	)
	baseLayout := config.MergeColumnConfigs(dLayout.Base, sLayout.Base)
	numCommentsLayout := config.MergeColumnConfigs(
		dLayout.NumComments,
		sLayout.NumComments,
	)
	reviewStatusLayout := config.MergeColumnConfigs(
		dLayout.ReviewStatus,
		sLayout.ReviewStatus,
	)
	stateLayout := config.MergeColumnConfigs(dLayout.State, sLayout.State)
	ciLayout := config.MergeColumnConfigs(dLayout.Ci, sLayout.Ci)
	labelsLayout := config.MergeColumnConfigs(dLayout.Labels, sLayout.Labels)
	linesLayout := config.MergeColumnConfigs(dLayout.Lines, sLayout.Lines)

	if !ctx.Config.Theme.Ui.Table.Compact {
		return []table.Column{
			{
				Title:  "",
				Width:  utils.IntPtr(3),
				Hidden: stateLayout.Hidden,
			},
			{
				Title:  "Title",
				Grow:   utils.BoolPtr(true),
				Hidden: titleLayout.Hidden,
			},
			{
				Title:  constants.LabelsIcon,
				Width:  labelsLayout.Width,
				Hidden: labelsLayout.Hidden,
			},
			{
				Title:  "Assignees",
				Width:  assigneesLayout.Width,
				Hidden: assigneesLayout.Hidden,
			},
			{
				Title:  "Base",
				Width:  baseLayout.Width,
				Hidden: baseLayout.Hidden,
			},
			{
				Title:  constants.CommentsIcon,
				Width:  utils.IntPtr(4),
				Hidden: numCommentsLayout.Hidden,
			},
			{
				Title:  "󰯢",
				Width:  utils.IntPtr(4),
				Hidden: reviewStatusLayout.Hidden,
			},
			{
				Title:  "",
				Width:  &ctx.Styles.PrSection.CiCellWidth,
				Grow:   new(bool),
				Hidden: ciLayout.Hidden,
			},
			{
				Title:  "",
				Width:  linesLayout.Width,
				Hidden: linesLayout.Hidden,
			},
			{
				Title:  "󱦻",
				Width:  updatedAtLayout.Width,
				Hidden: updatedAtLayout.Hidden,
			},
			{
				Title:  "󱡢",
				Width:  createdAtLayout.Width,
				Hidden: createdAtLayout.Hidden,
			},
		}
	}

	return []table.Column{
		{
			Title:  "",
			Width:  utils.IntPtr(3),
			Hidden: stateLayout.Hidden,
		},
		{
			Title:  "",
			Width:  repoLayout.Width,
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  "Author",
			Width:  authorLayout.Width,
			Hidden: authorLayout.Hidden,
		},
		{
			Title:  constants.LabelsIcon,
			Width:  labelsLayout.Width,
			Hidden: labelsLayout.Hidden,
		},
		{
			Title:  "Assignees",
			Width:  assigneesLayout.Width,
			Hidden: assigneesLayout.Hidden,
		},
		{
			Title:  "Base",
			Width:  baseLayout.Width,
			Hidden: baseLayout.Hidden,
		},
		{
			Title:  constants.CommentsIcon,
			Width:  utils.IntPtr(4),
			Hidden: numCommentsLayout.Hidden,
		},
		{
			Title:  "󰯢",
			Width:  utils.IntPtr(4),
			Hidden: reviewStatusLayout.Hidden,
		},
		{
			Title:  "",
			Width:  &ctx.Styles.PrSection.CiCellWidth,
			Grow:   new(bool),
			Hidden: ciLayout.Hidden,
		},
		{
			Title:  "",
			Width:  linesLayout.Width,
			Hidden: linesLayout.Hidden,
		},
		{
			Title:  "󱦻",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "󱡢",
			Width:  createdAtLayout.Width,
			Hidden: createdAtLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	currItem := m.Table.GetCurrItem()
	for i, currPr := range m.Prs {
		prModel := prrow.PullRequest{
			Ctx:     m.Ctx,
			Data:    &currPr,
			Columns: m.Table.Columns, ShowAuthorIcon: m.ShowAuthorIcon,
			// every row is from the current repo, no need to repeat its name
			HideRepoName: m.IsFilteredByCurrentRemote,
		}
		rows = append(
			rows,
			prModel.ToTableRow(currItem == i),
		)
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Prs)
}

type SectionPullRequestsFetchedMsg struct {
	Prs        []prrow.Data
	TotalCount int
	PageInfo   data.PageInfo
	TaskId     string
}

func (m *Model) GetCurrRow() data.RowData {
	idx := m.Table.GetCurrItem()
	if idx < 0 || idx >= len(m.Prs) {
		return nil
	}
	pr := m.Prs[idx]
	return &pr
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	// this isn't the first page and there isn't a next page
	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd
	m.fetchSeq++

	startCursor := time.Now().String()
	if m.PageInfo != nil {
		startCursor = m.PageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_prs_%d_%s", m.Id, startCursor)
	m.LastFetch.TaskId = taskId
	m.LastFetch.At = time.Now()
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching PRs for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`PRs for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}

		res, err := data.FetchPullRequests(m.GetFilters(), *limit, m.PageInfo)
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		prs := make([]prrow.Data, 0)
		for _, pr := range res.Prs {
			prs = append(prs, prrow.Data{Primary: &pr})
		}
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionPullRequestsFetchedMsg{
				Prs:        prs,
				TotalCount: res.TotalCount,
				PageInfo:   res.PageInfo,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	cmds = append(cmds, m.SetIsLoading(true))

	return cmds
}

func (m *Model) ResetRows() {
	m.fetchSeq++
	m.Prs = nil
	m.BaseModel.ResetRows()
}

func InitSections(ctx *context.ProgramContext) []section.Section {
	sections := make([]section.Section, len(ctx.Config.PRSections))
	for i, sectionConfig := range ctx.Config.PRSections {
		sections[i] = new(NewModel(
			i,
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		))
	}

	return sections
}

func FetchAllSections(
	ctx *context.ProgramContext,
	sections []section.Section,
) tea.Cmd {
	fetchPRsCmds := make([]tea.Cmd, 0)
	prSections := section.ToImplSections[*Model](sections)
	for _, prSection := range prSections {
		prSection.SetLastUpdated(time.Now())
		prSection.SetCreatedAt(time.Now())
		prSection.PageInfo = nil
		fetchPRsCmds = append(fetchPRsCmds, prSection.FetchNextPageSectionRows()...)
	}

	return tea.Batch(fetchPRsCmds...)
}

func addAssignees(assignees, addedAssignees []data.Assignee) []data.Assignee {
	newAssignees := assignees
	for _, assignee := range addedAssignees {
		if !assigneesContains(newAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func removeAssignees(
	assignees, removedAssignees []data.Assignee,
) []data.Assignee {
	newAssignees := []data.Assignee{}
	for _, assignee := range assignees {
		if !assigneesContains(removedAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func assigneesContains(assignees []data.Assignee, assignee data.Assignee) bool {
	return slices.Contains(assignees, assignee)
}

func (m Model) GetItemSingularForm() string {
	return "PR"
}

func (m Model) GetItemPluralForm() string {
	return "PRs"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.currRowLastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v (fetched %v)",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

// currRowLastUpdated returns when the data of the selected PR was last
// fetched: either with the rest of the section or, if that happened more
// recently, on its own.
func (m Model) currRowLastUpdated() time.Time {
	lastUpdated := m.LastUpdated()
	idx := m.Table.GetCurrItem()
	if idx >= 0 && idx < len(m.Prs) && m.Prs[idx].RefreshedAt.After(lastUpdated) {
		return m.Prs[idx].RefreshedAt
	}
	return lastUpdated
}

func (m *Model) ConfigGetter() config.SectionConfig {
	return m.cfg.ToSectionConfig()
}
