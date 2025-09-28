package users

import "golang.org/x/crypto/bcrypt"

const (
	PreComputedHash = "$2a$12$BcH.9w4Gmo4Nz1WaHgzjDOGHkQiUWMv6MjsPqfEogPKJODCNJd2rS"
)

type Password struct {
	plaintext *string
	Hash      []byte
}

func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.Hash = hash

	return nil
}

func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
