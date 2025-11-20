package page

import (
	"github.com/bit8bytes/goalkeepr/ui/layout"
)

// Page represents an HTML page with its associated layout.
type Page struct {
	path   string
	layout layout.Layout
}

// Name returns the file path to the page template.
func (p Page) Name() string {
	return p.path
}

// Layout returns the layout associated with this page.
func (p Page) Layout() layout.Layout {
	return p.layout
}

// New creates a new Page with the specified path and layout.
func New(path string, l layout.Layout) Page {
	return Page{
		path:   path,
		layout: l,
	}
}

// Predefined pages for the application.
var (
	SignUp            = New("(auth)/signup.html", layout.Auth)
	SignIn            = New("(auth)/signin.html", layout.Auth)
	Goals             = New("goals/index.html", layout.Goals)
	AddGoal           = New("goals/add.html", layout.Goals)
	EditGoal          = New("goals/edit.html", layout.Goals)
	ShareGoals        = New("goals/share.html", layout.Goals)
	Settings          = New("settings/index.html", layout.Settings)
	Share             = New("s/index.html", layout.Share)
	NotFound          = New("(center)/not-found.html", layout.Center)
	Error             = New("(center)/error.html", layout.Center)
	RateLimitExceeded = New("(center)/rate-limit-exceeded.html", layout.Center)
	Landing           = New("(landing)/landing.html", layout.Landing)
	Privacy           = New("(landing)/privacy.html", layout.Landing)
	Imprint           = New("(landing)/imprint.html", layout.Landing)
)

// All returns all predefined pages in the application.
func All() []Page {
	return []Page{
		SignUp, SignIn,
		Goals, AddGoal, EditGoal, ShareGoals,
		Settings,
		Share,
		NotFound, Error, RateLimitExceeded,
		Landing, Privacy, Imprint,
	}
}
