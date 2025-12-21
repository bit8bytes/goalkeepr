package success_criteria

import "time"

type View struct {
	ID          int
	GoalID      int
	UserID      int
	Description string
	Completed   bool
	Position    int
	CreatedAt   time.Time
}

func (s SuccessCriterium) ToView() View {
	completed := false
	if s.Completed.Valid && s.Completed.Int64 == 1 {
		completed = true
	}

	position := 0
	if s.Position.Valid {
		position = int(s.Position.Int64)
	}

	return View{
		ID:          int(s.ID),
		GoalID:      int(s.GoalID),
		UserID:      int(s.UserID),
		Description: s.Description,
		Completed:   completed,
		Position:    position,
		CreatedAt:   time.Unix(s.CreatedAt, 0),
	}
}
