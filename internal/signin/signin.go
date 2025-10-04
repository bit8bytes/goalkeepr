package signin

import "github.com/bit8bytes/toolbox/validator"

type Form struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	Website             string `form:"website"`
	validator.Validator `form:"-"`
}
