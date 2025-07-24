package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestResolvePath(t *testing.T) {
	// Get actual home directory for testing
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	// Get current working directory for absolute path tests
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple_relative",
			path: ".",
			want: cwd,
		},
		{
			name: "parent_directory",
			path: "..",
			want: filepath.Dir(cwd),
		},
		{
			name: "home_tilde",
			path: "~",
			want: home,
		},
		{
			name: "home_with_subdir",
			path: "~/Documents",
			want: filepath.Join(home, "Documents"),
		},
		{
			name: "absolute_path",
			path: "/tmp",
			want: "/tmp",
		},
		{
			name: "relative_with_dots",
			path: "./test/../test",
			want: filepath.Join(cwd, "test"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolvePath(test.path)
			if err != nil {
				t.Errorf("ResolvePath(%q) returned error: %v", test.path, err)
				return
			}
			if got != test.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestResolvePathErrors(t *testing.T) {
	t.Run("invalid_home_path", func(t *testing.T) {
		// Test with invalid tilde expansion that could fail
		// We'll simulate a scenario where filepath.Abs could fail
		// by using a very long path that exceeds system limits
		longPath := "~/" + strings.Repeat("a", 4096)
		got, err := ResolvePath(longPath)
		// This may or may not fail depending on the system, but we test the error handling
		if err != nil && !strings.Contains(err.Error(), "resolvePath:") {
			t.Errorf("ResolvePath(%q) error = %v, want error containing 'resolvePath:' if any", longPath[:50]+"...", err)
		}
		_ = got // Avoid unused variable
	})
}

func TestSafeStat(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create test directory
	testDir := filepath.Join(tmpDir, "testdir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("regular_file_no_follow", func(t *testing.T) {
		visited := make(map[string]struct{})
		info, err := SafeStat(testFile, false, visited)
		if err != nil {
			t.Errorf("SafeStat(%q, false, visited) returned error: %v", testFile, err)
			return
		}
		if info == nil {
			t.Errorf("SafeStat(%q, false, visited) = nil, want FileInfo", testFile)
			return
		}
		if info.IsDir() {
			t.Errorf("SafeStat(%q, false, visited).IsDir() = true, want false", testFile)
		}
	})

	t.Run("directory", func(t *testing.T) {
		visited := make(map[string]struct{})
		info, err := SafeStat(testDir, false, visited)
		if err != nil {
			t.Errorf("SafeStat(%q, false, visited) returned error: %v", testDir, err)
			return
		}
		if info == nil {
			t.Errorf("SafeStat(%q, false, visited) = nil, want FileInfo", testDir)
			return
		}
		if !info.IsDir() {
			t.Errorf("SafeStat(%q, false, visited).IsDir() = false, want true", testDir)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		visited := make(map[string]struct{})
		nonexistent := filepath.Join(tmpDir, "nonexistent")
		info, err := SafeStat(nonexistent, false, visited)
		if err == nil {
			t.Errorf("SafeStat(%q, false, visited) = %v, want error", nonexistent, info)
		}
		if !strings.Contains(err.Error(), "lstat") {
			t.Errorf("SafeStat(%q, false, visited) error = %v, want error containing 'lstat'", nonexistent, err)
		}
	})

	// Test symlink detection if supported on this platform
	if runtime.GOOS != "windows" {
		t.Run("symlink_no_follow", func(t *testing.T) {
			linkPath := filepath.Join(tmpDir, "link")
			err := os.Symlink(testFile, linkPath)
			if err != nil {
				t.Skip("Cannot create symlink:", err)
			}

			visited := make(map[string]struct{})
			info, err := SafeStat(linkPath, false, visited)
			if err != nil {
				t.Errorf("SafeStat(%q, false, visited) returned error: %v", linkPath, err)
				return
			}
			if (info.Mode() & os.ModeSymlink) == 0 {
				t.Errorf("SafeStat(%q, false, visited).Mode() = %v, want symlink", linkPath, info.Mode())
			}
		})

		t.Run("symlink_follow", func(t *testing.T) {
			linkPath := filepath.Join(tmpDir, "link2")
			err := os.Symlink(testFile, linkPath)
			if err != nil {
				t.Skip("Cannot create symlink:", err)
			}

			visited := make(map[string]struct{})
			info, err := SafeStat(linkPath, true, visited)
			if err != nil {
				t.Errorf("SafeStat(%q, true, visited) returned error: %v", linkPath, err)
				return
			}
			if (info.Mode() & os.ModeSymlink) != 0 {
				t.Errorf("SafeStat(%q, true, visited).Mode() = %v, want regular file", linkPath, info.Mode())
			}
		})

		t.Run("symlink_loop_detection", func(t *testing.T) {
			link1 := filepath.Join(tmpDir, "loop1")
			link2 := filepath.Join(tmpDir, "loop2")

			err := os.Symlink(link2, link1)
			if err != nil {
				t.Skip("Cannot create symlink:", err)
			}
			err = os.Symlink(link1, link2)
			if err != nil {
				t.Skip("Cannot create symlink:", err)
			}

			visited := make(map[string]struct{})
			info, err := SafeStat(link1, true, visited)
			if err == nil {
				t.Errorf("SafeStat(%q, true, visited) = %v, want symlink loop error", link1, info)
			}
			// EvalSymlinks detects loops and returns "too many links" error
			if !strings.Contains(err.Error(), "too many links") && !strings.Contains(err.Error(), "symlink loop detected") {
				t.Errorf("SafeStat(%q, true, visited) error = %v, want error containing 'too many links' or 'symlink loop detected'", link1, err)
			}
		})

		t.Run("broken_symlink_follow", func(t *testing.T) {
			brokenLink := filepath.Join(tmpDir, "broken")
			err := os.Symlink("nonexistent", brokenLink)
			if err != nil {
				t.Skip("Cannot create symlink:", err)
			}

			visited := make(map[string]struct{})
			info, err := SafeStat(brokenLink, true, visited)
			if err == nil {
				t.Errorf("SafeStat(%q, true, visited) = %v, want error", brokenLink, info)
			}
			// The error could be either "eval symlink" or "stat resolved" depending on where it fails
			if !strings.Contains(err.Error(), "eval symlink") && !strings.Contains(err.Error(), "stat resolved") {
				t.Errorf("SafeStat(%q, true, visited) error = %v, want error containing 'eval symlink' or 'stat resolved'", brokenLink, err)
			}
		})
	}
}

func TestInodeKey(t *testing.T) {
	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "inode_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	info, err := tmpFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("valid_fileinfo", func(t *testing.T) {
		key, err := inodeKey(info)
		if err != nil {
			t.Errorf("inodeKey(%v) returned error: %v", info, err)
			return
		}
		if key == "" {
			t.Errorf("inodeKey(%v) = %q, want non-empty string", info, key)
		}

		// Key should be consistent
		key2, err := inodeKey(info)
		if err != nil {
			t.Errorf("inodeKey(%v) returned error on second call: %v", info, err)
			return
		}
		if key != key2 {
			t.Errorf("inodeKey(%v) inconsistent: first=%q, second=%q", info, key, key2)
		}
	})

	t.Run("nil_fileinfo", func(t *testing.T) {
		key, err := inodeKey(nil)
		if err == nil {
			t.Errorf("inodeKey(nil) = %q, want error", key)
		}
		if !strings.Contains(err.Error(), "nil FileInfo") {
			t.Errorf("inodeKey(nil) error = %v, want error containing 'nil FileInfo'", err)
		}
	})

	t.Run("platform_specific_format", func(t *testing.T) {
		key, err := inodeKey(info)
		if err != nil {
			t.Errorf("inodeKey(%v) returned error: %v", info, err)
			return
		}

		if strings.Contains(runtime.GOOS, "windows") {
			// Windows format: name:size:modTime
			if !strings.Contains(key, ":") {
				t.Errorf("inodeKey(%v) on Windows = %q, want format with colons", info, key)
			}
		} else {
			// Unix format should be dev:X_ino:Y or fallback
			if !strings.Contains(key, "dev:") && !strings.Contains(key, "fallback:") {
				t.Errorf("inodeKey(%v) on Unix = %q, want dev: or fallback: format", info, key)
			}
		}
	})
}

func TestInodeKeyUnixFallback(t *testing.T) {
	if strings.Contains(runtime.GOOS, "windows") {
		t.Skip("Unix-specific test")
	}

	// Create a mock FileInfo that will cause type assertion to fail
	mockInfo := &mockFileInfo{
		name: "test",
		size: 123,
	}

	key, err := inodeKey(mockInfo)
	if err != nil {
		t.Errorf("inodeKey(mockInfo) returned error: %v", err)
		return
	}
	if !strings.Contains(key, "fallback:") {
		t.Errorf("inodeKey(mockInfo) = %q, want fallback format", key)
	}
}

// mockFileInfo implements os.FileInfo for testing fallback behavior
type mockFileInfo struct {
	name string
	size int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil } // This causes type assertion to fail
