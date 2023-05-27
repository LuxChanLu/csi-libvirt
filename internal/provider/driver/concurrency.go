package driver

import (
	"sync"
)

var globalDiskLock = &sync.Mutex{}

func (d *Driver) DiskLock(poolName, name string) func() {
	// key := strings.Join([]string{poolName, name}, ":")
	// lock, _ := d.diskLocks.LoadOrStore(key, &sync.Mutex{})
	// (lock.(*sync.Mutex)).Lock()
	// return func() {
	// 	(lock.(*sync.Mutex)).Unlock()
	// 	d.diskLocks.Delete(key)
	// }
	globalDiskLock.Lock()
	return func() {
		globalDiskLock.Unlock()
	}
}
