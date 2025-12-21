package share

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type Service struct {
	queries *Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: New(db),
	}
}

func (s *Service) Create(ctx context.Context, userID int) (*Share, error) {
	count, err := s.queries.CountByUserID(ctx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to count existing shares: %w", err)
	}

	if count >= 10 {
		return nil, fmt.Errorf("share limit reached: maximum 10 shares per user")
	}

	publicID, err := generatePublicID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate public ID: %w", err)
	}

	result, err := s.queries.Create(ctx, CreateParams{
		UserID:   int64(userID),
		PublicID: publicID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &Share{
		ID:       id,
		UserID:   int64(userID),
		PublicID: publicID,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	result, err := s.queries.Delete(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete share: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("share not found")
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context, userID int) ([]Share, error) {
	shares, err := s.queries.GetAll(ctx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to query shares: %w", err)
	}

	return shares, nil
}

func (s *Service) GetUserIDByPublicID(ctx context.Context, publicID string) (int, error) {
	userID, err := s.queries.GetUserIDByPublicID(ctx, publicID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID by public ID: %w", err)
	}
	return int(userID), nil
}

// generatePublicID generates a crypto random string. On error, it returns an emtry string.
func generatePublicID() (string, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
