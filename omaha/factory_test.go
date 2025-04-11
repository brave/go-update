package omaha

import (
	"testing"

	v3 "github.com/brave/go-update/omaha/v3"
	v4 "github.com/brave/go-update/omaha/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFactoryCreateProtocol(t *testing.T) {
	factory := &DefaultFactory{}

	// Test v3 protocol versions
	protocol, err := factory.CreateProtocol("3.0")
	assert.NoError(t, err)
	assert.NotNil(t, protocol)
	assert.IsType(t, &v3.VersionedHandler{}, protocol)
	assert.Equal(t, "3.0", protocol.GetVersion())

	protocol, err = factory.CreateProtocol("3.1")
	assert.NoError(t, err)
	assert.NotNil(t, protocol)
	assert.IsType(t, &v3.VersionedHandler{}, protocol)
	assert.Equal(t, "3.1", protocol.GetVersion())

	// Test v4 protocol versions
	protocol, err = factory.CreateProtocol("4.0")
	assert.NoError(t, err)
	assert.NotNil(t, protocol)
	assert.IsType(t, &v4.VersionedHandler{}, protocol)
	assert.Equal(t, "4.0", protocol.GetVersion())

	// Test unsupported version
	protocol, err = factory.CreateProtocol("3.11")
	assert.Error(t, err)
	assert.Nil(t, protocol)
	assert.Contains(t, err.Error(), "unsupported protocol version")

	protocol, err = factory.CreateProtocol("4.99")
	assert.Error(t, err)
	assert.Nil(t, protocol)
	assert.Contains(t, err.Error(), "unsupported protocol version")

	// Test unsupported version that should default to v3
	protocol, err = factory.CreateProtocol("2.0")
	assert.Error(t, err)
	assert.Nil(t, protocol)
	assert.Contains(t, err.Error(), "unsupported protocol version")
}
