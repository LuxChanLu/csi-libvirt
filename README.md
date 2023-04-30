# csi-libvirt
K8S CSI Using libvirt VMs

[![Test](https://github.com/LuxChanLu/csi-libvirt/actions/workflows/test.yaml/badge.svg)](https://github.com/LuxChanLu/csi-libvirt/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/LuxChanLu/csi-libvirt/badge.svg?branch=main)](https://coveralls.io/github/LuxChanLu/csi-libvirt?branch=main)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/1e543884918542358390722aa106e419)](https://app.codacy.com/gh/LuxChanLu/csi-libvirt/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Maintainability](https://api.codeclimate.com/v1/badges/d281fd717dcfc06b3e8f/maintainability)](https://codeclimate.com/github/LuxChanLu/csi-libvirt/maintainability)
[![Go Report Card](https://goreportcard.com/badge/github.com/LuxChanLu/libvirt-csi)](https://goreportcard.com/report/github.com/LuxChanLu/libvirt-csi)

# StroageClass example
```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: storage-libvirt-xfs-retain
provisioner: lu.lxc.csi.libvirt
parameters:
  fstype: xfs
  pool: pool
  bus: virtio # virtio/usb/sata/ide
reclaimPolicy: Retain
allowVolumeExpansion: true
```

# TODO
- SCSI or other Rescan on node