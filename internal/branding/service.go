package branding

import (
	"context"
	"database/sql"

	"github.com/bit8bytes/toolbox/validator"
)

type Form struct {
	Title               string `form:"title"`
	Description         string `form:"description"`
	validator.Validator `form:"-"`
}

type Service struct {
	queries *Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: New(db),
	}
}

func (s *Service) GetByUserID(ctx context.Context, userID int) (*Branding, error) {
	branding, err := s.queries.GetByUserID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	return &branding, nil
}

func (s *Service) CreateOrUpdate(ctx context.Context, userID int, title, description string) error {
	_, err := s.queries.CreateOrUpdate(ctx, CreateOrUpdateParams{
		UserID:      int64(userID),
		Title:       sql.NullString{String: title, Valid: title != ""},
		Description: sql.NullString{String: description, Valid: description != ""},
	})
	return err
}

func (s *Service) DeleteByUserID(ctx context.Context, userID int) error {
	result, err := s.queries.Delete(ctx, int64(userID))
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
