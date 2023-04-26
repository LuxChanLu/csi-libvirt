//go:build !integration

package options_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/options"
	"github.com/stretchr/testify/assert"
)

func TestAddOptions(t *testing.T) {
	assert.NotNil(t, options.AppOptions())
}
