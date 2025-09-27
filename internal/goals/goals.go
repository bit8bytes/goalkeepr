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

func (s *Service) GetAll(ctx context.Context, userID int) ([]Goal, error) {
	stmt := `SELECT id, user_id, goal, due, visible_to_public, achieved FROM goals WHERE user_id = ? ORDER BY due ASC`

	rows, err := s.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		var goal Goal
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.Goal, &goal.Due, &goal.VisibleToPublic, &goal.Achieved)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}

func (s *Service) Get(ctx context.Context, goalID, userID int) (Goal, error) {
	var goal Goal
	stmt := `SELECT id, user_id, goal, due, visible_to_public, achieved 
			 FROM goals 
			 WHERE id = ? AND user_id = ?`

	err := s.db.QueryRowContext(ctx, stmt, goalID, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Goal,
		&goal.Due,
		&goal.VisibleToPublic,
		&goal.Achieved,
	)

	if err != nil {
		return Goal{}, err
	}

	return goal, nil
}

func (s *Service) Update(ctx context.Context, goal *Goal) (int, error) {
	stmt := `UPDATE goals 
			 SET goal = ?, due = ?, visible_to_public = ?, achieved = ?
			 WHERE id = ? AND user_id = ?`

	result, err := s.db.ExecContext(ctx, stmt, goal.Goal, goal.Due, goal.VisibleToPublic, goal.Achieved, goal.ID, goal.UserID)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func (s *Service) Delete(ctx context.Context, goalID, userID int) (int, error) {
	stmt := `DELETE FROM goals WHERE id = ? AND user_id = ?`

	result, err := s.db.ExecContext(ctx, stmt, goalID, userID)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
