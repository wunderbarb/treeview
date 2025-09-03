//go:build windows
// +build windows

package utils

import (
	"fmt"
	"os"
)

// inodeKey returns a unique key for the given FileInfo to detect symlink loops.
// On Windows, it uses a composite key of name, size, and modification time.
func inodeKey(info os.FileInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("nil FileInfo")
	}

	// Windows: Use name + size + modTime as a weak unique key
	return fmt.Sprintf("%s:%d:%d", info.Name(), info.Size(), info.ModTime().UnixNano()), nil
}
