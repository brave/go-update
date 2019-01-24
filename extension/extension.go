package extension

import (
	"strconv"
	"strings"
)

// Extension represents an extension which is both used in update checks
// and responses.
type Extension struct {
	ID          string
	Version     string
	SHA256      string
	Title       string
	URL         string
	Blacklisted bool
	Status      string
}

// Extensions is type for a slice of Extension.
type Extensions []Extension

// UpdateRequest represents an extension XML request.
type UpdateRequest Extensions

// UpdateResponse represents an extension XML response.
type UpdateResponse Extensions

// WebStoreUpdateResponse represents a webstore XML response.
// There is no symmetric WebStoreUpdateRequest becuase the request is URL query parameters.
type WebStoreUpdateResponse Extensions

// CompareVersions compares 2 versions:
// returns 0 if both versions are the same.
// returns 1 if version1 is more recent.
// returns -1 if version2 is more recent.
func CompareVersions(version1 string, version2 string) int {
	version1Parts := strings.Split(version1, ".")
	version2Parts := strings.Split(version2, ".")

	smallerVersionParts := version2Parts
	if len(version1Parts) < len(version2Parts) {
		smallerVersionParts = version1Parts
	}
	for i := range smallerVersionParts {
		part1, err := strconv.Atoi(version1Parts[i])
		if err != nil {
			part1 = 0
		}
		part2, err := strconv.Atoi(version2Parts[i])
		if err != nil {
			part2 = 0
		}
		if part1 < part2 {
			return -1
		}
		if part2 < part1 {
			return 1
		}
	}
	return 0
}

// LoadExtensionsIntoMap converts a slice of extensions into a map from ID to extension.Extension
func LoadExtensionsIntoMap(extensions *Extensions) map[string]Extension {
	m := make(map[string]Extension)
	for _, extension := range *extensions {
		m[extension.ID] = extension
	}
	return m
}

// FilterForUpdates filters `extensions` down to only the extensions that are being checked,
// and only the ones that we have updates for.
func (updateRequest *UpdateRequest) FilterForUpdates(allExtensionsMap *map[string]Extension) UpdateResponse {
	filteredExtensions := UpdateResponse{}
	for _, extensionBeingChecked := range *updateRequest {
		foundExtension, ok := (*allExtensionsMap)[extensionBeingChecked.ID]
		if ok {
			status := CompareVersions(extensionBeingChecked.Version, foundExtension.Version)
			if !foundExtension.Blacklisted && status <= 0 {
				if status == 0 {
					foundExtension.Status = "noupdate"
				}
				filteredExtensions = append(filteredExtensions, foundExtension)
			}
		}
	}
	return filteredExtensions
}
