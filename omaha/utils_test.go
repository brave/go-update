package omaha

import (
	"testing"
)

func TestIsJSONRequest(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "JSON content type",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "XML content type",
			contentType: "application/xml",
			want:        false,
		},
		{
			name:        "Empty content type",
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSONRequest(tt.contentType); got != tt.want {
				t.Errorf("IsJSONRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
