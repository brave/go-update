package extension

import (
	"os"
)

var (
	torClientMacExtensionID        = "cldoidikboihgcjfkhdeidbpclkineef"
	torClientWindowsExtensionID    = "cpoalefficncklhjfpglfiplenlpccdb"
	torClientLinuxExtensionID      = "biahpgbdmdkfgndcmfiipgcebobojjkp"
	torClientLinuxArm64ExtensionID = "monolafkoghdlanndjfeebmdfkbklejg"
)

var (
	torPluggableTransportsMacExtensionID     = "einfndjnccmoohcngmlldpmellegjjnk"
	torPluggableTransportsWindowsExtensionID = "dnkcahhmfcanmkjhnjejoomdihffoefm"
	torPluggableTransportsLinuxExtensionID   = "apfggiafobakjahnkchiecbomjgigkkn"
)

// TorClientExtensionIDs is used to add an exception to return the dedicated
// proxy url for downloading the tor client crx
var (
	TorClientExtensionIDs              = []string{torClientMacExtensionID, torClientWindowsExtensionID, torClientLinuxExtensionID, torClientLinuxArm64ExtensionID}
	TorPluggableTransportsExtensionIDs = []string{torPluggableTransportsMacExtensionID, torPluggableTransportsWindowsExtensionID, torPluggableTransportsLinuxExtensionID}
)

func isTorExtension(id string) bool {
	for _, torID := range TorClientExtensionIDs {
		if torID == id {
			return true
		}
	}
	for _, torPtID := range TorPluggableTransportsExtensionIDs {
		if torPtID == id {
			return true
		}
	}
	return false
}

func lookupEnvFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetS3ExtensionBucketHost returns the url to use for accessing crx files
func GetS3ExtensionBucketHost(id string) string {
	if isTorExtension(id) {
		return GetS3TorExtensionBucketHost()
	}

	return lookupEnvFallback("S3_EXTENSIONS_BUCKET_HOST", "brave-core-ext.s3.brave.com")
}

// GetS3TorExtensionBucketHost returns the url to use for accessing tor client crx
func GetS3TorExtensionBucketHost() string {
	return lookupEnvFallback("S3_EXTENSIONS_BUCKET_HOST_TOR", "tor.bravesoftware.com")
}

// GetUpdateStatus returns the status of an update response for an extension
func GetUpdateStatus(extension Extension) string {
	if extension.Status == "" {
		return "ok"
	}
	return extension.Status
}

// GetComponentUpdaterHost returns the url to use for component updates (@updater=BraveComponentUpdater)
func GetComponentUpdaterHost() string {
	return lookupEnvFallback("COMPONENT_UPDATER_HOST", "componentupdater.brave.com")
}

// GetExtensionUpdaterHost returns the url to use for extension updates (@updater=chromiumcrx)
func GetExtensionUpdaterHost() string {
	return lookupEnvFallback("EXTENSION_UPDATER_HOST", "extensionupdater.brave.com")
}

// GetUpdaterHostByType returns the appropriate updater host based on the updater type
func GetUpdaterHostByType(updaterType string) string {
	switch updaterType {
	case "chromiumcrx":
		return GetExtensionUpdaterHost()
	case "BraveComponentUpdater":
		return GetComponentUpdaterHost()
	default:
		// Backward compatibility
		return GetComponentUpdaterHost()
	}
}
