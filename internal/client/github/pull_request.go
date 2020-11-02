package github

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v32/github"
)

// PullRequestSummaryCollection is a wrapper around a slice of `PullRequestSummary`
// structs, and a ChangeChecker. It's used for interacting with the items returned
// from the Github API.
type PullRequestSummaryCollection struct {
	dirtyChecker *ChangeChecker
	// Summary versions of the PullRequests taken from the Github API
	Items []PullRequestSummary
}

// NewPullRequestSummaryCollection returns a new PullRequestSummaryCollection from
// a slice of Summary entities, it also configures the Collection with a ChangeChecker
// using the default keyChecker.
func NewPullRequestSummaryCollection(items []PullRequestSummary) *PullRequestSummaryCollection {
	collection := &PullRequestSummaryCollection{
		Items: items,
	}

	collection.dirtyChecker = NewChangeChecker(collection, nil)
	return collection
}

// VersionKeys returns a list of keys for the contained `PullRequestSummary` entities.
func (collection *PullRequestSummaryCollection) VersionKeys() []string {
	versionKeys := make([]string, len(collection.Items))
	for idx, item := range collection.Items {
		versionKeys[idx] = item.versionKey()
	}
	return versionKeys
}

// Update accepts a new slice of `PullRequestSummary` structs, and returns a boolean
// indicating whether the new entities differ from the existing ones.
func (collection *PullRequestSummaryCollection) Update(latestItems []PullRequestSummary) bool {
	collection.Items = latestItems
	return collection.dirtyChecker.HasChanged(collection)
}

// PullRequestSummary is a simplified PullRequest object which only contains a
// subset of the properties returned from the Github API.
type PullRequestSummary struct {
	Draft         bool
	Author        string
	Title         string
	Repository    string
	ID            string
	ReviewerCount int
	Status        string
	OpenedAt      time.Time
	URL           string
}

// NewPullRequestFromAPI generates a `PullRequestSummary` from the records returned
// from the Github API - both a `github.Issue` and a `github.PullRequest`. A quirk
// of the API is that a Pull Request *is an Issue*, but it has additional fields
// that need to be retrieved via a secondary API cal - for this reason both params
// are required.
func NewPullRequestFromAPI(issue *github.Issue, pr *github.PullRequest) PullRequestSummary {
	return PullRequestSummary{
		Draft:      pr.GetDraft(),
		Author:     issue.GetUser().GetLogin(),
		Title:      issue.GetTitle(),
		Repository: issue.Repository.GetName(),
		ID:         strconv.Itoa(issue.GetNumber()),
		Status:     issue.GetState(),
		OpenedAt:   issue.GetCreatedAt(),
		URL:        pr.GetHTMLURL(),
	}
}

// We care about whether it's the same repository, whether it's status has changed
// whether it remains a draft. Key = [repository]:[prNum]:[status]:[draft]
func (pr PullRequestSummary) versionKey() string {
	draftStr := "N"
	if pr.Draft {
		draftStr = "Y"
	}

	return fmt.Sprintf("%s:%s:%s:%s", pr.Repository, pr.ID, pr.Status, draftStr)
}
