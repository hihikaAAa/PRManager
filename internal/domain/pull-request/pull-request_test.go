package pullrequest

import (
	"testing"
	"time"
)

func TestPullRequestMerge_FirstTime(t *testing.T) {
	t.Parallel()

	pr := &PullRequest{
		ID:       "pr-1",
		Name:     "Test",
		AuthorID: "u1",
		Status:   StatusOpen,
	}

	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	pr.Merge(now)

	if pr.Status != StatusMerged {
		t.Fatalf("expected status %q, got %q", StatusMerged, pr.Status)
	}
	if pr.MergedAt == nil {
		t.Fatalf("expected non-nil MergedAt")
	}
	if !pr.MergedAt.Equal(now) {
		t.Fatalf("expected MergedAt %v, got %v", now, *pr.MergedAt)
	}
}

func TestPullRequestMerge_Idempotent(t *testing.T) {
	t.Parallel()

	first := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	second := first.Add(10 * time.Minute)

	pr := &PullRequest{
		ID:       "pr-1",
		Name:     "Test",
		AuthorID: "u1",
		Status:   StatusMerged,
		MergedAt: &first,
	}

	pr.Merge(second)

	if pr.Status != StatusMerged {
		t.Fatalf("expected status to stay %q, got %q", StatusMerged, pr.Status)
	}
	if pr.MergedAt == nil || !pr.MergedAt.Equal(first) {
		t.Fatalf("expected MergedAt to stay %v, got %v", first, pr.MergedAt)
	}
}
