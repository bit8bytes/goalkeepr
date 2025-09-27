package share

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type Service struct {
	db *sql.DB
}

type Share struct {
	ID       int    `db:"id"`
	UserID   int    `db:"user_id"`
	PublicID string `db:"public_id"`
}

func New(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Create(ctx context.Context, userID int) (*Share, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM share WHERE user_id = ?", userID).Scan(&count)
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

	result, err := s.db.ExecContext(ctx, "INSERT INTO share (user_id, public_id) VALUES (?, ?)", userID, publicID)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &Share{
		ID:       int(id),
		UserID:   userID,
		PublicID: publicID,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM share WHERE id = ?", id)
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
	rows, err := s.db.QueryContext(ctx, "SELECT id, user_id, public_id FROM share WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shares: %w", err)
	}
	defer rows.Close()

	var shares []Share
	for rows.Next() {
		var share Share
		err := rows.Scan(&share.ID, &share.UserID, &share.PublicID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan share: %w", err)
		}
		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return shares, nil
}

func (s *Service) GetUserIDByPublicID(ctx context.Context, publicID string) (int, error) {
	var userID int
	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM share WHERE public_id = ?", publicID).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID by public ID: %w", err)
	}
	return userID, nil
}

func generatePublicID() (string, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
