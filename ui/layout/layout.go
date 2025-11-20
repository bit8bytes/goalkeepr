package layout

// Layout represents an HTML layout template with optional partials.
type Layout struct {
	path     string
	partials []string
}

// Name returns the file path to the layout template.
func (l Layout) Name() string {
	return l.path
}

// Partials returns the file paths to partial templates.
func (l Layout) Partials() []string {
	return l.partials
}

// New creates a new Layout with the specified path and optional partial paths.
func New(path string, partials ...string) Layout {
	return Layout{
		path:     path,
		partials: partials,
	}
}

// Predefined layouts for the application.
var (
	Center   = New("(center)/layout.html")
	Auth     = New("(auth)/layout.html")
	Goals    = New("#shared/app/layout.html", "#shared/app/+partials/*.html")
	Settings = New("#shared/app/layout.html", "#shared/app/+partials/*.html")
	Share    = New("#shared/public/layout.html", "#shared/public/+partials/*.html")
	Landing  = New("#shared/public/layout.html", "#shared/public/+partials/*.html")
)
