package omaha

import (
	"testing"
)

func TestDetectProtocolVersion(t *testing.T) {
	// Test with empty data
	t.Run("Empty data", func(t *testing.T) {
		version, err := DetectProtocolVersion(nil, "")
		if err != nil {
			t.Errorf("DetectProtocolVersion() error = %v", err)
			return
		}
		if version != "3.1" {
			t.Errorf("DetectProtocolVersion() = %v, want %v", version, "3.1")
		}
	})

	// Test with valid JSON data
	t.Run("Valid JSON", func(t *testing.T) {
		jsonData := []byte(`{"request":{"protocol":"3.0","app":[{"appid":"test-app-id"}]}}`)
		version, err := DetectProtocolVersion(jsonData, "application/json")
		if err != nil {
			t.Errorf("DetectProtocolVersion() error = %v", err)
			return
		}
		if version != "3.0" {
			t.Errorf("DetectProtocolVersion() = %v, want %v", version, "3.0")
		}
	})

	// Test with valid XML data
	t.Run("Valid XML", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.1">
			<app appid="test-app-id"></app>
		</request>`)
		version, err := DetectProtocolVersion(xmlData, "application/xml")
		if err != nil {
			t.Errorf("DetectProtocolVersion() error = %v", err)
			return
		}
		if version != "3.1" {
			t.Errorf("DetectProtocolVersion() = %v, want %v", version, "3.1")
		}
	})
}

func TestIsProtocolVersionSupported(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "Version 3.0",
			version: "3.0",
			want:    true,
		},
		{
			name:    "Version 3.1",
			version: "3.1",
			want:    true,
		},
		{
			name:    "Unsupported version",
			version: "3.99",
			want:    false,
		},
		{
			name:    "Empty version",
			version: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProtocolVersionSupported(tt.version); got != tt.want {
				t.Errorf("IsProtocolVersionSupported(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
