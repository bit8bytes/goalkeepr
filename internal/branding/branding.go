package branding

import (
	"context"
	"database/sql"
)

type Branding struct {
	ID          int    `db:"id"`
	UserID      int    `db:"user_id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) GetByUserID(ctx context.Context, userID int) (Branding, error) {
	query := `
		SELECT id, user_id, title, description
		FROM branding
		WHERE user_id = ?
	`

	row := s.db.QueryRowContext(ctx, query, userID)

	var b Branding
	err := row.Scan(&b.ID, &b.UserID, &b.Title, &b.Description)
	if err != nil {
		return Branding{}, err
	}

	return b, nil
}

func (s *Service) CreateOrUpdate(ctx context.Context, branding *Branding) error {
	query := `
		INSERT INTO branding (user_id, title, description)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description
	`

	_, err := s.db.ExecContext(ctx, query, branding.UserID, branding.Title, branding.Description)
	return err
}

func (s *Service) DeleteByUserID(ctx context.Context, userID int) error {
	query := `
		DELETE FROM branding
		WHERE user_id = ?
	`

	result, err := s.db.ExecContext(ctx, query, userID)
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
