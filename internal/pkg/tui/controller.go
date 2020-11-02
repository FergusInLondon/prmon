// Package tui provides abstractions required for building the TUI, as well
// as helper functions for displaying data.
//
// It's unlikely that the calling package will require interactions beyond
// the three exposed functions available via `Controller`: `NewController`,
// `Update`, and `Run`.
//
// Additionally, a `State` struct is available which wraps around all the
// State that can be displayed to the user. This struct is used for when
// creating a new `Controller` - via `NewController(*State)` - and when an
// update to the TUI is available - via `Controller.Update(*State)`.
package tui

import (
	"context"
	"time"

	"github.com/FergusInLondon/PRList/internal/client/github"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	statusForegroundColours = map[string]tcell.Color{
		"bad":     tcell.GetColor("red"),
		"default": tcell.GetColor("yellow"),
		"good":    tcell.GetColor("green"),
	}
	statusBackgroundColours = map[string]tcell.Color{}
)

// Controller exposes no properties, and acts as an interface for managing the TUI.
type Controller struct {
	app             *tview.Application
	assignedPRTable *Table
	createdPRTable  *Table
	statusBar       *StatusBar
}

// State contains all the data required by the UI, it also acts as part of the
// interface for updating the UI.
type State struct {
	GithubUsername string
	PollInterval   int
	Assigned       []github.PullRequestSummary
	Created        []github.PullRequestSummary
	LastSync       time.Time
}

// NewController initialises all required UI components, returning a Controller
// that's ready to execute  and function.
func NewController(state *State) *Controller {
	return &Controller{
		assignedPRTable: NewTable("Assigned Pull Requests", PullRequestCollection{state.Assigned}),
		createdPRTable:  NewTable("Created Pull Requests", PullRequestCollection{state.Created}),
		statusBar:       NewStatusBar(state.GithubUsername, state.PollInterval, state.LastSync),
	}
}

// Update accepts a State struct, and conditionally updates the applicable UI
// components. It does so via a queued update to minimise UI redraws/prevent
// race conditions.
func (tui *Controller) Update(newState *State) {
	tui.app.QueueUpdateDraw(func() {
		if len(newState.Assigned) > 0 {
			tui.assignedPRTable.Update(PullRequestCollection{newState.Assigned})
		}

		if len(newState.Created) > 0 {
			tui.assignedPRTable.Update(PullRequestCollection{newState.Created})
		}

		tui.statusBar.Update(newState.LastSync)
	})
}

// Run executes the `tview.App` - enabling the TUI. Execution can be stopped by
// cancelling the provided Context.
func (tui *Controller) Run(ctx context.Context) error {
	grid := tview.NewGrid().
		SetRows(0, 0, 1).
		SetBorders(true).
		AddItem(tui.assignedPRTable.Primitive, 0, 0, 1, 1, 0, 0, true).
		AddItem(tui.createdPRTable.Primitive, 1, 0, 1, 1, 0, 0, false).
		AddItem(tui.statusBar.Primitive, 2, 0, 1, 1, 0, 0, false)

	tui.app = tview.NewApplication()
	tui.app.SetRoot(grid, true).
		SetFocus(grid).
		EnableMouse(true)

	go func(ctx context.Context) {
		<-ctx.Done()
		tui.app.Stop()
	}(ctx)

	return tui.app.Run()
}
