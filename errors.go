package treeview

import (
	"errors"
	"fmt"
)

var (
	// ErrEmptyID is returned when a node is created with an empty ID string,
	// which would prevent proper tree construction and lookups.
	ErrEmptyID = errors.New("empty node ID")

	// ErrTraversalLimit is raised when TraversalCap
	// has been exceeded during a build or file-system scan.
	ErrTraversalLimit = errors.New("traversal limit exceeded")

	// ErrNodeNotFound is returned by lookup helpers when the requested node
	// does not exist in the tree.
	ErrNodeNotFound = errors.New("node not found in tree")

	// ErrCyclicReference is returned when building a tree encounters a cycle
	// in parent-child relationships.
	ErrCyclicReference = errors.New("cyclic reference detected in tree")

	// ErrTreeConstruction is returned when tree building fails at a high level.
	ErrTreeConstruction = errors.New("tree construction failed")

	// ErrFileSystem is returned when file system operations fail.
	ErrFileSystem = errors.New("file system operation failed")

	// ErrPathResolution is returned when a path cannot be resolved.
	ErrPathResolution = errors.New("path resolution failed")

	// ErrDirectoryScan is returned when directory scanning fails.
	ErrDirectoryScan = errors.New("directory scan failed")
)

// pathError creates an error that includes path context.
// It's used internally for file system operations where the path is important.
func pathError(sentinel error, path string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", sentinel, path)
	}
	return fmt.Errorf("%w: %s: %w", sentinel, path, cause)
}

// cyclicReferenceError creates an error for cyclic references with path details.
func cyclicReferenceError(nodeID, parentID string) error {
	return fmt.Errorf("%w: node %q -> parent %q", ErrCyclicReference, nodeID, parentID)
}
