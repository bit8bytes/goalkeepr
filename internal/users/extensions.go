package users

// SetPassword hashes the plaintext password and stores it in the User
func (u *User) SetPassword(plaintext string) error {
	pw := &Password{}
	if err := pw.Set(plaintext); err != nil {
		return err
	}
	u.PasswordHash = string(pw.Hash)
	return nil
}

// MatchesPassword checks if the provided plaintext matches the stored password hash
func (u *User) MatchesPassword(plaintext string) (bool, error) {
	pw := &Password{Hash: []byte(u.PasswordHash)}
	return pw.Matches(plaintext)
}
