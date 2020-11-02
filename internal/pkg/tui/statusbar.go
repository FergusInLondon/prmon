package tui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

const (
	// STATUS_FORMAT_STR is the format string for use with fmt.Sprintf, providing
	// the contents of the associated StatusBar in the UI.
	STATUS_FORMAT_STR = "[#AAAAAA]Signed in as [::b]%s[::-]. Polling at [::b]%d[::-] minute intervals. (Last synchronised at [::b]%s[::-])[-]"
	// STATUS_TIMESTAMP_FORMAT specifies the desired outputting format for any
	// timestamps.
	STATUS_TIMESTAMP_FORMAT = "15:04:05"
)

// StatusBar is a wrapper around the `TextView` `tview.Primitive`, and provides
// a helper method for updating the status based upon the time of a given sync.
// A future step may be to introduce some form of error output - i.e via a new
// format string, and an update function which accepts an `error`.
type StatusBar struct {
	generator func(time.Time) string
	Primitive *tview.TextView
}

// Create a new StatusBar complete with all state required.
func NewStatusBar(username string, pollInterval int, initialSync time.Time) *StatusBar {
	// Closure to capture pollInterview and Username, meaning subsequent updates
	// only require the `lastSync` value
	generator := func(lastSync time.Time) string {
		return fmt.Sprintf(STATUS_FORMAT_STR, username, pollInterval, lastSync.Format(STATUS_TIMESTAMP_FORMAT))
	}

	return &StatusBar{
		generator: generator,
		Primitive: tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(false).
			SetTextAlign(tview.AlignCenter).
			SetText(generator(initialSync)),
	}
}

// Update the StatusBar with the latest sync time; simply updates the internal
// TextView primitive.
func (sb *StatusBar) Update(latestSync time.Time) {
	sb.Primitive.SetText(sb.generator(latestSync))
}
