// Package treeview provides a flexible tree structure for building interactive
// terminal user interfaces. You can create trees from various data sources,
// navigate with keyboard controls, search and filter nodes, and customize
// rendering with your own styling.
//
// The core [Tree] type manages a collection of [Node] instances and provides
// thread-safe operations for focusing, expanding, searching, and rendering.
// Trees support context cancellation and can be customized through
// functional options.
//
// Basic usage:
//
//	tree := treeview.New(
//		treeview.WithSearcher(mySearchFn),
//	)
//	renderedTree, err := tree.Render(ctx)
package treeview

import (
	"context"
	"sync"
)

// Tree wraps a collection of nodes and offers rich operations such as
// filtering, searching, focusing, and rendering. All public methods are safe
// for concurrent use; a sync.RWMutex guards internal state.
type Tree[T any] struct {
	mu      sync.RWMutex
	nodes   []*Node[T]
	focused *Node[T]

	searcher SearchFn[T]
	focusPol FocusPolicyFn[T]
	provider NodeProvider[T]
}

// Nodes returns the current root slice. The caller must treat the returned
// slice as read-only unless they can guarantee exclusive access.
func (t *Tree[T]) Nodes() []*Node[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodes
}

// GetFocusedID returns the ID of the currently focused node or "" if none.
func (t *Tree[T]) GetFocusedID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.focused == nil {
		return ""
	}
	return t.focused.ID()
}

