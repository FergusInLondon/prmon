// The `tui` app runs a basic TUI showing information on open assigned/created
// PullRequests for the currently authenticated Github User. Additionally, it
// has the functionality to open a browser window targeting a selected PR.
//
// It works by polling Github on a regular basis, comparing the returned results
// with from the previous poll, and updating the TUI if there's any change.
package main

import (
	"os"

	"github.com/FergusInLondon/PRList/internal/client/github"
	"github.com/FergusInLondon/PRList/internal/pkg/tui"
	"github.com/mkideal/cli"

	"context"
	"time"
)

type CLIArgs struct {
	cli.Helper
	Debug        bool   `cli:"debug" usage:"debug - poll more frequently" dft:"false"`
	GithubToken  string `cli:"token" usage:"github personal access token" dft:"$GH_TOKEN"`
	PollDuration int    `cli:"duration" usage:"duration - in minutes - to wait between polling github" dft:"5"`
}

func app(gh_token string, poll_duration int, debug bool) error {
	waitMins := poll_duration
	if debug {
		// Debug is quite literally "don't wait as much, and hopefully any errors
		// will happen quicker/more frequently". Should likely implement some form
		// of runtime logging.
		waitMins = 1
	}

	ctx, stopper := context.WithCancel(context.Background())
	defer stopper()

	// Initialise our Github Poller, and generate the State required for the TUI
	ghPoller := github.NewPoller(ctx, gh_token)
	tuiController := tui.NewController(&tui.State{
		GithubUsername: ghPoller.Username,
		PollInterval:   waitMins,
		Assigned:       ghPoller.AssignedPullRequests.Items,
		Created:        ghPoller.CreatedPullRequests.Items,
		LastSync:       ghPoller.LastPolled,
	})

	// Glue together notifications from the Github Poller with the TUI Controller
	go func(notifyChans *github.PollerNotificationChannels) {
		ghPoller.Poll(notifyChans, (time.Duration(waitMins) * time.Minute))
		for {
			select {
			case <-ctx.Done():
				return
			case latestTimestamp := <-notifyChans.LatestPollTimestamp:
				tuiController.Update(&tui.State{
					LastSync: latestTimestamp,
				})
			case <-notifyChans.NewDataAvailable:
				// ! Bug: the timestamps associated with a PR in the TUI will
				// ! become stale if there are no updates from Github..!
				tuiController.Update(&tui.State{
					Assigned: ghPoller.AssignedPullRequests.Items,
					Created:  ghPoller.CreatedPullRequests.Items,
				})
			}
		}
	}(github.NewPollerNotificationChannels())

	return tuiController.Run(ctx)
}

func main() {
	os.Exit(cli.Run(new(CLIArgs), func(ctx *cli.Context) error {
		params := ctx.Argv().(*CLIArgs)
		return app(params.GithubToken, params.PollDuration, params.Debug)
	}))
}
