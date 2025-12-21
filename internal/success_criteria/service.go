package success_criteria

import (
	"context"
	"database/sql"
	"time"

	"github.com/bit8bytes/toolbox/validator"
)

type Form struct {
	ID                  int    `form:"id"`
	GoalID              int    `form:"goal_id"`
	Description         string `form:"description"`
	Completed           bool   `form:"completed"`
	Position            int    `form:"position"`
	validator.Validator `form:"-"`
}

func (f *Form) Validate() {
	f.Check(validator.NotBlank(f.Description), "description", "Description cannot be blank")
	f.Check(validator.MaxChars(f.Description, 500), "description", "Description cannot be more than 500 characters")
}

type Service struct {
	queries *Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: New(db),
	}
}

func (s *Service) Add(ctx context.Context, goalID, userID int, form *Form) error {
	completed := int64(0)
	if form.Completed {
		completed = 1
	}

	position := sql.NullInt64{}
	if form.Position > 0 {
		position = sql.NullInt64{
			Int64: int64(form.Position),
			Valid: true,
		}
	}

	_, err := s.queries.CreateSuccessCriteria(ctx, CreateSuccessCriteriaParams{
		GoalID:      int64(goalID),
		UserID:      int64(userID),
		Description: form.Description,
		Completed: sql.NullInt64{
			Int64: completed,
			Valid: true,
		},
		Position:  position,
		CreatedAt: time.Now().Unix(),
	})
	return err
}

func (s *Service) GetAllByGoal(ctx context.Context, goalID, userID int) ([]SuccessCriterium, error) {
	criteria, err := s.queries.GetAllSuccessCriteriaByGoal(ctx, GetAllSuccessCriteriaByGoalParams{
		GoalID: int64(goalID),
		UserID: int64(userID),
	})
	if err != nil {
		return nil, err
	}

	return criteria, nil
}

func (s *Service) Get(ctx context.Context, criteriaID, userID int) (SuccessCriterium, error) {
	criteria, err := s.queries.GetSuccessCriteria(ctx, GetSuccessCriteriaParams{
		ID:     int64(criteriaID),
		UserID: int64(userID),
	})
	if err != nil {
		return SuccessCriterium{}, err
	}

	return criteria, nil
}

func (s *Service) Update(ctx context.Context, criteriaID, userID int, form *Form) (int, error) {
	completed := int64(0)
	if form.Completed {
		completed = 1
	}

	position := sql.NullInt64{}
	if form.Position > 0 {
		position = sql.NullInt64{
			Int64: int64(form.Position),
			Valid: true,
		}
	}

	result, err := s.queries.UpdateSuccessCriteria(ctx, UpdateSuccessCriteriaParams{
		Description: form.Description,
		Completed: sql.NullInt64{
			Int64: completed,
			Valid: true,
		},
		Position: position,
		ID:       int64(criteriaID),
		UserID:   int64(userID),
	})
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func (s *Service) Toggle(ctx context.Context, criteriaID, userID int) (int, error) {
	result, err := s.queries.ToggleSuccessCriteriaCompleted(ctx, ToggleSuccessCriteriaCompletedParams{
		ID:     int64(criteriaID),
		UserID: int64(userID),
	})
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func (s *Service) Delete(ctx context.Context, criteriaID, userID int) (int, error) {
	result, err := s.queries.DeleteSuccessCriteria(ctx, DeleteSuccessCriteriaParams{
		ID:     int64(criteriaID),
		UserID: int64(userID),
	})
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func (s *Service) DeleteAllByGoal(ctx context.Context, goalID, userID int) (int, error) {
	result, err := s.queries.DeleteAllSuccessCriteriaByGoal(ctx, DeleteAllSuccessCriteriaByGoalParams{
		GoalID: int64(goalID),
		UserID: int64(userID),
	})
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
