package tui

import (
	"encoding/json"
	"fmt"
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/recovery"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
)

type sectionCheckpoint struct {
	View              string `json:"view"`
	Title             string `json:"title"`
	ConfiguredFilters string `json:"configuredFilters"`
	Search            string `json:"search"`
	SmartFiltered     bool   `json:"smartFiltered"`
	Limit             *int   `json:"limit,omitempty"`
	Selected          string `json:"selected,omitempty"`
	Row               int    `json:"row"`
	Loaded            int    `json:"loaded"`
	Offset            int    `json:"offset"`
}

type uiCheckpoint struct {
	View          string              `json:"view"`
	Section       int                 `json:"section"`
	Sections      []sectionCheckpoint `json:"sections"`
	PreviewOpen   bool                `json:"previewOpen"`
	Position      string              `json:"position"`
	PreviewTab    string              `json:"previewTab"`
	Expanded      bool                `json:"expanded"`
	PreviewOffset int                 `json:"previewOffset"`
	HelpOpen      bool                `json:"helpOpen"`
}

type recoveryState struct {
	store    *recovery.Store
	launch   recovery.Checkpoint
	initial  *uiCheckpoint
	pending  map[string]sectionCheckpoint
	last     map[string]sectionCheckpoint
	preview  *uiCheckpoint
	quitting bool
}

// EnableRecovery attaches a pane's checkpoint store. Tests and terminals
// outside Rex have no store and retain the original startup behavior.
func (m *Model) EnableRecovery(store *recovery.Store, launch recovery.Checkpoint) error {
	r := &recoveryState{
		store: store, launch: launch,
		pending: make(map[string]sectionCheckpoint), last: make(map[string]sectionCheckpoint),
	}
	if len(launch.UI) > 0 {
		var saved uiCheckpoint
		if err := json.Unmarshal(launch.UI, &saved); err != nil {
			return fmt.Errorf("invalid saved dashboard state: %w", err)
		}
		switch config.ViewType(saved.View) {
		case config.PRsView, config.IssuesView, config.NotificationsView, config.RepoView:
		default:
			return fmt.Errorf("invalid saved dashboard view %q", saved.View)
		}
		r.initial = &saved
	}
	m.recovery = r
	return nil
}

func sectionBase(s section.Section) *section.BaseModel {
	switch s := s.(type) {
	case *prssection.Model:
		return &s.BaseModel
	case *issuessection.Model:
		return &s.BaseModel
	case *notificationssection.Model:
		return &s.BaseModel
	}
	return nil
}

func recoveryKey(s section.Section) string { return fmt.Sprintf("%s/%d", s.GetType(), s.GetId()) }

func rowIdentities(s section.Section) []string {
	ids := make([]string, 0, s.NumRows())
	switch s := s.(type) {
	case *prssection.Model:
		for _, pr := range s.Prs {
			if pr.Primary != nil {
				ids = append(ids, pr.Primary.Url)
			} else {
				ids = append(ids, "")
			}
		}
	case *issuessection.Model:
		for _, issue := range s.Issues {
			ids = append(ids, issue.GetUrl())
		}
	case *notificationssection.Model:
		for _, n := range s.Notifications {
			ids = append(ids, n.GetId())
		}
	}
	return ids
}

// initRecovery runs before the first fetch. It restores section definitions,
// committed search filters and smart filtering, without restoring any pending
// edits, confirmations, or commands that could perform a GitHub mutation.
func (m *Model) initRecovery() {
	if m.recovery == nil || m.recovery.initial == nil {
		return
	}
	r := m.recovery
	saved := r.initial
	r.initial = nil
	prs := make([]config.PrsSectionConfig, 0)
	issues := make([]config.IssuesSectionConfig, 0)
	notifications := make([]config.NotificationsSectionConfig, 0)
	for _, s := range saved.Sections {
		switch config.ViewType(s.View) {
		case config.PRsView:
			cfg := config.PrsSectionConfig{}
			for _, existing := range m.ctx.Config.PRSections {
				if existing.Title == s.Title {
					cfg = existing
					break
				}
			}
			cfg.Title, cfg.Filters, cfg.Limit = s.Title, s.ConfiguredFilters, s.Limit
			prs = append(prs, cfg)
		case config.IssuesView:
			cfg := config.IssuesSectionConfig{}
			for _, existing := range m.ctx.Config.IssuesSections {
				if existing.Title == s.Title {
					cfg = existing
					break
				}
			}
			cfg.Title, cfg.Filters, cfg.Limit = s.Title, s.ConfiguredFilters, s.Limit
			issues = append(issues, cfg)
		case config.NotificationsView:
			notifications = append(notifications, config.NotificationsSectionConfig{
				Title: s.Title, Filters: s.ConfiguredFilters, Limit: s.Limit,
			})
		}
	}
	if len(prs) > 0 {
		m.ctx.Config.PRSections = prs
	}
	if len(issues) > 0 {
		m.ctx.Config.IssuesSections = issues
	}
	if len(notifications) > 0 {
		m.ctx.Config.NotificationsSections = notifications
	}
	m.initSections()
	indices := make(map[string]int)
	for _, s := range saved.Sections {
		sections := m.sectionsForView(config.ViewType(s.View))
		i := indices[s.View]
		indices[s.View]++
		if i >= len(sections) {
			continue
		}
		base := sectionBase(sections[i])
		if base == nil {
			continue
		}
		base.SearchValue = s.Search
		base.IsFilteredByCurrentRemote = s.SmartFiltered && m.ctx.HasGHRepo()
		base.ResetFilters()
		if s.Loaded > 0 || s.Selected != "" {
			r.pending[recoveryKey(sections[i])] = s
		}
	}
	m.ctx.View = config.ViewType(saved.View)
	if m.ctx.View == config.RepoView && !config.IsFeatureEnabled(config.FF_REPO_VIEW) {
		m.ctx.View = config.PRsView
	}
	m.currSectionId = max(0, min(saved.Section, len(m.getCurrentViewSections())-1))
	m.sidebar.IsOpen = saved.PreviewOpen
	if saved.Position == "right" || saved.Position == "bottom" {
		m.positionOverride = saved.Position
	}
	m.footer.ShowAll = saved.HelpOpen
	r.preview = saved
	m.syncMainContentDimensions()
	m.syncProgramContext()
}

