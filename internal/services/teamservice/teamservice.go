package teamservice

import (
	"context"

	"github.com/hihikaAAa/PRManager/internal/domain/team"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type TeamService struct{
	userRepo *postgres.UserRepository
	teamRepo *postgres.TeamRepository
}

func New(userRepo *postgres.UserRepository, teamRepo *postgres.TeamRepository) *TeamService{
	return &TeamService{ userRepo: userRepo, teamRepo: teamRepo}
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
