package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
)

var (
	pullRequestColumns = []string{
		"Repository", "ID", "Author", "Title", "Reviewers", "Status", "Age",
	}
)

const (
	// EXPANDING_TABLE_COLUMN_IDX is the index of the column which should expand
	// to take up all available spare space.
	EXPANDING_TABLE_COLUMN_IDX = 3
	// REFERENCE_COLUMN is the column of a row that's expected to contain any
	// reference data - i.e a URL for the Pull Request associated with a row.
	REFERENCE_COLUMN = 0
)

// RowCollection defines an interface for managing collections of entities to
// display in a Table. It's largely a workaround due to the fact it's not possible
// to have slices of interfaces.
type RowCollection interface {
	PopulateTable(*tview.Table)
}

// Table is a wrapper around the `tview.Table` struct, and holds state for
// selections and events.
type Table struct {
	Primitive           *tview.Table
	currentRowReference string
}

// NewTable initialises and configures a `tview.Table` for displaying in the
// TUI; it configures the table with sane defaults such as borders, padding,
// selectability, and title options.
func NewTable(title string, rows RowCollection) *Table {
	t := &Table{
		Primitive: tview.NewTable().SetSelectable(true, false).SetBorders(true),
	}

	t.Primitive.SetSelectionChangedFunc(t.handlerSelection).
		SetInputCapture(t.handlerEventKey).
		SetBorderPadding(1, 1, 1, 1).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	t.Update(rows)
	return t
}

// Update a table, clearing it's contents and then re-populating it's cells from
// a `RowCollection`. This function does not re-draw the table. It is expected
// that this will happen via the Queue internal to the parent `tview.App`.
func (t *Table) Update(rows RowCollection) {
	for idx, title := range pullRequestColumns {
		t.Primitive.SetCell(0, idx, tview.NewTableCell(title).
			SetAlign(tview.AlignCenter).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false))
	}

	rows.PopulateTable(t.Primitive)
	t.expandColumn(EXPANDING_TABLE_COLUMN_IDX)
}

func (t *Table) expandColumn(col int) {
	// Iterate through each column and set the correct expansion.
	// @todo - could we just apply this to the header...?
	rowCount := t.Primitive.GetRowCount()
	for rowIdx := 1; rowIdx <= rowCount; rowIdx++ {
		t.Primitive.GetCell(rowIdx, col).SetExpansion(1)
	}
}

func (t *Table) handlerSelection(row, _ int) {
	// Resets the current "row reference" when a new row is selected.
	t.currentRowReference = ""
	if row < 1 || row > t.Primitive.GetRowCount() {
		// gaurd against OOB or header selection
		return
	}

	ref := t.Primitive.GetCell(row, REFERENCE_COLUMN).GetReference()
	if refString, isValid := ref.(string); isValid {
		t.currentRowReference = refString
	}
}

func (t *Table) handlerEventKey(evt *tcell.EventKey) *tcell.EventKey {
	// Detect user keypresses; at the moment 'o' = 'Open in Browser'.
	switch evt.Rune() {
	case 'o':
		if t.currentRowReference != "" {
			browser.OpenURL(t.currentRowReference)
		}
	}

	return evt
}
