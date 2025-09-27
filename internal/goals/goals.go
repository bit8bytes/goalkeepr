package goals

import (
	"context"
	"database/sql"
	"time"
)

type Goal struct {
	ID              int
	UserID          int
	Goal            string
	Due             time.Time
	VisibleToPublic bool
	Achieved        bool
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Add(ctx context.Context, goal *Goal) error {
	stmt := `INSERT INTO goals (user_id, goal, due, visible_to_public, achieved)
				VALUES(?, ?, ?, ?, ?) RETURNING id`

	var id int
	args := []any{goal.UserID, goal.Goal, goal.Due, goal.VisibleToPublic, goal.Achieved}
	err := s.db.QueryRowContext(ctx, stmt, args...).Scan(&id)
	if err != nil {
		return err
	}

	goal.ID = id
	return nil
}
