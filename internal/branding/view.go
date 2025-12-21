package branding

type View struct {
	Title       string
	Description string
}

func (b *Branding) ToView() View {
	view := View{}

	if b.Title.Valid {
		view.Title = b.Title.String
	}

	if b.Description.Valid {
		view.Description = b.Description.String
	}

	return view
}
