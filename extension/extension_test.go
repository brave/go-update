package extension

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContains(t *testing.T) {
	extension1 := OfferedExtensions[0]
	extension2 := OfferedExtensions[1]
	unknownExtension := extension1
	unknownExtension.ID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	allExtensions := Extensions{extension1, extension2}
	extension, err := allExtensions.Contains(extension1.ID)
	assert.Nil(t, err)
	assert.Equal(t, extension1.ID, extension.ID)

	extension, err = allExtensions.Contains(extension2.ID)
	assert.Nil(t, err)
	assert.Equal(t, extension2.ID, extension.ID)

	extension, err = allExtensions.Contains(unknownExtension.ID)
	assert.NotNil(t, err, "Not found extension should throw an error")
}

func TestCompareVersions(t *testing.T) {
	// 3 component versions match
	assert.Equal(t, 0, CompareVersions("1.1.1", "1.1.1"))
	// 4 component versions match
	assert.Equal(t, 0, CompareVersions("1.1.1.9", "1.1.1.9"))
	// Can detect larger versions for major component
	assert.Equal(t, 1, CompareVersions("2.1.1", "1.1.1"))
	assert.Equal(t, 1, CompareVersions("2.1.1.1", "1.9.9.9"))
	assert.Equal(t, 1, CompareVersions("1.1.1", "0.9.9"))
	assert.Equal(t, -1, CompareVersions("0.1.1", "1.0.0"))
	// Comparing 3 component vs 4 component versions uses only the leading component parts
	assert.Equal(t, -1, CompareVersions("0.1.1.0", "1.0.0"))
	assert.Equal(t, 0, CompareVersions("1.9.0", "1.9.0.9"))
	// Numbers are treated as numbers and not as strings
	assert.Equal(t, 1, CompareVersions("10.1.1", "1.1.1"))
	// Non integers components are treated as 0
	assert.Equal(t, -1, CompareVersions("zugzug.1.1", "1.1.daboo"))
}

func TestFilterForUpdates(t *testing.T) {
	lightThemeExtension, err := OfferedExtensions.Contains("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.Nil(t, err)
	darkThemeExtension, err := OfferedExtensions.Contains("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.Nil(t, err)

	allExtensions := Extensions{lightThemeExtension, OfferedExtensions[1]}

	// No updates when nothing to check
	check := allExtensions.FilterForUpdates(UpdateRequest{})
	assert.Equal(t, 0, len(check))

	olderExtensionCheck1 := lightThemeExtension
	olderExtensionCheck1.Version = "0.1.0"
	outdatedExtensionCheck := UpdateRequest{olderExtensionCheck1}

	check = allExtensions.FilterForUpdates(outdatedExtensionCheck)
	assert.Equal(t, 1, len(check))
	assert.Equal(t, lightThemeExtension.ID, check[0].ID)
	// Check that the newer version,SHA, title are returned
	assert.Equal(t, lightThemeExtension.Version, check[0].Version)
	assert.Equal(t, lightThemeExtension.SHA256, check[0].SHA256)
	assert.Equal(t, lightThemeExtension.Title, check[0].Title)
	// Check that even if a URL is provided, we use the server's URL
	assert.Equal(t, lightThemeExtension.URL, check[0].URL)

	// Newer extensions have no items returned
	newerExtensionCheck := lightThemeExtension
	newerExtensionCheck.Version = "2.1.0"
	check = allExtensions.FilterForUpdates(UpdateRequest{newerExtensionCheck})
	assert.Equal(t, 0, len(check))

	// 2 outdated extensions both get returned from 1 check
	olderExtensionCheck2 := darkThemeExtension
	olderExtensionCheck2.Version = "0.1.0"
	outdatedExtensionsCheck := UpdateRequest{olderExtensionCheck1, olderExtensionCheck2}
	check = allExtensions.FilterForUpdates(outdatedExtensionsCheck)
	assert.Equal(t, 2, len(check))
	assert.Equal(t, olderExtensionCheck1.ID, check[0].ID)
	assert.Equal(t, olderExtensionCheck2.ID, check[1].ID)

	// Outdated extension that's blacklisted doesn't get updates
	allExtensionsBlacklisted := allExtensions
	for i := range allExtensionsBlacklisted {
		allExtensionsBlacklisted[i].Blacklisted = true
	}
	check = allExtensionsBlacklisted.FilterForUpdates(outdatedExtensionCheck)
	assert.Equal(t, 0, len(check))
}
