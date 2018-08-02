package extension

import (
	"github.com/brave/go-update/extension/extensiontest"
	"testing"
)

func TestContains(t *testing.T) {
	extension1 := OfferedExtensions[0]
	extension2 := OfferedExtensions[1]
	unknownExtension := extension1;
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
