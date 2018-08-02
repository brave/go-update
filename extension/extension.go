package extension

// Extension represents an extension which is both used in update checks
// and responses.
type Extension struct {
	ID      string
	Version string
	SHA256  string
	Title   string
	URL     string
}

// Extensions is an array of Extension.
type Extensions []Extension
