package system

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/mem"
)

var (
	// Variables for mocking in tests
	goarch    = runtime.GOARCH
	// hasAVX2 is set in arch-specific files
	getMemory = func() (uint64, error) {
		v, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		return v.Total, nil
	}
)

// IsParakeetSupported checks if the system supports the Parakeet ONNX model.
// It requires an amd64 or arm64 architecture, AVX2 support for amd64, and at least 8GB of RAM.
func IsParakeetSupported() bool {
	// 1. Check Architecture
	if goarch != "amd64" && goarch != "arm64" {
		return false
	}

	// 2. Check AVX2 for amd64
	if goarch == "amd64" && !hasAVX2 {
		return false
	}

	// 3. Check RAM >= 8GB
	totalMem, err := getMemory()
	if err != nil {
		// If we can't check memory, we default to unsupported
		return false
	}

	const minMem uint64 = 8 * 1024 * 1024 * 1024 // 8GB
	if totalMem < minMem {
		return false
	}

	return true
}
