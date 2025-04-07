package common

// IsJSONRequest is used to check if JSON parser should be used
func IsJSONRequest(contentType string) bool {
	return contentType == "application/json"
}
