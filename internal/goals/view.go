package goals

import "time"

// GoalView is the render-ready version of Goal
type GoalView struct {
	ID              int64
	UserID          int64
	Goal            string
	Year            string
	DueDate         string // formatted as HTML date (YYYY-MM-DD)
	VisibleToPublic bool
	Achieved        bool
}

// ToView converts Goal to GoalView
func (g *Goal) ToView() GoalView {
	view := GoalView{
		ID:              g.ID,
		UserID:          g.UserID,
		Goal:            g.Goal.String,
		VisibleToPublic: g.VisibleToPublic.Int64 == 1,
		Achieved:        g.Achieved.Int64 == 1,
	}

	if g.Due.Valid {
		dueTime := time.Unix(g.Due.Int64, 0)
		view.Year = dueTime.Format("2006")
		view.DueDate = dueTime.Format("2006-01-02")
	}

	return view
}
