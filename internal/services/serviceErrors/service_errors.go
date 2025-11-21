package serviceerrors

import "errors"

var (
	ErrPRExists = errors.New("pr already exists")
	ErrPRMerged = errors.New("pr already merged")
	ErrReviewerNotFound = errors.New("reviewer not found")
	ErrNoCandidates = errors.New("no candidates")
	ErrTeamExists = errors.New("team already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
)