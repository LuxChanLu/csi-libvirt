# libvirt-csi
K8S CSI Using libvirt VMs

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
reclaimPolicy: Retain
allowVolumeExpansion: true
```