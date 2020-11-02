// Package github implements the interactions required to interface
// with Github, notably a simple client that regularly polls Github's
// API for new Pull Requests/Issues.
//
// Additional functionality (in the way of change detection) is also
// included in this repository.
//
// Note that the implementation isn't ideal, and a likely improvement
// will be the usage of the Github v4 API (GraphQL based) to reduce
// the number of requests required, and also migrate away from `go-github`
// which is full of a lot of auto-generated noise.
package github

import (
	"context"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// Gotcha alert! You would imagine that `created` Issues would be a subset of
// those returned by `all`... But actually it's a entirely different set. This
// is one of a few inconsistencies with the Github API, and another reason for
// looking towards the GraphQL API.
const (
	// ASSIGNED_FILTER specifies that ALL Issues should be returned from Github
	ASSIGNED_FILTER = "all"
	// CREATED_FILTER specifies that ONLY CREATED Issues should be returned from Github
	CREATED_FILTER = "created"
)

// PollerNotificationChannels is a wrapper around two channels used for notifying
// calling code of new data being made available via the Poller.
type PollerNotificationChannels struct {
	// LatestPollTimestamp updates the calling code whenever a new Poll is made
	LatestPollTimestamp chan time.Time
	// NewDataAvailable informs the calling code that a difference has been detected
	// in the most recent API response.
	NewDataAvailable chan struct{}
}

// NewPollerNotificationChannels provides a ready-to-use `PollerNotificationChannels`.
func NewPollerNotificationChannels() *PollerNotificationChannels {
	return &PollerNotificationChannels{
		LatestPollTimestamp: make(chan time.Time),
		NewDataAvailable:    make(chan struct{}),
	}
}

// Poller regularly polls the Github API for new Pull Requests, and
// communicates any changes via a `PollerNotificationChannels` struct.
// Poller embeds a RWMutex, this is used when updating the internal
// PullRequest collections.
type Poller struct {
	sync.Mutex
	ctx    context.Context
	client *github.Client
	// Timestamp of the last time the Poller queries the Github API
	LastPolled time.Time
	// Username - or 'Login' - of the currently authenticated Github user
	Username string
	// Collection of Pull Requests assigned to the current user
	AssignedPullRequests *PullRequestSummaryCollection
	// Collection of Pull Requests *created* by the current user
	CreatedPullRequests *PullRequestSummaryCollection
}

// NewPoller configures a new `Poller` struct, authenticating with Github and
// making the initial API request. The returned `Poller` is fully populated
// data, however is not configured to Poll automatically - this happens after
// `Poll` is called.
func NewPoller(ctx context.Context, token string) *Poller {
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	))

	ghClient := github.NewClient(oauthClient)
	currentUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}

	poller := &Poller{
		ctx:      ctx,
		client:   ghClient,
		Username: currentUser.GetLogin(),
	}

	assigned, created := poller.pullRequests(ASSIGNED_FILTER, CREATED_FILTER)
	poller.AssignedPullRequests = NewPullRequestSummaryCollection(assigned)
	poller.CreatedPullRequests = NewPullRequestSummaryCollection(created)
	poller.LastPolled = time.Now()
	return poller
}

// Poll queries the Github API every `pauseInterval`, and updates the calling
// code via the `PollerNotificationChannels`. Polling can be stopped via the
// context provided when calling `NewPoller`.
func (poller *Poller) Poll(notificationChannels *PollerNotificationChannels, pauseInterval time.Duration) {
	for {
		select {
		case <-time.After(pauseInterval):
			assigned, created := poller.pullRequests(ASSIGNED_FILTER, CREATED_FILTER)
			notificationChannels.LatestPollTimestamp <- time.Now()

			poller.Lock()
			haveUpdatedAssignations := poller.AssignedPullRequests.Update(assigned)
			haveUpdatedCreations := poller.CreatedPullRequests.Update(created)

			if haveUpdatedAssignations || haveUpdatedCreations {
				notificationChannels.NewDataAvailable <- struct{}{}
			}
			poller.Unlock()
		case <-poller.ctx.Done():
			return
		}
	}
}

func (poller *Poller) pullRequests(assignedFilterQuery, createdFilterQuery string) ([]PullRequestSummary, []PullRequestSummary) {
	// simply retrieve pullRequests. we retrieve two sets, so there's an enclosed function
	// just to simplify the flow.
	get := func(collection []PullRequestSummary, filterString string) []PullRequestSummary {
		issues, _, err := poller.client.Issues.List(poller.ctx, true, &github.IssueListOptions{
			Filter: filterString,
		})
		if err != nil {
			panic(err)
		}

		for _, issue := range issues {
			if issue.IsPullRequest() {
				owner := issue.Repository.GetOwner().GetLogin()
				repo := issue.Repository.GetName()

				if pullRequest, _, err := poller.client.PullRequests.Get(
					poller.ctx, owner, repo, issue.GetNumber(),
				); err == nil {
					collection = append(collection, NewPullRequestFromAPI(issue, pullRequest))
				}
			}
		}

		return collection
	}

	return get(make([]PullRequestSummary, 0), assignedFilterQuery),
		get(make([]PullRequestSummary, 0), createdFilterQuery)
}
