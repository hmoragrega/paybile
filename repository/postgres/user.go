package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hmoragrega/paybile/service"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (u service.User, err error) {
	query := `SELECT id, hashed_password FROM users WHERE login = $1`
	row := r.DB.QueryRowContext(ctx, query, login)
	if err := row.Scan(&u.ID, &u.HashedPassword); err != nil {
		if err == sql.ErrNoRows {
			return u, service.ErrUserNotFound
		}

		return u, fmt.Errorf("%w: %v", fmt.Errorf("cannot query user by login"), err)
	}

	u.Login = login

	return u, nil
}
