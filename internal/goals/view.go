package goals

import "time"

// GoalView is the render-ready version of Goal
type GoalView struct {
	ID              int64
	UserID          int64
	Goal            string
	Year            string
	Due             time.Time
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
		view.Due = dueTime
	}

	return view
}
