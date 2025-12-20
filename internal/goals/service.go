package goals

import (
	"context"
	"database/sql"
	"time"

	"github.com/bit8bytes/toolbox/validator"
)

const HTMLDateFormat = "2006-01-02"

type Form struct {
	ID                  int    `form:"id"`
	Goal                string `form:"goal"`
	Due                 string `form:"due"`
	Achieved            bool   `form:"achieved"`
	VisibleToPublic     bool   `form:"visible"`
	validator.Validator `form:"-"`
}

func (f *Form) Validate() {
	f.Check(validator.NotBlank(f.Goal), "goal", "Goal cannot be blank")
	f.Check(validator.MaxChars(f.Goal, 500), "goal", "Goal cannot be more than 500 characters")
	f.Check(validator.NotBlank(f.Due), "due", "Due date cannot be blank")
}

type Service struct {
	queries *Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: New(db),
	}
}

func (s *Service) Add(ctx context.Context, userID int, form *Form) error {
	dueTime, err := time.Parse(HTMLDateFormat, form.Due)
	if err != nil {
		return err
	}

	visibleToPublic := int64(0)
	if form.VisibleToPublic {
		visibleToPublic = 1
	}

	_, err = s.queries.Create(ctx, CreateParams{
		UserID: int64(userID),
		Goal: sql.NullString{
			String: form.Goal,
			Valid:  true,
		},
		Due: sql.NullInt64{
			Int64: dueTime.Unix(),
			Valid: true,
		},
		VisibleToPublic: sql.NullInt64{
			Int64: visibleToPublic,
			Valid: true,
		},
		Achieved: sql.NullInt64{
			Int64: 0,
			Valid: true,
		},
	})
	return err
}

func (s *Service) GetAll(ctx context.Context, userID int) ([]GoalView, error) {
	goals, err := s.queries.GetAll(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	views := make([]GoalView, len(goals))
	for i, goal := range goals {
		views[i] = goal.ToView()
	}

	return views, nil
}

func (s *Service) Get(ctx context.Context, goalID, userID int) (GoalView, error) {
	goal, err := s.queries.Get(ctx, GetParams{
		ID:     int64(goalID),
		UserID: int64(userID),
	})
	if err != nil {
		return GoalView{}, err
	}

	return goal.ToView(), nil
}

func (s *Service) Update(ctx context.Context, goalID, userID int, form *Form) (int, error) {
	dueTime, err := time.Parse(HTMLDateFormat, form.Due)
	if err != nil {
		return 0, err
	}

	visibleToPublic := int64(0)
	if form.VisibleToPublic {
		visibleToPublic = 1
	}
	achieved := int64(0)
	if form.Achieved {
		achieved = 1
	}

	result, err := s.queries.Update(ctx, UpdateParams{
		Goal: sql.NullString{
			String: form.Goal,
			Valid:  true,
		},
		Due: sql.NullInt64{
			Int64: dueTime.Unix(),
			Valid: true,
		},
		VisibleToPublic: sql.NullInt64{
			Int64: visibleToPublic,
			Valid: true,
		},
		Achieved: sql.NullInt64{
			Int64: achieved,
			Valid: true,
		},
		ID:     int64(goalID),
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

func (s *Service) Delete(ctx context.Context, goalID, userID int) (int, error) {
	result, err := s.queries.Delete(ctx, DeleteParams{
		ID:     int64(goalID),
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

func (s *Service) GetAllShared(ctx context.Context, userID int) ([]GoalView, error) {
	goals, err := s.queries.GetAllShared(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	views := make([]GoalView, len(goals))
	for i, goal := range goals {
		views[i] = goal.ToView()
	}

	return views, nil
}
