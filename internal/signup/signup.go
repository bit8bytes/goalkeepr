package signup

import "github.com/bit8bytes/toolbox/validator"

type Form struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	RepeatPassword      string `form:"repeat_password"`
	validator.Validator `form:"-"`
}
