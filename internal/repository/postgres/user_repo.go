package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hihikaAAa/PRManager/internal/domain/user"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type UserRepository struct{
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository{
	return &UserRepository{db: db}
}

func (r *UserRepository) UpsertManyForTeam(ctx context.Context,teamName string, users []*user.User) error{
	const op = "internal.repository.postgres.user_repo.UpsertManyForTeam"

	tx, err := r.db.BeginTx(ctx,nil)
	if err != nil{
		return fmt.Errorf("%s, BeginTx: %w", op,err)
	}
	defer tx.Rollback()

	const q = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = now();
	`
	stmt, err := tx.PrepareContext(ctx,q)
	if err != nil{
		return fmt.Errorf("%s, PrepareContext: %w", op,err)
	}
	defer stmt.Close()

	for _, u := range users{
		if _, err := stmt.ExecContext(ctx, u.ID, u.Name, teamName,u.IsActive); err!= nil{
			return fmt.Errorf("%s, ExecContext: %w", op, err)
		}	
	}
	if err := tx.Commit(); err != nil{
		return fmt.Errorf("%s, Commit: %w", op, err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string)(*user.User, error){
	const op = "internal.repository.postgres.user_repo.GetByID"

	const q = `
	SELECT user_id, username, team_name, is_active
	FROM users
	WHERE user_id = $1;
	`

	u := &user.User{}
	err := r.db.QueryRowContext(ctx,q,id).Scan(&u.ID,&u.Name,&u.TeamName,&u.IsActive)
	if err == sql.ErrNoRows{
		return nil, fmt.Errorf("%s: %w",op,repo_errors.ErrUserNotFound)
	}
	if err !=nil{
		return nil, fmt.Errorf("%s, QueryRow: %w", op, err)
	}
	return u, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, id string, active bool)(*user.User, error){
	const op = "internal.repository.postgres.user_repo.SetIsActive"

	const q = `
		UPDATE users
		SET is_active = $2, updated_at = now()
		WHERE user_id = $1
		RETURNING user_id,username,team_name,is_active;
	`

	u := &user.User{}
	err  := r.db.QueryRowContext(ctx,q,id,active).Scan(&u.ID,&u.Name,&u.TeamName,&u.IsActive)
	if err == sql.ErrNoRows{
		return nil, fmt.Errorf("%s: %w",op,repo_errors.ErrUserNotFound)
	}
	if err !=nil{
		return nil, fmt.Errorf("%s, QueryRow: %w", op, err)
	}
	return u, nil
}

func (r *UserRepository) FindActiveByTeamExcept(ctx context.Context, teamName string, excluded []string ) ([]*user.User,error){
	const op = "internal.repository.postgres.user_repo.FindActiveByTeamExceptAuthor"

	const q = `
	SELECT user_id, username, team_name, is_active
	FROM users
	WHERE team_name = $1 AND is_active = $2
	`

	rows, err := r.db.QueryContext(ctx, q, teamName, true)
	if err != nil{
		return nil, fmt.Errorf("%s, QueryContext: %w", op, err)
	}
	defer rows.Close()

	cands := make([]*user.User,0)
	for rows.Next(){
		user := &user.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.TeamName, &user.IsActive)
		if err != nil{
			return nil, fmt.Errorf("%s, Scan: %w", op, err)
		}
		if checkIfExcluded(user.ID, excluded){
			continue
		}
		cands = append(cands, user)
	}

	if err := rows.Err(); err != nil{
		return nil, fmt.Errorf("%s, rows.Err: %w", op,err)
	}

	return cands, nil
}

func checkIfExcluded(ID string, excluded []string) bool{
	for _, elem := range excluded{
		if elem == ID{
			return true
		}
	}
	return false
}