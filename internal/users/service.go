package users

import (
	"context"
	"database/sql"

	"github.com/bit8bytes/toolbox/validator"
)

type SignInForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type SignUpForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	RepeatPassword      string `form:"repeat_password"`
	validator.Validator `form:"-"`
}

type UpdateUserForm struct {
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

type Service struct {
	queries *Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		queries: New(db),
	}
}

func (s *Service) Add(ctx context.Context, user *User) (int, error) {
	result, err := s.queries.Create(ctx, CreateParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.queries.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (*User, error) {
	user, err := s.queries.GetByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	result, err := s.queries.Delete(ctx, int64(id))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (f *SignUpForm) Validate() {
	f.Check(validator.NotBlank(f.Email), "email", "This field cannot be blank")
	f.Check(validator.Matches(f.Email, validator.EmailRX), "email", "This field must be a valid email address")

	f.Check(validator.NotBlank(f.Password), "password", "This field cannot be blank")
	f.Check(validator.MinChars(f.Password, 8), "password", "This field must be at least 8 characters long")

	f.Check(validator.NotBlank(f.RepeatPassword), "repeat_password", "This field cannot be blank")
	f.Check(f.Password == f.RepeatPassword, "repeat_password", "Passwords do not match")
}

func (f *SignInForm) Validate() {
	f.Check(validator.NotBlank(f.Email), "email", "This field cannot be blank")
	f.Check(validator.Matches(f.Email, validator.EmailRX), "email", "This field must be a valid email address")

	f.Check(validator.NotBlank(f.Password), "password", "This field cannot be blank")
	f.Check(validator.MinChars(f.Password, 8), "password", "This field must be at least 8 characters long")
}
