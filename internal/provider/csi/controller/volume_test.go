//go:build integration

package controller_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

func TestCreateVolume(t *testing.T) {
	controller := testContoller(t)
	name := "pvc-9a4901d9-b802-49a3-a1ac-4623a1f50513"
	capacity, response := controller.createTestVolume(name)

	// Assert actually created in libvirt
	pool, err := controller.libvirt.StoragePoolLookupByName("default")
	assert.NoError(t, err)
	vol, err := controller.libvirt.StorageVolLookupByName(pool, name)
	assert.NoError(t, err)
	volType, volCapacity, volAllocation, err := controller.libvirt.StorageVolGetInfo(vol)
	assert.NoError(t, err)
	assert.Equal(t, capacity.LimitBytes, int64(volAllocation))
	assert.Equal(t, response.Volume.CapacityBytes, int64(volCapacity))
	assert.Equal(t, volType, int8(0))
	assert.NoError(t, controller.libvirt.StorageVolDelete(vol, 0))
}

func TestDeleteVolume(t *testing.T) {
	controller := testContoller(t)
	name := "pvc-aa4901d9-b802-ffa3-a1ac-1123a1f50544"
	_, createResponse := controller.createTestVolume(name)
	response, err := controller.DeleteVolume(context.Background(), &csi.DeleteVolumeRequest{VolumeId: createResponse.Volume.VolumeId})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Assert actually deleted in libvirt
	pool, err := controller.libvirt.StoragePoolLookupByName("default")
	assert.NoError(t, err)
	_, err = controller.libvirt.StorageVolLookupByName(pool, name)
	assert.Error(t, err)
}
