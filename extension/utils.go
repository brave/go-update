package extension

import (
	"os"
)

// GetS3ExtensionBucketHost returns the bucket to use for accessing the crx files
func GetS3ExtensionBucketHost() string {
	s3BucketHost, ok := os.LookupEnv("S3_EXTENSIONS_BUCKET_HOST")
	if !ok {
		s3BucketHost = "brave-core-ext.s3.brave.com"
	}
	return s3BucketHost
}

// UpdateStatus returns the status of an update response for an extension
func UpdateStatus(extension Extension) string {
	if extension.Status == "" {
		return "ok"
	}
	return extension.Status
}
