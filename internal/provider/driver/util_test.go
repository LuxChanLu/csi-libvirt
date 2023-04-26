//go:build !integration

package driver_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/stretchr/testify/assert"
)

func TestEncodeNumberToAlphabet(t *testing.T) {
	tests := map[int]string{
		1:    "a",
		26:   "z",
		27:   "aa",
		52:   "az",
		53:   "ba",
		8000: "kur",
	}
	driver := &driver.Driver{}
	for number, letter := range tests {
		assert.Equal(t, letter, driver.EncodeNumberToAlphabet(number))
	}
}
