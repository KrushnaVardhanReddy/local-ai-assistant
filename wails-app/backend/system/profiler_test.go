package system

import (
	"errors"
	"testing"
)

// We need to store original getMemory at init to access it in TestGetMemory
var origGetMemoryVar = getMemory

func TestIsParakeetSupported(t *testing.T) {
	// Save original vars
	origGoarch := goarch
	origHasAVX2 := hasAVX2
	origGetMemory := getMemory

	// Restore after tests
	defer func() {
		goarch = origGoarch
		hasAVX2 = origHasAVX2
		getMemory = origGetMemory
	}()

	const minMem uint64 = 8 * 1024 * 1024 * 1024

	tests := []struct {
		name      string
		arch      string
		avx2      bool
		memSize   uint64
		memErr    error
		supported bool
	}{
		{
			name:      "Supported - amd64, AVX2, enough RAM",
			arch:      "amd64",
			avx2:      true,
			memSize:   minMem + 1024,
			memErr:    nil,
			supported: true,
		},
		{
			name:      "Supported - arm64, AVX2 NA, enough RAM",
			arch:      "arm64",
			avx2:      false,
			memSize:   minMem,
			memErr:    nil,
			supported: true,
		},
		{
			name:      "Unsupported - unsupported arch",
			arch:      "386",
			avx2:      true,
			memSize:   minMem,
			memErr:    nil,
			supported: false,
		},
		{
			name:      "Unsupported - amd64, no AVX2",
			arch:      "amd64",
			avx2:      false,
			memSize:   minMem,
			memErr:    nil,
			supported: false,
		},
		{
			name:      "Unsupported - not enough RAM",
			arch:      "amd64",
			avx2:      true,
			memSize:   minMem - 1024,
			memErr:    nil,
			supported: false,
		},
		{
			name:      "Unsupported - memory check error",
			arch:      "amd64",
			avx2:      true,
			memSize:   minMem,
			memErr:    errors.New("mock memory error"),
			supported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			goarch = tt.arch
			hasAVX2 = tt.avx2
			getMemory = func() (uint64, error) {
				return tt.memSize, tt.memErr
			}

			// Run
			result := IsParakeetSupported()

			// Assert
			if result != tt.supported {
				t.Errorf("IsParakeetSupported() = %v; want %v", result, tt.supported)
			}
		})
	}
}

func TestGetMemory(t *testing.T) {
	_, _ = origGetMemoryVar()
}
