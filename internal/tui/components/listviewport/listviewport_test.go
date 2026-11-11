package listviewport

import (
	"strings"
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type testModelOpts struct {
	numItems       int
	viewportHeight int
	itemHeight     int
}

// numItems, viewportHeight, itemHeight int
func newTestModel(opts testModelOpts) Model {
	ctx := context.ProgramContext{}
	dims := constants.Dimensions{Width: 80, Height: opts.viewportHeight}
	now := time.Now()
	return NewModel(ctx, dims, now, now, "items", opts.numItems, opts.itemHeight)
}

func TestPrevItemScrollsAtTopBound(t *testing.T) {
	// Initial bounds: topBoundId=0, bottomBoundId=4, currId=0
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	// Navigate down past the first page boundary
	// After 5 NextItem calls: currId=5, topBoundId=1, bottomBoundId=5
	for range 5 {
		m.NextItem()
	}

	if m.GetCurrItem() != 5 {
		t.Fatalf("expected currId=5 after 5 NextItem calls, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Fatalf("expected topBoundId=1, got %d", m.topBoundId)
	}

	// Navigate back up to the top bound
	// 4 PrevItem calls: currId goes 4, 3, 2, 1
	for range 4 {
		m.PrevItem()
	}

	if m.GetCurrItem() != 1 {
		t.Fatalf("expected currId=1, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Fatalf("expected topBoundId=1 (no scroll yet), got %d", m.topBoundId)
	}

	// Now one more PrevItem: currId should go to 0, and the viewport
	// should scroll up so that item 0 is visible.
	// Before fix: topBoundId stays at 1, item 0 is above the viewport.
	// After fix: topBoundId becomes 0, item 0 is at the top of the viewport.
	m.PrevItem()

	if m.GetCurrItem() != 0 {
		t.Fatalf("expected currId=0, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Errorf(
			"PrevItem did not scroll up at the top boundary: expected topBoundId=0, got %d (item 0 is above the visible viewport)",
			m.topBoundId,
		)
	}
}

func TestNextItemScrollsAtBottomBound(t *testing.T) {
	// Verify NextItem scrolling works symmetrically
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	// Navigate down to bottomBoundId (4)
	for range 4 {
		m.NextItem()
	}

	if m.GetCurrItem() != 4 {
		t.Fatalf("expected currId=4, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Fatalf("expected topBoundId=0 (no scroll yet), got %d", m.topBoundId)
	}

	// One more NextItem at the bottom boundary should scroll
	m.NextItem()

	if m.GetCurrItem() != 5 {
		t.Fatalf("expected currId=5, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Errorf(
			"NextItem did not scroll down at the bottom boundary: expected topBoundId=1, got %d",
			m.topBoundId,
		)
	}
}

func TestPrevItemAtFirstItem(t *testing.T) {
	// Pressing up when already at the first item should stay at 0
	// and should not corrupt viewport bounds.
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	m.PrevItem()

	if m.GetCurrItem() != 0 {
		t.Errorf("expected currId=0, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Errorf("expected topBoundId=0 (should not scroll), got %d", m.topBoundId)
	}
	if m.bottomBoundId != 4 {
		t.Errorf("expected bottomBoundId=4 (should not scroll), got %d", m.bottomBoundId)
	}
}

func TestNextItemAtLastItem(t *testing.T) {
	// Pressing down when already at the last item should stay at last
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	for range 20 {
		m.NextItem()
	}

	if m.GetCurrItem() != 9 {
		t.Errorf("expected currId=9, got %d", m.GetCurrItem())
	}
}

func TestSetCurrItemKeepsScrollPositionWhenVisible(t *testing.T) {
	m := newTestModel(testModelOpts{numItems: 20, viewportHeight: 10, itemHeight: 2})
	m.SyncViewPort(strings.Repeat("row\n", 40))

	// scroll down so that items 3-7 are visible
	for range 7 {
		m.NextItem()
	}
	if m.viewport.YOffset() != 6 {
		t.Fatalf("expected the viewport to be scrolled to line 6, got %d", m.viewport.YOffset())
	}

	// selecting another visible item must not scroll
	if got := m.SetCurrItem(4); got != 4 {
		t.Fatalf("expected currId=4, got %d", got)
	}
	if m.viewport.YOffset() != 6 {
		t.Errorf("expected the scroll position to be kept, got %d", m.viewport.YOffset())
	}
	if m.topBoundId != 3 || m.bottomBoundId != 7 {
		t.Errorf("expected bounds 3-7, got %d-%d", m.topBoundId, m.bottomBoundId)
	}
}

func TestSetCurrItemScrollsHiddenItemIntoView(t *testing.T) {
	m := newTestModel(testModelOpts{numItems: 20, viewportHeight: 10, itemHeight: 2})
	m.SyncViewPort(strings.Repeat("row\n", 40))

	// below the viewport: the item becomes the last visible one
	m.SetCurrItem(12)
	if m.viewport.YOffset() != 16 {
		t.Errorf("expected the viewport to be scrolled to line 16, got %d", m.viewport.YOffset())
	}
	if m.topBoundId != 8 || m.bottomBoundId != 12 {
		t.Errorf("expected bounds 8-12, got %d-%d", m.topBoundId, m.bottomBoundId)
	}

	// above the viewport: the item becomes the first visible one
	m.SetCurrItem(2)
	if m.viewport.YOffset() != 4 {
		t.Errorf("expected the viewport to be scrolled to line 4, got %d", m.viewport.YOffset())
	}
	if m.topBoundId != 2 || m.bottomBoundId != 6 {
		t.Errorf("expected bounds 2-6, got %d-%d", m.topBoundId, m.bottomBoundId)
	}
}

func TestSetCurrItemClampsToAvailableItems(t *testing.T) {
	m := newTestModel(testModelOpts{numItems: 3, viewportHeight: 10, itemHeight: 1})

	if got := m.SetCurrItem(10); got != 2 {
		t.Errorf("expected currId to be clamped to 2, got %d", got)
	}
	if got := m.SetCurrItem(-1); got != 0 {
		t.Errorf("expected currId to be clamped to 0, got %d", got)
	}

	m.SetNumItems(0)
	if got := m.SetCurrItem(5); got != 0 {
		t.Errorf("expected currId=0 for an empty list, got %d", got)
	}
}
