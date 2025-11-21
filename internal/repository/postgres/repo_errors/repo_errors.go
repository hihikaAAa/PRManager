package repo_errors

import "errors"

var (
	ErrUserNotFound = errors.New("users not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrReviewersNotFound = errors.New("reviewers not found")
	ErrPRNotFound = errors.New("pr not found")
	ErrPRMerged = errors.New("pull request already merged")
)