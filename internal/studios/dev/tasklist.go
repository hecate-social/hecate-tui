package dev

import (
	"strings"

	"github.com/hecate-social/hecate-tui/internal/client"
)

// TaskItem is a single row in the task list — either a venture task,
// a division header, or a division sub-task.
type TaskItem struct {
	Verb       string // "initiate-venture", "design", etc.
	Label      string // Human-readable: "Initiate Venture", "Design", etc.
	Scope      string // "venture" or "division"
	Subject    string // Division name (empty for venture tasks)
	Phase      string // "DnA", "AnP", "TnI", "DnO"
	AIRole     string
	State      string // "blocked", "pending", "active", "paused", "running", "done"
	Depth      int    // 0 = venture task, 1 = division header, 2 = division sub-task
	IsHeader   bool   // true for division group headers
	Collapsed  bool   // true if this division group is collapsed (only on headers)
	DivisionID string // set on headers and sub-tasks
}

// TaskList manages the visible task items, cursor, and scroll state.
type TaskList struct {
	allItems   []TaskItem // everything, including collapsed sub-tasks
	cursor     int
	viewHeight int
}

// BuildFromResponse populates the task list from a daemon response.
func (tl *TaskList) BuildFromResponse(resp *client.VentureTaskList) {
	tl.allItems = tl.allItems[:0]

	// Venture-level tasks
	for _, t := range resp.Tasks {
		tl.allItems = append(tl.allItems, TaskItem{
			Verb:   t.Verb,
			Label:  verbLabel(t.Verb),
			Scope:  "venture",
			Phase:  t.Phase,
			AIRole: t.AIRole,
			State:  t.State,
			Depth:  0,
		})
	}

	// Division groups
	for _, div := range resp.Divisions {
		// Division header
		tl.allItems = append(tl.allItems, TaskItem{
			Verb:       "",
			Label:      div.Name,
			Scope:      "division",
			Subject:    div.Name,
			Depth:      1,
			IsHeader:   true,
			DivisionID: div.ID,
		})
		// Division sub-tasks
		for _, t := range div.Tasks {
			tl.allItems = append(tl.allItems, TaskItem{
				Verb:       t.Verb,
				Label:      verbLabel(t.Verb),
				Scope:      "division",
				Subject:    div.Name,
				Phase:      t.Phase,
				AIRole:     t.AIRole,
				State:      t.State,
				Depth:      2,
				DivisionID: div.ID,
			})
		}
	}

	// Reset cursor if out of bounds
	visible := tl.VisibleItems()
	if tl.cursor >= len(visible) {
		tl.cursor = max(0, len(visible)-1)
	}
}

// VisibleItems returns items respecting collapsed state.
func (tl *TaskList) VisibleItems() []TaskItem {
	var items []TaskItem
	collapsed := ""
	for _, item := range tl.allItems {
		if item.IsHeader {
			collapsed = ""
			if item.Collapsed {
				collapsed = item.DivisionID
			}
			items = append(items, item)
			continue
		}
		if collapsed != "" && item.DivisionID == collapsed {
			continue
		}
		items = append(items, item)
	}
	return items
}

// Len returns the visible item count.
func (tl *TaskList) Len() int {
	return len(tl.VisibleItems())
}

// Cursor returns the current cursor position.
func (tl *TaskList) Cursor() int {
	return tl.cursor
}

// SetViewHeight sets the visible viewport height.
func (tl *TaskList) SetViewHeight(h int) {
	tl.viewHeight = h
}

// Up moves the cursor up by one.
func (tl *TaskList) Up() {
	if tl.cursor > 0 {
		tl.cursor--
	}
}

// Down moves the cursor down by one.
func (tl *TaskList) Down() {
	n := tl.Len()
	if tl.cursor < n-1 {
		tl.cursor++
	}
}

// Top moves the cursor to the first item.
func (tl *TaskList) Top() {
	tl.cursor = 0
}

// Bottom moves the cursor to the last item.
func (tl *TaskList) Bottom() {
	n := tl.Len()
	if n > 0 {
		tl.cursor = n - 1
	}
}

// ToggleCollapse toggles the collapsed state of the division header at the cursor.
// If the cursor is on a sub-task, it toggles the parent division header.
func (tl *TaskList) ToggleCollapse() {
	visible := tl.VisibleItems()
	if tl.cursor >= len(visible) {
		return
	}
	item := visible[tl.cursor]
	divID := item.DivisionID
	if divID == "" {
		return // venture task, nothing to collapse
	}

	// Find the header in allItems and toggle it
	for i := range tl.allItems {
		if tl.allItems[i].IsHeader && tl.allItems[i].DivisionID == divID {
			tl.allItems[i].Collapsed = !tl.allItems[i].Collapsed
			break
		}
	}

	// If cursor is now past visible items, clamp it
	newVisible := tl.VisibleItems()
	if tl.cursor >= len(newVisible) {
		tl.cursor = max(0, len(newVisible)-1)
	}
}

// SelectedItem returns the item at the cursor, or nil if empty.
func (tl *TaskList) SelectedItem() *TaskItem {
	visible := tl.VisibleItems()
	if tl.cursor >= len(visible) {
		return nil
	}
	item := visible[tl.cursor]
	return &item
}

// ScrollOffset returns the scroll offset for rendering within the viewport.
func (tl *TaskList) ScrollOffset() int {
	if tl.viewHeight <= 0 {
		return 0
	}
	// Keep cursor visible within the viewport
	offset := 0
	if tl.cursor >= tl.viewHeight {
		offset = tl.cursor - tl.viewHeight + 1
	}
	return offset
}

// verbLabel converts a verb slug to a human-readable label.
func verbLabel(verb string) string {
	switch verb {
	case "initiate-venture":
		return "Initiate Venture"
	case "refine-vision":
		return "Refine Vision"
	case "submit-vision":
		return "Submit Vision"
	case "refine-divisions":
		return "Refine Divisions"
	case "design":
		return "Design"
	case "plan":
		return "Plan"
	case "generate":
		return "Generate"
	case "test":
		return "Test"
	case "deploy":
		return "Deploy"
	case "monitor":
		return "Monitor"
	case "rescue":
		return "Rescue"
	default:
		return strings.Title(strings.ReplaceAll(verb, "-", " ")) //nolint:staticcheck
	}
}
