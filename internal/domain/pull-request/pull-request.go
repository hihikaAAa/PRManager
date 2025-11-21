package pullrequest

import (
	"time"
)

type Status string

const(
	StatusOpen Status = "OPEN"
	StatusMerged Status = "MERGED"
)

type PullRequest struct{
	ID string
	Name string
	AuthorID string
	Status Status
	Reviewers []string

	CreatedAt time.Time
	MergedAt *time.Time
}

type PullRequestShort struct{
	ID string `json:"pull_request_id"`
	Name string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status Status `json:"status"`
}

func (pr *PullRequest) Merge(t time.Time) {
    if pr.Status == StatusMerged {
        return
    }
    pr.Status = StatusMerged
    pr.MergedAt = &t
}
