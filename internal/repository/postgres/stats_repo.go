package postgres

import (
	"context"
	"fmt"

	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
)

type PRStats struct{
	TotalPR int
	OpenPR int
	MergedPR int
}

type ReviewerStat struct{
	UserID string
	Count int
}

func (r *PRRepository) GetStats(ctx context.Context) (PRStats, []ReviewerStat, error){
	const op = "internal.repository.postgres.stats_repo.GetStats"

	stats := PRStats{}

	const qStatus = `
	SELECT status, COUNT(*)
	FROM pull_requests 
	GROUP BY status;
	`

	rows, err := r.db.QueryContext(ctx,qStatus)
	if err != nil{
		return stats, nil, fmt.Errorf("%s, QueryContext status: %w", op, err)
	}

	defer rows.Close()

	for rows.Next(){
		var status pullrequest.Status
		var cnt int
		if err := rows.Scan(&status, &cnt); err != nil{
			return stats, nil, fmt.Errorf("%s, Scan status: %w", op, err)
		}
		stats.TotalPR += cnt
		switch status{
			case pullrequest.StatusOpen:
				stats.OpenPR = cnt
			case pullrequest.StatusMerged:
				stats.MergedPR = cnt
		}
	}

	if err := rows.Err(); err != nil{
		return stats, nil, fmt.Errorf("%s, rowsErr status: %w", op, err)
	}

	const qReviewers = `
	SELECT user_id, COUNT(*)
	FROM pull_request_reviewers
	GROUP BY user_id
	`

	rRows, err := r.db.QueryContext(ctx,qReviewers)
	if err != nil{
		return stats, nil, fmt.Errorf("%s, QueryContext reviewers: %w", op, err)
	}

	defer rRows.Close()

	var reviewers []ReviewerStat
	for rRows.Next(){
		var s ReviewerStat
		if err := rRows.Scan(&s.UserID, &s.Count); err != nil{
			return stats, nil, fmt.Errorf("%s, Scan stats: %w", op, err)
		}
		reviewers = append(reviewers, s)
	}

	if err := rRows.Err(); err != nil{
		return stats, nil, fmt.Errorf("%s, rowsErr stats: %w", op, err)
	}

	return stats, reviewers, nil
}