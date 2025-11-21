package prservice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type PRService struct{
	prRepo *postgres.PRRepository
	userRepo *postgres.UserRepository
}

func New(prRepo *postgres.PRRepository, userRepo *postgres.UserRepository) *PRService{
	return &PRService{prRepo: prRepo, userRepo: userRepo}
}

func (s *PRService) Create(ctx context.Context, id,name,authorID string)(*pullrequest.PullRequest, error){
	const op = "internal.services.prservice.Create"

	if _, err := s.prRepo.GetWithReviewers(ctx,id); err == nil{
		return nil, serviceerrors.ErrPRExists
	} else if !errors.Is(err, repo_errors.ErrPRNotFound){
		return nil, fmt.Errorf("%s: %w", op , err)
	}

	author, err := s.userRepo.GetByID(ctx,authorID); 
	if err != nil{
		if errors.Is(err, repo_errors.ErrUserNotFound){
			return nil, serviceerrors.ErrUserNotFound
		}
		return nil, err
	}
	excluded := []string{authorID}
	candidates, err := s.userRepo.FindActiveByTeamExcept(ctx, author.TeamName, excluded)
	if err != nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	reviewers := pickRandomReviewers(candidates,2)
	now := time.Now()
	pr := pullrequest.PullRequest{
		ID : id, Name: name, AuthorID: authorID, Status: pullrequest.StatusOpen, Reviewers: reviewers, CreatedAt: now, MergedAt: nil,
	}

	if err := s.prRepo.CreateWithReviewers(ctx, pr); err != nil{
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) Merge(ctx context.Context, id string)(*pullrequest.PullRequest, error){
	now := time.Now().UTC()
	return s.prRepo.Merge(ctx,id,now)
}

func (s *PRService) Reassign(ctx context.Context, prID, oldReviewerID string)(*pullrequest.PullRequest, string, error){
	pr, err := s.prRepo.GetWithReviewers(ctx, prID)
	if err != nil{
		return nil, "", err
	}

	if pr.Status == pullrequest.StatusMerged{
		return nil, "", serviceerrors.ErrPRMerged
	}
	
	found := false
	for _, rev := range pr.Reviewers{
		if rev == oldReviewerID{
			found = true
			break
		}
	}
	if !found{
		return nil, "", serviceerrors.ErrReviewerNotFound
	}

	oldUser, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil{
		return nil, "", err
	}
	exclude := make([]string,0,len(pr.Reviewers)+1)
	exclude = append(exclude, pr.AuthorID)
	exclude = append(exclude, pr.Reviewers...)

	candidates , err := s.userRepo.FindActiveByTeamExcept(ctx,oldUser.TeamName,exclude)
	if err != nil{
		return nil, "", err
	}
	if len(candidates) == 0{
		return nil, "", serviceerrors.ErrNoCandidates
	}
	newUserID := candidates[rand.Intn(len(candidates))].ID
	if err := s.prRepo.ReplaceReviewers(ctx,prID,oldReviewerID,newUserID); err != nil{
		return nil, "", err
	}

	updatedPR , err := s.prRepo.GetWithReviewers(ctx,prID)
	if err != nil{
		return nil, "", err
	}
	return updatedPR, newUserID, nil
}

func pickRandomReviewers(available []*user.User,limit int)[]string{
	if len(available) == 0{
		return nil
	}

	if len(available) <= limit{
		out := make([]string,0,len(available))
		for _, u := range available{
			out = append(out, u.ID)
		}
		return out
	}

	tmp := make([]*user.User,len(available))
	copy(tmp,available)

	for i := len(tmp)-1; i>0; i--{
		j := rand.Intn(i+1)
		tmp[i],tmp[j] = tmp[j],tmp[i]
	}

	out := make([]string,0,limit)
	for i := 0; i< limit; i++{
		out = append(out, tmp[i].ID)
	}
	return out
}