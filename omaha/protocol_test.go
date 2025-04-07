package omaha

import (
	"testing"

	"github.com/brave/go-update/omaha/common"
)

func TestCommonIsJSONRequest(t *testing.T) {
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
			got := common.IsJSONRequest(tt.contentType)
			if got != tt.want {
				t.Errorf("common.IsJSONRequest() = %v, want %v", got, tt.want)
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
