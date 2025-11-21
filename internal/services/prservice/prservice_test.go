package prservice

import (
	"testing"

	"github.com/hihikaAAa/PRManager/internal/domain/user"
)

func TestPickRandomReviewers_Empty(t *testing.T) {
	t.Parallel()

	got := pickRandomReviewers(nil, 2)
	if got != nil {
		t.Fatalf("expected nil slice, got %#v", got)
	}
}

func TestPickRandomReviewers_LessOrEqualThanLimit(t *testing.T) {
	t.Parallel()

	available := []*user.User{
		{ID: "u1"},
		{ID: "u2"},
	}

	got := pickRandomReviewers(available, 3)
	if len(got) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(got))
	}
	if got[0] != "u1" || got[1] != "u2" {
		t.Fatalf("unexpected reviewers slice: %#v", got)
	}
}

func TestPickRandomReviewers_MoreThanLimit(t *testing.T) {
	t.Parallel()

	available := []*user.User{
		{ID: "u1"},
		{ID: "u2"},
		{ID: "u3"},
		{ID: "u4"},
	}

	limit := 2
	got := pickRandomReviewers(available, limit)

	if len(got) != limit {
		t.Fatalf("expected %d reviewers, got %d", limit, len(got))
	}

	seen := map[string]bool{}
	for _, id := range got {
		seen[id] = true
		found := false
		for _, u := range available {
			if u.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("id %q not from available list", id)
		}
	}

	if len(seen) != len(got) {
		t.Fatalf("expected all reviewers unique, got %v", got)
	}
}

func TestPickRandomReviewers_BigLimit(t *testing.T) {
	t.Parallel()

	available := []*user.User{
		{ID: "u1"},
		{ID: "u2"},
	}

	got := pickRandomReviewers(available, 100)
	if len(got) != len(available) {
		t.Fatalf("expected %d reviewers, got %d", len(available), len(got))
	}
}
