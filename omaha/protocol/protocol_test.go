package protocol

import (
	"testing"
)

func TestIsJSONContentType(t *testing.T) {
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
			name:        "JSON with charset",
			contentType: "application/json; charset=utf-8",
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
		{
			name:        "Text plain",
			contentType: "text/plain",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSONContentType(tt.contentType); got != tt.want {
				t.Errorf("IsJSONContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAcceptsJSON(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{
			name:   "JSON only",
			accept: "application/json",
			want:   true,
		},
		{
			name:   "JSON with quality value",
			accept: "application/json;q=0.9",
			want:   true,
		},
		{
			name:   "Accept header with JSON first",
			accept: "application/json, application/xml",
			want:   true,
		},
		{
			name:   "Accept header with JSON second",
			accept: "application/xml, application/json",
			want:   true,
		},
		{
			name:   "Accept header with quality values",
			accept: "text/html, application/json;q=0.9, */*;q=0.8",
			want:   true,
		},
		{
			name:   "Accept header without JSON",
			accept: "text/html, application/xml;q=0.9, */*;q=0.8",
			want:   false,
		},
		{
			name:   "Wildcard only",
			accept: "*/*",
			want:   false,
		},
		{
			name:   "Empty accept",
			accept: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AcceptsJSON(tt.accept); got != tt.want {
				t.Errorf("AcceptsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

	// Test with invalid JSON data
	t.Run("Invalid JSON", func(t *testing.T) {
		jsonData := []byte(`{"request":{"protocol":"3.0",}`)
		_, err := DetectProtocolVersion(jsonData, "application/json")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for invalid JSON")
		}
	})

	// Test with JSON missing request object
	t.Run("JSON missing request object", func(t *testing.T) {
		jsonData := []byte(`{"protocol":"3.0"}`)
		_, err := DetectProtocolVersion(jsonData, "application/json")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for missing request object")
		}
	})

	// Test with JSON missing protocol field
	t.Run("JSON missing protocol field", func(t *testing.T) {
		jsonData := []byte(`{"request":{"app":[{"appid":"test-app-id"}]}}`)
		_, err := DetectProtocolVersion(jsonData, "application/json")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for missing protocol field")
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

	// Test with invalid XML data
	t.Run("Invalid XML", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request protocol="3.1"
			<app appid="test-app-id"></app>
		</request>`)
		_, err := DetectProtocolVersion(xmlData, "application/xml")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for invalid XML")
		}
	})

	// Test with XML missing request element
	t.Run("XML missing request element", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<dummy protocol="3.1">
			<app appid="test-app-id"></app>
		</dummy>`)
		_, err := DetectProtocolVersion(xmlData, "application/xml")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for missing request element")
		}
	})

	// Test with XML missing protocol attribute
	t.Run("XML missing protocol attribute", func(t *testing.T) {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<request>
			<app appid="test-app-id"></app>
		</request>`)
		_, err := DetectProtocolVersion(xmlData, "application/xml")
		if err == nil {
			t.Errorf("DetectProtocolVersion() error = nil, expected an error for missing protocol attribute")
		}
	})
}
