package provider_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestProvideLogger(t *testing.T) {
	assert.NotNil(t, provider.ProvideLogger())
}
