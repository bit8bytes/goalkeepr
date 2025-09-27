package users

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID          int        `db:"id"`
	Email       string     `db:"email"`
	Password    password   `db:"-"`
	LockedUntil *time.Time `db:"locked_until"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Add(ctx context.Context, user *User) (int, error) {
	query := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, user.Email, user.Password.hash)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *Service) Get(ctx context.Context, user *User) (*User, error) {
	query := `
		SELECT id, email, password_hash, locked_until, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	row := s.db.QueryRowContext(ctx, query, user.Email)

	var u User
	var passwordHash string
	err := row.Scan(&u.ID, &u.Email, &passwordHash, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	u.Password.hash = []byte(passwordHash)
	return &u, nil
}
