package extension

import (
	"github.com/brave/go-update/extension/extensiontest"
	"testing"
)

func TestContains(t *testing.T) {
	extension1 := OfferedExtensions[0]
	extension2 := OfferedExtensions[1]
	unknownExtension := extension1
	unknownExtension.ID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	allExtensions := []Extension{extension1, extension2}
	extension, err := Contains(allExtensions, extension1.ID)
	extensiontest.AssertEqual(t, err, nil)
	extensiontest.AssertEqual(t, extension.ID, extension1.ID)

	extension, err = Contains(allExtensions, extension2.ID)
	extensiontest.AssertEqual(t, err, nil)
	extensiontest.AssertEqual(t, extension.ID, extension2.ID)

	extension, err = Contains(allExtensions, unknownExtension.ID)
	if err == nil {
		t.Fatalf("Not found extension should throw an error")
	}
}

func TestCompareVersions(t *testing.T) {
	// 3 component versions match
	extensiontest.AssertEqual(t, CompareVersions("1.1.1", "1.1.1"), 0)
	// 4 component versions match
	extensiontest.AssertEqual(t, CompareVersions("1.1.1.9", "1.1.1.9"), 0)
	// Can detect larger versions for major component
	extensiontest.AssertEqual(t, CompareVersions("2.1.1", "1.1.1"), 1)
	extensiontest.AssertEqual(t, CompareVersions("2.1.1.1", "1.9.9.9"), 1)
	extensiontest.AssertEqual(t, CompareVersions("1.1.1", "0.9.9"), 1)
	extensiontest.AssertEqual(t, CompareVersions("0.1.1", "1.0.0"), -1)
	// Comparing 3 component vs 4 component versions uses only the leading component parts
	extensiontest.AssertEqual(t, CompareVersions("0.1.1.0", "1.0.0"), -1)
	extensiontest.AssertEqual(t, CompareVersions("1.9.0", "1.9.0.9"), 0)
	// Numbers are treated as numbers and not as strings
	extensiontest.AssertEqual(t, CompareVersions("10.1.1", "1.1.1"), 1)
	// Non integers components are treated as 0
	extensiontest.AssertEqual(t, CompareVersions("zugzug.1.1", "1.1.daboo"), -1)
}

func TestFilterForUpdates(t *testing.T) {
	lightThemeExtension, err := Contains(OfferedExtensions, "ldimlcelhnjgpjjemdjokpgeeikdinbm")
	extensiontest.AssertEqual(t, err, nil)
	darkThemeExtension, err := Contains(OfferedExtensions, "bfdgpgibhagkpdlnjonhkabjoijopoge")
	extensiontest.AssertEqual(t, err, nil)

	allExtensions := []Extension{lightThemeExtension, OfferedExtensions[1]}

	// No updates when nothing to check
	check := FilterForUpdates(allExtensions, Extensions{})
	extensiontest.AssertEqual(t, len(check), 0)

	olderExtensionCheck1 := lightThemeExtension
	olderExtensionCheck1.Version = "0.1.0"
	outdatedExtensionCheck := []Extension{olderExtensionCheck1}

	check = FilterForUpdates(allExtensions, outdatedExtensionCheck)
	extensiontest.AssertEqual(t, len(check), 1)
	extensiontest.AssertEqual(t, check[0].ID, lightThemeExtension.ID)
	// Check that the newer version,SHA, title are returned
	extensiontest.AssertEqual(t, check[0].Version, lightThemeExtension.Version)
	extensiontest.AssertEqual(t, check[0].SHA256, lightThemeExtension.SHA256)
	extensiontest.AssertEqual(t, check[0].Title, lightThemeExtension.Title)
	// Check that even if a URL is provided, we use the server's URL
	extensiontest.AssertEqual(t, check[0].URL, lightThemeExtension.URL)

	// Newer extensions have no items returned
	newerExtensionCheck := lightThemeExtension
	newerExtensionCheck.Version = "2.1.0"
	check = FilterForUpdates(allExtensions, Extensions{newerExtensionCheck})
	extensiontest.AssertEqual(t, len(check), 0)

	// 2 outdated extensions both get returned from 1 check
	olderExtensionCheck2 := darkThemeExtension
	olderExtensionCheck2.Version = "0.1.0"
	outdatedExtensionsCheck := []Extension{olderExtensionCheck1, olderExtensionCheck2}
	check = FilterForUpdates(allExtensions, outdatedExtensionsCheck)
	extensiontest.AssertEqual(t, len(check), 2)
	extensiontest.AssertEqual(t, check[0].ID, olderExtensionCheck1.ID)
	extensiontest.AssertEqual(t, check[1].ID, olderExtensionCheck2.ID)

	// Outdated extension that's blacklisted doesn't get updates
	allExtensionsBlacklisted := allExtensions
	for i := range allExtensionsBlacklisted {
		allExtensionsBlacklisted[i].Blacklisted = true
	}
	check = FilterForUpdates(allExtensionsBlacklisted, outdatedExtensionCheck)
	extensiontest.AssertEqual(t, len(check), 0)
}
