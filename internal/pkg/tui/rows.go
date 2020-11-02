package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/FergusInLondon/PRList/internal/client/github"
	"github.com/gdamore/tcell/v2"
	"github.com/hako/durafmt"
	"github.com/rivo/tview"
)

// PullRequestCollection provides a `RowCollection` interface for a collection
// of `github.PullRequestSummary` structs.
type PullRequestCollection struct {
	PullReqs []github.PullRequestSummary
}

// PopulateTable populates a provided `tview.Table` with rows associated with
// available `github.PullRequestSummary` structs.
func (collection PullRequestCollection) PopulateTable(table *tview.Table) {
	for idx, pullReq := range collection.PullReqs {
		(PullRequestRow{pullReq}).Cells(idx+1, table)
	}
}

// `PullRequestRow` is a wrapper around the `github.PullRequestSummary` object,
// providing a display layer.
type PullRequestRow struct {
	github.PullRequestSummary
}

// Cells generates all the `tview.TableCell` structs required for a row representing
// a `github.PullRequestSummary`. It additionally sets the meta-data - i.e Reference
// - of the given Row.
func (pr PullRequestRow) Cells(idx int, table *tview.Table) {
	if pr.Draft {
		// Draft Rows are slightly different; no variable styling, and dimmed text
		pr.draftRow(idx, table)
		pr.setReference(idx, table)
		return
	}

	table.SetCell(idx, 0, tview.NewTableCell(pr.Repository).SetAlign(tview.AlignCenter))
	table.SetCell(idx, 1, tview.NewTableCell(pr.ID).SetAlign(tview.AlignCenter))
	table.SetCell(idx, 2, tview.NewTableCell(pr.Author).SetAlign(tview.AlignCenter))
	table.SetCell(idx, 3, tview.NewTableCell(pr.Title).SetAlign(tview.AlignCenter))
	table.SetCell(idx, 4, pr.reviewerCountCell())
	table.SetCell(idx, 5, pr.statusCell())
	table.SetCell(idx, 6, pr.openedAtCell())
	pr.setReference(idx, table)
}

func (pr PullRequestRow) setReference(idx int, table *tview.Table) {
	// store the URL of the PR in the 'reference column'; the 'reference column'
	// is a known column in the table used for storing meta-data.
	table.GetCell(idx, REFERENCE_COLUMN).SetReference(pr.URL)
}

func (pr PullRequestRow) statusCell() *tview.TableCell {
	// provide basic formating on the status of a given PullRequest
	statusCell := tview.NewTableCell(pr.Status).
		SetAttributes(tcell.AttrBold).
		SetAlign(tview.AlignCenter)

	statusKey := strings.ToLower(pr.Status)
	if textColour, hasColour := statusForegroundColours[statusKey]; hasColour {
		statusCell.SetTextColor(textColour)
	}

	if backgroundColour, hasColour := statusBackgroundColours[statusKey]; hasColour {
		statusCell.SetBackgroundColor(backgroundColour)
	}

	return statusCell
}

func (pr PullRequestRow) reviewerCountCell() *tview.TableCell {
	// provide basic formating on the number of reviewers for a given PullRequest
	textColour := "red"
	if pr.ReviewerCount > 0 {
		textColour = "green"
	}

	return tview.
		NewTableCell(strconv.Itoa(pr.ReviewerCount)).
		SetAlign(tview.AlignCenter).
		SetAttributes(tcell.AttrBold).
		SetTextColor(tcell.GetColor(textColour))
}

func (pr PullRequestRow) openedAtCell() *tview.TableCell {
	// Traffic-Light style colour codes for the "freshness" of a given PullRequest.
	since := time.Now().Sub(pr.OpenedAt)

	textColour := tcell.GetColor("#FF0000")
	if since.Hours() <= 1 {
		textColour = tcell.GetColor("#00FF00")
	} else if since.Hours() <= 4 {
		textColour = tcell.GetColor("#0000FF")
	} else if since.Hours() <= 8 {
		textColour = tcell.GetColor("#FF0D0E")
	}

	return tview.
		NewTableCell(prettyPrintDuration(int(since.Seconds()))).
		SetAlign(tview.AlignCenter).
		SetAttributes(tcell.AttrBold).
		SetTextColor(textColour)
}

func (pr PullRequestRow) draftRow(idx int, table *tview.Table) {
	// Draft Rows are just dimmed with minimal styling.
	rowContents := []string{pr.Repository, pr.ID, pr.Author, pr.Title, "-", "draft", "-"}

	for colIdx, colVal := range rowContents {
		cell := tview.NewTableCell(colVal).
			SetAlign(tview.AlignCenter).
			SetAttributes(tcell.AttrDim)

		table.SetCell(idx, colIdx, cell)
	}
}

func prettyPrintDuration(seconds int) string {
	// `durafmt` isn't flexible enough to allow us to avoid `ms` units; so this
	// is an ugly workaround to ensure that there are never `ms` in any durations!
	reducedDuration := time.Duration(seconds) * time.Second
	return durafmt.Parse(reducedDuration).LimitFirstN(3).String()
}
