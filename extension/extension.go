package extension

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
)

type PatchInfo struct {
	Hashdiff string `json:"hashdiff"`
	Namediff string `json:"namediff"`
	Sizediff int    `json:"sizediff"`
}

// Extension represents an extension which is both used in update checks
// and responses.
type Extension struct {
	ID          string
	FP          string
	Version     string
	SHA256      string
	Title       string
	URL         string
	Blacklisted bool
	Status      string
	PatchList   map[string]*PatchInfo
}

// Extensions is type for a slice of Extension.
type Extensions []Extension

// ExtensionsMap is safe for use across goroutines.
type ExtensionsMap struct {
	sync.RWMutex
	data map[string]Extension
}

// UpdateRequest represents an extension XML request.
type UpdateRequest Extensions

// UpdateResponse represents an extension XML response.
type UpdateResponse Extensions

// WebStoreUpdateResponse represents a webstore XML response.
// There is no symmetric WebStoreUpdateRequest because the request is URL query parameters.
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

// FilterForUpdates filters `extensions` down to only the extensions that are being checked,
// and only the ones that we have updates for.
func (updateRequest *UpdateRequest) FilterForUpdates(allExtensionsMap *ExtensionsMap) UpdateResponse {
	filteredExtensions := UpdateResponse{}
	allExtensionsMap.RLock()
	defer allExtensionsMap.RUnlock()
	for _, extensionBeingChecked := range *updateRequest {
		foundExtension, ok := allExtensionsMap.data[extensionBeingChecked.ID]
		if ok {
			status := CompareVersions(extensionBeingChecked.Version, foundExtension.Version)
			if !foundExtension.Blacklisted && status <= 0 {
				if status == 0 {
					foundExtension.Status = "noupdate"
				}

				foundExtension.FP = extensionBeingChecked.FP

				filteredExtensions = append(filteredExtensions, foundExtension)
			}
		}
	}
	return filteredExtensions
}

// Store adds or overwrites the key in the map with the Extension
func (m *ExtensionsMap) Store(key string, extension Extension) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = extension
}

// Load looks up the Extension in the map by it's key
func (m *ExtensionsMap) Load(key string) (extension Extension, ok bool) {
	m.RLock()
	defer m.RUnlock()
	extension, ok = m.data[key]
	return
}

// StoreExtensions converts a slice of extensions into a map from ID to extension.Extension
func (m *ExtensionsMap) StoreExtensions(extensions *Extensions) {
	m.Lock()
	defer m.Unlock()
	for _, extension := range *extensions {
		m.data[extension.ID] = extension
	}
}

// MarshalJSON marshals the Extension map into a JSON byte slice
func (m *ExtensionsMap) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	return json.Marshal(m.data)
}

// NewExtensionMap creates a new map of Extension structs where access is controlled by a RW mutex
func NewExtensionMap() *ExtensionsMap {
	return &ExtensionsMap{
		data: make(map[string]Extension),
	}
}
