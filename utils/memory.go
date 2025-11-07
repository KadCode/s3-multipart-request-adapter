package utils

import (
	"runtime"
	"sync"
)

var (
	maxMem uint64     // stores the maximum memory allocated so far
	memMux sync.Mutex // protects access to maxMem
)

// updateMaxMemory checks current memory usage and updates maxMem if higher
func updateMaxMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m) // read current memory stats

	memMux.Lock()
	if m.Alloc > maxMem { // compare current allocation with max
		maxMem = m.Alloc
	}
	memMux.Unlock()
}

// getMaxMemory returns the maximum memory usage recorded
func getMaxMemory() uint64 {
	memMux.Lock()
	defer memMux.Unlock()
	return maxMem
}
