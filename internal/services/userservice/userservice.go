package userservice

import(
	"context"
	
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
)
type UserService struct{
	userRepo *postgres.UserRepository
	prRepo *postgres.PRRepository
}

func New(prRepo *postgres.PRRepository, userRepo *postgres.UserRepository)*UserService{
	return &UserService{prRepo : prRepo, userRepo: userRepo}
}

func (u *UserService) SetIsActive(ctx context.Context,userID string, isActive bool) (*user.User, error){
	user, err := u.userRepo.SetIsActive(ctx,userID,isActive)
	return user,err
}

func (u *UserService) GetReviewPRs(ctx context.Context, userID string)([]pullrequest.PullRequestShort, error){
	if _, err := u.userRepo.GetByID(ctx, userID); err != nil{
		return nil, err
	}

	prs, err := u.prRepo.FindShortByReviewer(ctx, userID)
	if err != nil{
		return nil, err
	}

	return prs, nil
}