func (m *Model) sectionsForView(view config.ViewType) []section.Section {
	switch view {
	case config.PRsView:
		return m.prs
	case config.IssuesView:
		return m.issues
	case config.NotificationsView:
		return m.notifications
	}
	return nil
}

// restoreSection follows the saved identity across pagination. GitHub search
// exposes at most 1000 results; if the PR no longer matches, use the old index.
func (m *Model) restoreSection(s section.Section) (tea.Cmd, bool) {
	if m.recovery == nil || s == nil {
		return nil, false
	}
	r := m.recovery
	saved, ok := r.pending[recoveryKey(s)]
	if !ok {
		return nil, false
	}
	base := sectionBase(s)
	if base == nil || base.PageInfo == nil || s.GetIsLoading() {
		return nil, true
	}
	ids := rowIdentities(s)
	idx := -1
	if saved.Selected != "" {
		idx = slices.Index(ids, saved.Selected)
	}
	if len(ids) < 1000 && base.PageInfo.HasNextPage &&
		(len(ids) < saved.Loaded || (saved.Selected != "" && idx == -1)) {
		return tea.Batch(s.FetchNextPageSectionRows()...), true
	}
	delete(r.pending, recoveryKey(s))
	matched := idx != -1
	if !matched {
		idx = saved.Row
	}
	base.Table.RestorePosition(idx, saved.Offset)
	if s == m.getCurrSection() {
		if !matched {
			r.preview = nil
		}
		m.syncSidebar()
		if m.ctx.View == config.PRsView {
			return m.prView.EnrichCurrRow(), true
		}
	}
	return nil, true
}

// Restore the preview's offset only after the fresh PR details have rendered;
// doing so on the initial loading message would clamp it back to the top.
func (m *Model) finishRecoveryPreview() {
	if m.recovery == nil || m.recovery.preview == nil {
		return
	}
	r := m.recovery
	s := m.getCurrSection()
	if s == nil {
		return
	}
	if _, pending := r.pending[recoveryKey(s)]; pending {
		return
	}
	if !m.sidebar.IsOpen {
		// Keep the saved scroll position until the panel is opened. Applying
		// it to a closed, empty viewport would clamp it to zero.
		return
	}
	row, ok := m.getCurrRowData().(*prrow.Data)
	if m.sidebar.IsOpen && m.ctx.View == config.PRsView && ok && row != nil && !row.IsEnriched {
		return
	}
	m.prView.RestorePreview(r.preview.PreviewTab, r.preview.Expanded)
	m.syncSidebar()
	m.sidebar.SetYOffset(r.preview.PreviewOffset)
	r.preview = nil
}

func (m *Model) writeCheckpoint() {
	if m.recovery == nil || m.ctx.Config == nil {
		return
	}
	r := m.recovery
	saved := uiCheckpoint{
		View: string(m.ctx.View), Section: m.currSectionId,
		PreviewOpen: m.sidebar.IsOpen, Position: m.positionOverride,
		PreviewTab: m.prView.SelectedTab(), Expanded: m.prView.SummaryExpanded(),
		PreviewOffset: m.sidebar.YOffset(), HelpOpen: m.footer.ShowAll,
	}
	if r.preview != nil {
		saved.PreviewTab, saved.Expanded, saved.PreviewOffset =
			r.preview.PreviewTab, r.preview.Expanded, r.preview.PreviewOffset
	}
	for _, view := range []config.ViewType{config.PRsView, config.IssuesView, config.NotificationsView} {
		for _, s := range m.sectionsForView(view) {
			if pending, ok := r.pending[recoveryKey(s)]; ok {
				saved.Sections = append(saved.Sections, pending)
				continue
			}
			base := sectionBase(s)
			if base.GetIsLoading() {
				if previous, ok := r.last[recoveryKey(s)]; ok {
					saved.Sections = append(saved.Sections, previous)
					continue
				}
			}
			entry := sectionCheckpoint{
				View: string(view), Title: s.GetConfig().Title,
				ConfiguredFilters: base.Config.Filters, Search: base.SearchValue,
				SmartFiltered: base.IsFilteredByCurrentRemote, Limit: base.Config.Limit,
				Row: s.CurrRow(), Loaded: s.NumRows(), Offset: base.Table.YOffset(),
			}
			if row := s.GetCurrRow(); row != nil {
				if n, ok := row.(*notificationrow.Data); ok {
					entry.Selected = n.GetId()
				} else {
					entry.Selected = row.GetUrl()
				}
			}
			saved.Sections = append(saved.Sections, entry)
			r.last[recoveryKey(s)] = entry
		}
	}
	b, err := json.Marshal(saved)
	if err == nil {
		c := r.launch
		c.Active, c.UI = !r.quitting, b
		err = r.store.Save(c)
	}
	if err != nil {
		log.Error("failed saving recovery checkpoint", "err", err)
	}
}
