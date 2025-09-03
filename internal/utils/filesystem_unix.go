//go:build !windows
// +build !windows

package utils

import (
	"fmt"
	"os"
	"syscall"
)

// inodeKey returns a unique key for the given FileInfo to detect symlink loops.
// On Unix-like systems, it uses the device and inode numbers.
func inodeKey(info os.FileInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("nil FileInfo")
	}

	// Unix-like systems: Use device and inode numbers as unique identifier
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Fallback if type assertion fails
		return fmt.Sprintf("fallback:%s_%d", info.Name(), info.ModTime().UnixNano()), nil
	}

	return fmt.Sprintf("dev:%d_ino:%d", stat.Dev, stat.Ino), nil
}
