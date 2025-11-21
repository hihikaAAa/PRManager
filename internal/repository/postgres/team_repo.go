package postgres

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/hihikaAAa/PRManager/internal/domain/team"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
    "github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type TeamRepository struct {
    db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository{
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, name string) error{
	const op = "internal.repository.postgres.team_repo.CreateTeam"

	const q = `INSERT INTO teams (team_name) VALUES($1)`

	_, err := r.db.ExecContext(ctx,q,name)
	if err != nil{
		return fmt.Errorf("%s, ExecContext: %w", op, err)
	}
	return nil
}

func (r *TeamRepository) Exists(ctx context.Context, name string) (bool, error){
	const op = "internal.repository.postgres.team_repo.Exists"

	const q = `
	SELECT 1 FROM teams WHERE team_name = $1
	`
	var dummy int
	err := r.db.QueryRowContext(ctx,q,name).Scan(&dummy)
	if err == sql.ErrNoRows{
		return false, nil
	}
	if err != nil{
		return false, fmt.Errorf("%s, QueryRow: %w", op,err)
	}

	return true, nil
}

func (r *TeamRepository) GetWithMembers(ctx context.Context, name string)(*team.Team, error){
	const op = "internal.repository.postgres.team_repo.GetWithMembers"

	const qTeam = `
	SELECT team_name FROM teams WHERE team_name = $1
	`
	t := &team.Team{}
	if err := r.db.QueryRowContext(ctx, qTeam, name).Scan(&t.TeamName); err != nil{
		if err == sql.ErrNoRows{
			return nil, fmt.Errorf("%s: %w", op, repo_errors.ErrTeamNotFound)
		}
		return nil, fmt.Errorf("%s, QueryRow: %w", op, err)
	}

	const qMembers = `
	SELECT user_id, username, team_name, is_active
	FROM users
	WHERE team_name = $1
	`

	rows, err := r.db.QueryContext(ctx,qMembers,name)
	if err != nil{
		return nil, fmt.Errorf("%s, QueryContext: %w", op, err)
	}
	defer rows.Close()

	for rows.Next(){
		u := &user.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.TeamName, &u.IsActive); err != nil{
			return nil, fmt.Errorf("%s, Scan: %w", op ,err)
		}
		t.Members = append(t.Members, u)
	}

	if err := rows.Err(); err != nil{
		return nil, fmt.Errorf("%s, rows.Err: %w", op,err)
	}

	return t, nil
}