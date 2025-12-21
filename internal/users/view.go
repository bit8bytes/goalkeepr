package users

import "time"

type UserView struct {
	ID           int64
	Email        string
	PasswordHash string
	LockedUntil  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) ToView() UserView {
	view := UserView{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    time.Unix(u.CreatedAt, 0),
		UpdatedAt:    time.Unix(u.UpdatedAt, 0),
	}

	if u.LockedUntil.Valid {
		view.LockedUntil = time.Unix(u.LockedUntil.Int64, 0)
	}

	return view
}
