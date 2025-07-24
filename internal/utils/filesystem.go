package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// ResolvePath resolves the path to an absolute path, handling `~`, `..`, `.` expansion
func ResolvePath(path string) (string, error) {
	// Handle ~ by converting ~/folder to /home/user/folder
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolvePath: %w", err)
		}
		// Replace ~ with actual home directory
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path, resolving any relative components
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolvePath: %w", err)
	}

	// Clean removes any redundant path elements
	return filepath.Clean(absPath), nil
}

// SafeStat is a helper function that gets file info for a given path, handling symlinks and loops.
func SafeStat(path string, follow bool, visited map[string]struct{}) (os.FileInfo, error) {
	// Always start with lstat to check if it's a symlink
	// lstat doesn't follow symlinks, giving us the link's own info
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("lstat %s: %w", path, err)
	}

	// Handle symlinks if we're configured to follow them
	if follow && (info.Mode()&os.ModeSymlink) != 0 {
		// Resolve the symlink to its target
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("eval symlink %s: %w", path, err)
		}

		// Get info about the symlink target
		info, err = os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("stat resolved %s: %w", resolved, err)
		}
	}

	// Check for symlink loops using inode tracking
	// Get a unique key for this file (device:inode on Unix)
	key, err := inodeKey(info)
	if err != nil {
		return nil, err
	}

	// If we've seen this inode before, we have a loop
	if _, ok := visited[key]; ok {
		return nil, errors.New("symlink loop detected")
	}

	// Mark this inode as visited
	visited[key] = struct{}{}
	return info, nil
}

// inodeKey returns a unique key for the given FileInfo to detect symlink loops.
//
// On Unix-like systems (Linux, macOS, etc.), it uses the device and inode
// numbers which provide a stable unique identifier for each file.
//
// On Windows, it falls back to a composite key using file attributes since
// Windows does not expose inode numbers directly.
func inodeKey(info os.FileInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("nil FileInfo")
	}

	if strings.Contains(runtime.GOOS, "windows") {
		// Windows: Use name + size + modTime as a weak unique key
		return fmt.Sprintf("%s:%d:%d", info.Name(), info.Size(), info.ModTime().UnixNano()), nil
	}

	// Unix-like systems: Use device and inode numbers as unique identifier
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Fallback if type assertion fails
		return fmt.Sprintf("fallback:%s_%d", info.Name(), info.ModTime().UnixNano()), nil
	}

	return fmt.Sprintf("dev:%d_ino:%d", stat.Dev, stat.Ino), nil
}
