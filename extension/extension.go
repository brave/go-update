package extension

import "errors"

// Extension represents an extension which is both used in update checks
// and responses.
type Extension struct {
	ID      string
	Version string
	SHA256  string
	Title   string
	URL     string
}

// Contains checks if the specified extension is contained in the extensions list
func Contains(extensions Extensions, extensionID string) (Extension, error) {
	var foundExtension Extension
	for _, extension := range extensions {
		if extension.ID == extensionID {
			foundExtension = extension
			return foundExtension, nil
		}
	}
	return foundExtension, errors.New("no extensions found")
}

// Extensions is an array of Extension.
type Extensions []Extension