// GetFocusedNode returns the currently focused node or nil if none is focused.
func (t *Tree[T]) GetFocusedNode() *Node[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

// SetFocusedID focuses the node with the given ID. An empty ID clears the
// focus. The boolean return value indicates whether the focus actually
// changed. Returns ErrNodeNotFound if the ID doesn't exist, or context
// errors unwrapped.
func (t *Tree[T]) SetFocusedID(ctx context.Context, id string) (bool, error) {
	// Handle clearing focus when empty ID provided
	if id == "" {
		changed := t.focused != nil
		t.focused = nil
		return changed, nil
	}

	// First, try to find the node with the given ID
	node, err := t.FindByID(ctx, id)
	if err != nil {
		return false, err
	}

	// Check if we're already focused on this node
	if node == t.focused {
		return false, nil
	}

	// Now acquire the lock to update the focused node
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update focus and report that it changed
	t.focused = node
	return true, nil
}

// FindByID searches the tree for a node with the given ID. Returns
// ErrNodeNotFound if no node matches, or context errors unwrapped.
func (t *Tree[T]) FindByID(ctx context.Context, id string) (*Node[T], error) {
	// Use the iterator with context support
	for info, err := range t.All(ctx) {
		if err != nil {
			return nil, err
		}
		if info.Node.ID() == id {
			return info.Node, nil
		}
	}

	return nil, ErrNodeNotFound
}

// Move changes the focus by offset within the list of currently visible nodes.
// Negative values move upward, positive downward. The bool result reports
// whether focus moved. Returns context errors unwrapped.
func (t *Tree[T]) Move(ctx context.Context, offset int) (bool, error) {
	// Get all currently visible nodes first (before locking)
	// These are the nodes we can navigate between
	visible, err := t.visibleNodes(ctx)
	if err != nil {
		return false, err
	}

	// Lock only for reading current focus and updating it
	t.mu.Lock()
	defer t.mu.Unlock()

	// Use the configured focus policy to determine next node
	// The policy must handle edge cases like moving past boundaries
	next, err := t.focusPol(ctx, visible, t.focused, offset)
	if err != nil {
		return false, err
	}

	// Update focus if it actually changed
	if next != t.focused {
		t.focused = next
		return true, nil
	}

	// No change in focus
	return false, nil
}

// SetExpanded sets the expanded state of the given node. It returns false if
// the node wasn't found. Returns ErrNodeNotFound if the ID doesn't exist, or
// context errors unwrapped.
func (t *Tree[T]) SetExpanded(ctx context.Context, id string, expanded bool) (bool, error) {
	// Find the node to expand/collapse
	node, err := t.FindByID(ctx, id)
	if err != nil {
		return false, err // Node not found or context cancelled
	}

	// Apply the requested expansion state
	if expanded {
		node.Expand()
	} else {
		node.Collapse()
	}
	return true, nil
}

// ToggleFocused flips the expansion state of the focused node.
func (t *Tree[T]) ToggleFocused(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.focused != nil {
		t.focused.Toggle()
	}
}

// ExpandAll expands every node in the tree. Returns context errors unwrapped.
func (t *Tree[T]) ExpandAll(ctx context.Context) error {
	return t.setExpandedState(ctx, true)
}

// CollapseAll collapses every node. Returns context errors unwrapped.
func (t *Tree[T]) CollapseAll(ctx context.Context) error {
	return t.setExpandedState(ctx, false)
}

// ShowAll shows all nodes. Returns context errors unwrapped.
func (t *Tree[T]) ShowAll(ctx context.Context) error {
	return t.setVisibleState(ctx, true)
}

// HideAll hides all nodes. Returns context errors unwrapped.
func (t *Tree[T]) HideAll(ctx context.Context) error {
	return t.setVisibleState(ctx, false)
}

// Search scans the tree for nodes matching term. Returns context errors unwrapped.
func (t *Tree[T]) Search(ctx context.Context, term string) ([]*Node[T], error) {
	// Empty search term returns no results
	if term == "" {
		return nil, nil
	}

	// Collect all matching nodes using iterator
	var results []*Node[T]
	for info, err := range t.All(ctx) {
		if err != nil {
			return results, err
		}
		// Use the configured searcher to check for matches
		// This allows custom search logic per tree instance
		if t.searcher(ctx, info.Node, term) {
			results = append(results, info.Node)
		}
	}

	return results, nil
}

// SearchAndExpand performs Search and then expands all ancestor nodes of each
// match so the results become visible. The first match is focused.
// Returns context errors unwrapped.
func (t *Tree[T]) SearchAndExpand(ctx context.Context, term string) ([]*Node[T], error) {
	// Empty search term means clear search but stay in search mode
	if term == "" {
		return nil, nil
	}

	// Find all matching nodes
	matches, err := t.Search(ctx, term)
	if err != nil {
		return matches, err
	}

	// If no matches, expand all and return
	if len(matches) == 0 {
		// Just clear search highlights but don't change expansion state
		if err := t.ShowAll(ctx); err != nil {
			return nil, err
		}
		if err := t.ExpandAll(ctx); err != nil {
			return nil, err
		}
		return matches, nil
	}

	// Collapse and hide all nodes before acquiring lock
	if err := t.CollapseAll(ctx); err != nil {
		return nil, err
	}
	if err := t.HideAll(ctx); err != nil {
		return nil, err
	}

	// Lock only for the final expansion and focus operations
	t.mu.Lock()
	defer t.mu.Unlock()

	// Mark all matches and their ancestors for expansion
	for _, match := range matches {
		current := match
		for current != nil {
			current.Expand()
			current.SetVisible(true)
			current = current.Parent()
		}
	}

	// Focus the first match to position the view
	t.focused = matches[0]
	return matches, nil
}

// Render produces a string representation of the tree using the configured
// renderer. Returns context errors unwrapped.
func (t *Tree[T]) Render(ctx context.Context) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	output, _, err := renderTree(ctx, t)
	return output, err
}

// visibleNodes returns all nodes that are currently visible.
// This is used for navigation and focus management.
func (t *Tree[T]) visibleNodes(ctx context.Context) ([]*Node[T], error) {
	// Use iterator to collect visible nodes
	var nodes []*Node[T]
	for info, err := range t.AllVisible(ctx) {
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, info.Node)
	}
	return nodes, nil
}

// setExpandedState is a helper method used by ExpandAll and CollapseAll.
func (t *Tree[T]) setExpandedState(ctx context.Context, expanded bool) error {
	for info, err := range t.All(ctx) {
		if err != nil {
			return err
		}
		info.Node.SetExpanded(expanded)
	}
	return nil
}

func (t *Tree[T]) setVisibleState(ctx context.Context, visible bool) error {
	for info, err := range t.All(ctx) {
		if err != nil {
			return err
		}
		info.Node.SetVisible(visible)
	}
	return nil
}
