package omaha

import (
	"testing"
)

func TestDefaultFactory_CreateProtocol(t *testing.T) {
	factory := &DefaultFactory{}

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "Valid version 3.0",
			version: "3.0",
			wantErr: false,
		},
		{
			name:    "Valid version 3.1",
			version: "3.1",
			wantErr: false,
		},
		{
			name:    "Invalid version",
			version: "0.0.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol, err := factory.CreateProtocol(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultFactory.CreateProtocol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && protocol.GetVersion() != tt.version {
				t.Errorf("DefaultFactory.CreateProtocol().GetVersion() = %v, want %v", protocol.GetVersion(), tt.version)
			}
		})
	}
}
