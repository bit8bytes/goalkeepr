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

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, locked_until, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	row := s.db.QueryRowContext(ctx, query, email)

	var u User
	var passwordHash string
	err := row.Scan(&u.ID, &u.Email, &passwordHash, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	u.Password.hash = []byte(passwordHash)
	return &u, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, email, password_hash, locked_until, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var u User
	var passwordHash string
	err := row.Scan(&u.ID, &u.Email, &passwordHash, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	u.Password.hash = []byte(passwordHash)
	return &u, nil
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
