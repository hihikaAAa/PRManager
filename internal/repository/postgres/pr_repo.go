package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type PRRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *PRRepository{
	return &PRRepository{db: db}
}

func (r *PRRepository) CreateWithReviewers(ctx context.Context, pr pullrequest.PullRequest) error{
	const op = "internal.repository.postgres.pr_repo.CreateWithReviewers"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil{
		return fmt.Errorf("%s, BeginTx: %w", op ,err)
	}
	defer tx.Rollback()

	const qPR = `
	INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := tx.ExecContext(ctx, qPR, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt, pr.MergedAt); err != nil{
		return fmt.Errorf("%s, ExecContextPr: %w", op, err)
	}

	const qRev = `
	INSERT INTO pull_request_reviewers (pull_request_id, user_id)
	VALUES ($1, $2)
	`

	for _, reviewerID := range pr.Reviewers{
		if _, err := tx.ExecContext(ctx, qRev, pr.ID, reviewerID); err != nil{
			return fmt.Errorf("%s, ExecContextReviewer: %w", op, err)
		}
	}
	if err := tx.Commit(); err != nil{
		return fmt.Errorf("%s, Commit: %w", op, err)
	}

	return nil
}

func (r *PRRepository) GetWithReviewers(ctx context.Context, id string)(*pullrequest.PullRequest, error){
	const op = "internal.repository.postgres.pr_repo.GetWithReviewers"

	const qPR = `
	SELECT pull_request_id, pull_request_name, author_id, status, created_at,merged_at
	FROM pull_requests 
	WHERE pull_request_id = $1;
	`

	pr := &pullrequest.PullRequest{}
	if err := r.db.QueryRowContext(ctx, qPR, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil{
		if err == sql.ErrNoRows{
			return nil, fmt.Errorf("%s: %w", op, repo_errors.ErrPRNotFound)
		}
		return nil , fmt.Errorf("%s, QueryRow: %w", op, err)
	}

	const qRev = `
	SELECT user_id
	FROM pull_request_reviewers
	WHERE pull_request_id = $1
	`

	rows, err := r.db.QueryContext(ctx, qRev, id)
	if err != nil{
		return nil, fmt.Errorf("%s, QueryContextRev: %w", op, err)
	}
	defer rows.Close()

	for rows.Next(){
		var uid string
		if err := rows.Scan(&uid); err != nil{
			return nil, fmt.Errorf("%s, Scan: %w", op, err)
		}
		pr.Reviewers = append(pr.Reviewers, uid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return pr, nil
}

func (r *PRRepository) Merge(ctx context.Context, id string, now time.Time) (*pullrequest.PullRequest, error) {
	const op = "internal.repository.postgres.pr_repo.Merge"

	const qUpdate = `
		UPDATE pull_requests
		SET status   = $2, merged_at = COALESCE(merged_at, $3)
		WHERE pull_request_id = $1 AND status = 'OPEN'
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at;
	`
	pr := &pullrequest.PullRequest{}

	err := r.db.QueryRowContext(ctx, qUpdate, id, pullrequest.StatusMerged, now).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err == nil {
		return pr, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s, QueryRowContext: %w", op, err)
	}
	const qSelect = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1;
	`

	pr = &pullrequest.PullRequest{}
	if err = r.db.QueryRowContext(ctx, qSelect, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil{
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, repo_errors.ErrPRNotFound)
		}
		return nil, fmt.Errorf("%s, QueryRowContext: %w", op, err)
	}
	return pr, nil
}


func (r *PRRepository) ReplaceReviewers(ctx context.Context, prID, oldRevID, newRevID string) error {
	const op = "internal.repository.postgres.pr_repo.ReplaceReviewers"

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s, BeginTx: %w", op, err)
	}
	defer tx.Rollback()
	const qPR = `
		SELECT status
		FROM pull_requests
		WHERE pull_request_id = $1;
	`

	var status pullrequest.Status
	if err := tx.QueryRowContext(ctx, qPR, prID).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("%s: %w", op, repo_errors.ErrPRNotFound)
		}
		return fmt.Errorf("%s, QueryRow PR: %w", op, err)
	}

	if status == pullrequest.StatusMerged {
		return fmt.Errorf("%s: %w", op, repo_errors.ErrPRMerged)
	}

	const qDel = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND user_id = $2;
	`

	res, err := tx.ExecContext(ctx, qDel, prID, oldRevID)
	if err != nil {
		return fmt.Errorf("%s, Exec delete: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s, RowsAffected delete: %w", op, err)
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", op, repo_errors.ErrReviewersNotFound)
	}

	const qIns = `
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		VALUES ($1, $2);
	`

	if _, err := tx.ExecContext(ctx, qIns, prID, newRevID); err != nil {
		return fmt.Errorf("%s: Exec insert: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: Commit: %w", op, err)
	}

	return nil
}

func (r *PRRepository) FindShortByReviewer(ctx context.Context, userID string)([]pullrequest.PullRequestShort, error){
	const op = "internal.repository.postgres.pr_repo.FindShortByReviewer"

	const q = `
	SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
	FROM pull_requests pr INNER JOIN pull_request_reviewers r 
	ON pr.pull_request_id = r.pull_request_id
	WHERE r.user_id = $1;
	`

	rows , err := r.db.QueryContext(ctx,q,userID)
	if err != nil{
		return nil, fmt.Errorf("%s, QueryContext: %w", op, err)
	}
	defer rows.Close()

	var result []pullrequest.PullRequestShort

	for rows.Next(){
		var s pullrequest.PullRequestShort
		if err := rows.Scan(&s.ID,&s.Name, &s.AuthorID, &s.Status); err != nil{
			return nil, fmt.Errorf("%s, Scan: %w", op, err)
		}
		result = append(result, s)
	}

	if err := rows.Err(); err != nil{
		return nil, fmt.Errorf("%s, rows.Err: %w", op, err)
	}

	return result, nil
}