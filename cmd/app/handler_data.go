package main

import (
	"time"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/share"
	successCriteria "github.com/bit8bytes/goalkeepr/internal/success_criteria"
)

// GoalGroup represents a group of goals organized by date.
type GoalGroup struct {
	Date  time.Time
	Goals []goals.View
}

// SharePageData contains data for the public share page.
type SharePageData struct {
	Goals      []goals.View
	GoalGroups []GoalGroup
	Branding   branding.View
}

// GoalsPageData contains data for the user's goals page.
type GoalsPageData struct {
	Goals           []goals.View
	GoalGroups      []GoalGroup
	Branding        branding.View
	Now             time.Time
	GoalDefaultDues map[int64]string
}

// EditGoalPageData contains data for the edit goal page.
type EditGoalPageData struct {
	SuccessCriteria []successCriteria.View
	GoalID          int
}

// ShareGoalsPageData contains data for the share goals management page.
type ShareGoalsPageData struct {
	Links []share.View
	Host  string
}

// ErrorPageData contains data for the error page.
type ErrorPageData struct {
	TraceID string
	Message string
}
