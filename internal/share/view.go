package share

type View struct {
	ID       int64
	UserID   int64
	PublicID string
}

func (s *Share) ToView() View {
	v := View{
		ID:       s.ID,
		PublicID: s.PublicID,
		UserID:   s.UserID,
	}

	return v

}
