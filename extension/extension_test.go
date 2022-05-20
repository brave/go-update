package extension

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	allExtensionsMap := NewExtensionMap()
	allExtensionsMap.StoreExtensions(&OfferedExtensions)
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	darkThemeExtension, ok := allExtensionsMap.Load("bfdgpgibhagkpdlnjonhkabjoijopoge")
	assert.True(t, ok)

	testExtensions := Extensions{lightThemeExtension, OfferedExtensions[1]}
	testExtensionsMap := NewExtensionMap()
	testExtensionsMap.StoreExtensions(&testExtensions)

	// No updates when nothing to check
	updateRequest := UpdateRequest{}
	check := updateRequest.FilterForUpdates(testExtensionsMap)
	assert.Equal(t, 0, len(check))

	olderExtensionCheck1 := lightThemeExtension
	olderExtensionCheck1.Version = "0.1.0"
	outdatedExtensionCheck := UpdateRequest{olderExtensionCheck1}

	check = outdatedExtensionCheck.FilterForUpdates(testExtensionsMap)
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
	updateRequest = UpdateRequest{newerExtensionCheck}
	check = updateRequest.FilterForUpdates(testExtensionsMap)
	assert.Equal(t, 0, len(check))

	// 2 outdated extensions both get returned from 1 check
	olderExtensionCheck2 := darkThemeExtension
	olderExtensionCheck2.Version = "0.1.0"
	updateRequest = UpdateRequest{olderExtensionCheck1, olderExtensionCheck2}
	check = updateRequest.FilterForUpdates(testExtensionsMap)
	assert.Equal(t, 2, len(check))
	assert.Equal(t, olderExtensionCheck1.ID, check[0].ID)
	assert.Equal(t, olderExtensionCheck2.ID, check[1].ID)

	// Outdated extension that's blacklisted doesn't get updates
	allExtensionsBlacklistedMap := allExtensionsMap
	for k := range allExtensionsBlacklistedMap.data {
		elem := allExtensionsBlacklistedMap.data[k]
		elem.Blacklisted = true
		allExtensionsBlacklistedMap.data[k] = elem
	}
	check = outdatedExtensionCheck.FilterForUpdates(allExtensionsBlacklistedMap)
	assert.Equal(t, 0, len(check))
}

func TestS3BucketForExtension(t *testing.T) {
	allExtensionsMap := NewExtensionMap()
	allExtensionsMap.StoreExtensions(&OfferedExtensions)
	torExtensionMac, ok := allExtensionsMap.Load("cldoidikboihgcjfkhdeidbpclkineef")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(torExtensionMac.ID), "tor.bravesoftware.com")
	torExtensionWin, ok := allExtensionsMap.Load("cpoalefficncklhjfpglfiplenlpccdb")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(torExtensionWin.ID), "tor.bravesoftware.com")
	torExtensionLinux, ok := allExtensionsMap.Load("biahpgbdmdkfgndcmfiipgcebobojjkp")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(torExtensionLinux.ID), "tor.bravesoftware.com")
	ipfsExtensionMac, ok := allExtensionsMap.Load("nljcddpbnaianmglkpkneakjaapinabi")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(ipfsExtensionMac.ID), "ipfs.bravesoftware.com")
	ipfsExtensionWin, ok := allExtensionsMap.Load("lnbclahgobmjphilkalbhebakmblnbij")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(ipfsExtensionWin.ID), "ipfs.bravesoftware.com")
	ipfsExtensionLinux, ok := allExtensionsMap.Load("oecghfpdmkjlhnfpmmjegjacfimiafjp")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(ipfsExtensionLinux.ID), "ipfs.bravesoftware.com")
	lightThemeExtension, ok := allExtensionsMap.Load("ldimlcelhnjgpjjemdjokpgeeikdinbm")
	assert.True(t, ok)
	assert.Equal(t, GetS3ExtensionBucketHost(lightThemeExtension.ID), "brave-core-ext.s3.brave.com")
}
