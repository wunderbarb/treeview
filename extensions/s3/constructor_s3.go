package s3

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/s3"
)

// NewTreeFromS3 creates a new tree structure based on files fetched from an S3 path, using configurable options.
// Returns a pointer to a Tree structure or an error if an issue occurs during tree creation.
//
// Supported options:
// Build options:
//   - treeview.WithFilterFunc:   Filters items during tree building
//   - treeview.WithMaxDepth:     Limits tree depth during construction
//   - treeview.WithExpandFunc:   Sets initial expansion state for nodes
//   - treeview.WithTraversalCap: Limits total nodes processed (returns a partial tree + error if exceeded)
//   - treeview.WithProgressCallback: Invoked after each filesystem entry is processed (breadth-first per directory)
func NewTreeFromS3(ctx context.Context, path string, profile string,
	opts ...treeview.Option[treeview.FileInfo]) (*treeview.Tree[treeview.FileInfo], error) {
	cfg := treeview.NewMasterConfig(opts, treeview.WithProvider[treeview.FileInfo](treeview.NewDefaultNodeProvider(
		treeview.WithFileExtensionRules[treeview.FileInfo](),
	)))
	nodes, err := buildNodes(ctx, path, profile, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", treeview.ErrFileSystem, err)
	}
	tree := treeview.NewTreeFromCfg(nodes, cfg)
	return tree, nil
}

// pathError creates an error that includes path context.
// It's used internally for file system operations where the path is important.
func pathError(sentinel error, path string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", sentinel, path)
	}
	return fmt.Errorf("%w: %s: %w", sentinel, path, cause)
}

func buildNodes(ctx context.Context, path string, profile string,
	cfg *treeview.MasterConfig[treeview.FileInfo]) ([]*treeview.Node[treeview.FileInfo], error) {
	info, err := s3.Info(ctx, path, s3.WithProfile(profile))
	if err != nil {
		return nil, pathError(treeview.ErrPathResolution, path, err)
	}
	d, e, err := sortEntries(ctx, path, profile)
	if err != nil {
		return nil, pathError(treeview.ErrPathResolution, path, err)
	}
	rootNode := treeview.NewFileSystemNode(path, info)
	p := info.Path()
	m := make(map[string]*treeview.Node[treeview.FileInfo])
	if p != "" {
		m[p] = rootNode
	}
	var total int
	for _, oi := range d {
		a := strings.TrimSuffix(oi.Path(), "/")
		dir := strings.TrimPrefix(filepath.Dir(a), "/")
		if cfg.HasDepthLimitBeenReached(strings.Count(dir, "/")) {
			continue
		}
		var parentNode *treeview.Node[treeview.FileInfo]
		switch dir {
		case "", ".":
			parentNode = rootNode
		default:
			parentNode = m[dir]
		}
		info, err := oi.Info()
		if err != nil {
			return nil, pathError(treeview.ErrFileSystem, oi.Path(), err)
		}
		childNode := treeview.NewFileSystemNode(oi.Path(), info)
		cfg.HandleExpansion(childNode)
		m[strings.TrimPrefix(a, "/")] = childNode
		parentNode.AddChild(childNode)
	}
	for _, oi := range e {
		dir := strings.TrimPrefix(filepath.Dir(oi.Path()), "/")
		info, err := oi.Info()
		if err != nil {
			return nil, pathError(treeview.ErrFileSystem, oi.Path(), err)
		}
		if cfg.HasTraversalCapBeenReached(total) || cfg.HasDepthLimitBeenReached(strings.Count(dir,
			"/")) || cfg.ShouldFilter(treeview.FileInfo{
			FileInfo: info,
			Path:     oi.Path(),
		}) {
			continue
		}
		childNode := treeview.NewFileSystemNode(oi.Path(), info)
		var parentNode *treeview.Node[treeview.FileInfo]
		switch dir {
		case ".":
			parentNode = rootNode
		default:
			parentNode = m[dir]
		}
		if parentNode == nil {
			fmt.Println(dir)
			os.Exit(2)
		}
		parentNode.AddChild(childNode)
		cfg.HandleExpansion(childNode)
		total++
		cfg.ReportProgress(total, childNode)
	}
	return []*treeview.Node[treeview.FileInfo]{rootNode}, nil
}

func sortEntries(ctx context.Context, path string, profile string) ([]s3.ObjectInfo, []s3.ObjectInfo, error) {
	entries, err := s3.ListAllObjectsAndPrefixes(ctx, path, s3.WithProfile(profile))
	if err != nil {
		return nil, nil, pathError(treeview.ErrDirectoryScan, path, err)
	}
	var dir []s3.ObjectInfo
	var obj []s3.ObjectInfo
	for _, oi := range entries {
		switch oi.IsDir() {
		case true:
			dir = append(dir, oi)
		case false:
			obj = append(obj, oi)
		}
	}
	slices.SortFunc(dir, func(a, b s3.ObjectInfo) int {
		return strings.Compare(a.Path(), b.Path())
	})
	return dir, obj, nil
}
