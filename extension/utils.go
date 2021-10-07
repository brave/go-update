package extension

import (
	"os"
)

var torClientMacExtensionID = "cldoidikboihgcjfkhdeidbpclkineef"
var torClientWindowsExtensionID = "cpoalefficncklhjfpglfiplenlpccdb"
var torClientLinuxExtensionID = "biahpgbdmdkfgndcmfiipgcebobojjkp"

// TorClientExtensionIDs is used to add an exception to return the dedicated
// proxy url for downloading the tor client crx
var TorClientExtensionIDs = []string{torClientMacExtensionID, torClientWindowsExtensionID, torClientLinuxExtensionID}

var ipfsClientMacExtensionID = "nljcddpbnaianmglkpkneakjaapinabi"
var ipfsClientWindowsExtensionID = "lnbclahgobmjphilkalbhebakmblnbij"
var ipfsClientLinuxExtensionID = "oecghfpdmkjlhnfpmmjegjacfimiafjp"
var ipfsClientMacArm64ExtensionID = "lejaflgbgglfaomemffoaappaihfligf"

// ipfsClientExtensionIDs is used to add an exception to return the dedicated
// proxy url for downloading the ipfs crx
var ipfsClientExtensionIDs = []string{ipfsClientMacExtensionID, ipfsClientWindowsExtensionID, ipfsClientLinuxExtensionID, ipfsClientMacArm64ExtensionID}

func isTorExtension(id string) bool {
	for _, torID := range TorClientExtensionIDs {
		if torID == id {
			return true
		}
	}
	return false
}

func isIPFSExtension(id string) bool {
	for _, ipfsID := range ipfsClientExtensionIDs {
		if ipfsID == id {
			return true
		}
	}
	return false
}

// GetS3ExtensionBucketHost returns the url to use for accessing crx files
func GetS3ExtensionBucketHost(id string) string {
	if isTorExtension(id) {
		return GetS3TorExtensionBucketHost()
	}

	if isIPFSExtension(id) {
		return GetS3IPFSExtensionBucketHost()
	}

	s3BucketHost, ok := os.LookupEnv("S3_EXTENSIONS_BUCKET_HOST")
	if !ok {
		s3BucketHost = "brave-core-ext.s3.brave.com"
	}
	return s3BucketHost
}

// GetS3TorExtensionBucketHost returns the url to use for accessing tor client crx
func GetS3TorExtensionBucketHost() string {
	s3BucketHost, ok := os.LookupEnv("S3_EXTENSIONS_BUCKET_HOST_TOR")
	if !ok {
		s3BucketHost = "tor.bravesoftware.com"
	}
	return s3BucketHost
}

// GetS3IPFSExtensionBucketHost returns the url to use for accessing go-ipfs client crx
func GetS3IPFSExtensionBucketHost() string {
	s3BucketHost, ok := os.LookupEnv("S3_EXTENSIONS_BUCKET_HOST_IPFS")
	if !ok {
		s3BucketHost = "ipfs.bravesoftware.com"
	}
	return s3BucketHost
}

// GetUpdateStatus returns the status of an update response for an extension
func GetUpdateStatus(extension Extension) string {
	if extension.Status == "" {
		return "ok"
	}
	return extension.Status
}
