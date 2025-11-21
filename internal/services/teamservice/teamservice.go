package teamservice

import (
	"context"
	"errors"
	"math/rand"

	"github.com/hihikaAAa/PRManager/internal/domain/team"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
)

type TeamService struct{
	userRepo *postgres.UserRepository
	teamRepo *postgres.TeamRepository
	prRepo   *postgres.PRRepository
}

type DeactivateResult struct {
	TeamName string `json:"team_name"`
	Deactivated []string `json:"deactivated"`
	ReassignedCount int `json:"reassigned_count"`
	RemovedCount int `json:"removed_count"`
}


func New(userRepo *postgres.UserRepository, teamRepo *postgres.TeamRepository, prRepo  *postgres.PRRepository) *TeamService{
	return &TeamService{ userRepo: userRepo, teamRepo: teamRepo, prRepo: prRepo}
}

func (ts *TeamService) AddTeam(ctx context.Context, teamName string, members []*user.User) error{
	exists, err := ts.teamRepo.Exists(ctx,teamName)
	if err != nil{
		return err
	}
	if exists{
		return serviceerrors.ErrTeamExists
	}

	err = ts.teamRepo.CreateTeam(ctx,teamName)
	if err != nil {
		return err
	}
	
	err = ts.userRepo.UpsertManyForTeam(ctx,teamName,members)
	if err != nil{
		return err
	}
	return nil
}

func (ts *TeamService) GetTeam(ctx context.Context, teamName string)(*team.Team, error){
	team, err := ts.teamRepo.GetWithMembers(ctx,teamName)
	if err != nil{
		return nil, err
	}
	return team, nil
}

func (ts *TeamService) DeactivateAndReassign(ctx context.Context, teamName string, userIDs []string) (DeactivateResult, error) {
	res := DeactivateResult{TeamName: teamName}
	if len(userIDs) == 0 {
		return res, nil
	}
	exists, err := ts.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return res, err
	}
	if !exists {
		return res, serviceerrors.ErrTeamNotFound
	}
	deactivatedSet := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		deactivatedSet[id] = struct{}{}
	}
	for _, uid := range userIDs {
		u, err := ts.userRepo.GetByID(ctx, uid)
		if err != nil {
			if errors.Is(err, repo_errors.ErrUserNotFound) {
				return res, serviceerrors.ErrUserNotFound
			}
			return res, err
		}
		if u.TeamName != teamName {
			continue
		}
		if _, err := ts.userRepo.SetIsActive(ctx, uid, false); err != nil {
			return res, err
		}
		res.Deactivated = append(res.Deactivated, uid)
		prIDs, err := ts.prRepo.GetOpenPRIDsByReviewer(ctx, uid)
		if err != nil {
			return res, err
		}
		for _, prID := range prIDs {
			pr, err := ts.prRepo.GetWithReviewers(ctx, prID)
			if err != nil {
				return res, err
			}
			if pr.Status == pullrequest.StatusMerged {
				continue
			}
			exclude := make([]string, 0, len(pr.Reviewers)+1+len(userIDs))
			exclude = append(exclude, pr.AuthorID)
			exclude = append(exclude, pr.Reviewers...)
			for id := range deactivatedSet {
				exclude = append(exclude, id)
			}
			candidates, err := ts.userRepo.FindActiveByTeamExcept(ctx, teamName, exclude)
			if err != nil {
				return res, err
			}
			if len(candidates) == 0 {
				if err := ts.prRepo.RemoveReviewer(ctx, prID, uid); err != nil {
					return res, err
				}
				res.RemovedCount++
				continue
			}
			newUser := candidates[rand.Intn(len(candidates))]
			if err := ts.prRepo.ReplaceReviewers(ctx, prID, uid, newUser.ID); err != nil {
				return res, err
			}
			res.ReassignedCount++
		}
	}
	return res, nil
}
