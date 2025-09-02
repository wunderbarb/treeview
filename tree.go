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
	mu           sync.RWMutex
	nodes        []*Node[T]
	focusedNodes []*Node[T]
	focusedIDs   map[string]bool

	searcher SearchFn[T]
	focusPol FocusPolicyFn[T]
	provider NodeProvider[T]

	// truncateWidth specifies the maximum width for rendered lines.
	// 0 means no truncation (default).
	truncateWidth int
}

// Nodes returns the current root slice. The caller must treat the returned
// slice as read-only unless they can guarantee exclusive access.
func (t *Tree[T]) Nodes() []*Node[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodes
}

// SetNodes replaces the root nodes of the tree. This is useful for removing
// root nodes or restructuring the tree. The operation is thread-safe.
func (t *Tree[T]) SetNodes(nodes []*Node[T]) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nodes = nodes
}

// GetFocusedID returns the ID of the currently focused node or "" if none.
func (t *Tree[T]) GetFocusedID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.focusedNodes) == 0 {
		return ""
	}
	return t.focusedNodes[0].ID()
}

// GetFocusedNode returns the currently focused node or nil if none is focused.
func (t *Tree[T]) GetFocusedNode() *Node[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.focusedNodes) == 0 {
		return nil
	}
	return t.focusedNodes[0]
}

// SetFocusedID focuses the node with the given ID. An empty ID clears the
// focus. The boolean return value indicates whether the focus actually
// changed. Returns ErrNodeNotFound if the ID doesn't exist, or context
// errors unwrapped.
func (t *Tree[T]) SetFocusedID(ctx context.Context, id string) (bool, error) {
	// Handle clearing focus when empty ID provided
	if id == "" {
		t.mu.Lock()
		defer t.mu.Unlock()
		changed := len(t.focusedNodes) > 0
		t.focusedNodes = nil
		t.focusedIDs = make(map[string]bool)
		return changed, nil
	}

	// First, try to find the node with the given ID
	node, err := t.FindByID(ctx, id)
	if err != nil {
		return false, err
	}

	// Now acquire the lock to update the focused node
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if we're already focused on this node (and only this node)
	if len(t.focusedNodes) == 1 && t.focusedNodes[0] == node {
		return false, nil
	}

	// Clear multi-focus and set single focus (backward compatible)
	t.focusedNodes = []*Node[T]{node}
	t.focusedIDs = map[string]bool{node.ID(): true}
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

	// Get the current primary focused node
	var currentFocus *Node[T]
	if len(t.focusedNodes) > 0 {
		currentFocus = t.focusedNodes[0]
	}

	// Use the configured focus policy to determine next node
	// The policy must handle edge cases like moving past boundaries
	next, err := t.focusPol(ctx, visible, currentFocus, offset)
	if err != nil {
		return false, err
	}

	// Update focus if it actually changed
	if next != currentFocus {
		// Clear multi-focus and set single focus (backward compatible)
		t.focusedNodes = []*Node[T]{next}
		t.focusedIDs = map[string]bool{next.ID(): true}
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

// ToggleFocused flips the expansion state of all focused nodes.
func (t *Tree[T]) ToggleFocused(ctx context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, node := range t.focusedNodes {
		node.Toggle()
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

	// Focus all matching nodes
	t.focusedNodes = make([]*Node[T], len(matches))
	copy(t.focusedNodes, matches)
	t.focusedIDs = make(map[string]bool, len(matches))
	for _, match := range matches {
		t.focusedIDs[match.ID()] = true
	}
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

// GetAllFocusedIDs returns the IDs of all currently focused nodes.
func (t *Tree[T]) GetAllFocusedIDs() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, len(t.focusedNodes))
	for i, node := range t.focusedNodes {
		ids[i] = node.ID()
	}
	return ids
}

// GetAllFocusedNodes returns all currently focused nodes.
func (t *Tree[T]) GetAllFocusedNodes() []*Node[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	nodes := make([]*Node[T], len(t.focusedNodes))
	copy(nodes, t.focusedNodes)
	return nodes
}

// IsFocused checks if the node with the given ID is currently focused.
func (t *Tree[T]) IsFocused(id string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focusedIDs[id]
}

// AddFocusedID adds a node to the focused set. Returns ErrNodeNotFound if the ID doesn't exist.
func (t *Tree[T]) AddFocusedID(ctx context.Context, id string) error {
	// Find the node first
	node, err := t.FindByID(ctx, id)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if already focused
	if t.focusedIDs[id] {
		return nil // Already focused, no change needed
	}

	// Add to focused set
	t.focusedNodes = append(t.focusedNodes, node)
	if t.focusedIDs == nil {
		t.focusedIDs = make(map[string]bool)
	}
	t.focusedIDs[id] = true
	return nil
}

// RemoveFocusedID removes a node from the focused set.
func (t *Tree[T]) RemoveFocusedID(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if the node is focused
	if !t.focusedIDs[id] {
		return nil // Not focused, no change needed
	}

	// Remove from focused IDs map
	delete(t.focusedIDs, id)

	// Remove from focused nodes slice
	for i, node := range t.focusedNodes {
		if node.ID() == id {
			t.focusedNodes = append(t.focusedNodes[:i], t.focusedNodes[i+1:]...)
			break
		}
	}
	return nil
}

// SetAllFocusedIDs sets the complete list of focused node IDs, replacing any existing focus.
func (t *Tree[T]) SetAllFocusedIDs(ctx context.Context, ids []string) error {
	// Find all nodes first to validate they exist
	nodes := make([]*Node[T], 0, len(ids))
	for _, id := range ids {
		node, err := t.FindByID(ctx, id)
		if err != nil {
			return err
		}
		nodes = append(nodes, node)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Replace focused state
	t.focusedNodes = nodes
	t.focusedIDs = make(map[string]bool, len(ids))
	for _, id := range ids {
		t.focusedIDs[id] = true
	}
	return nil
}

// ClearAllFocus clears all focused nodes.
func (t *Tree[T]) ClearAllFocus() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.focusedNodes = nil
	t.focusedIDs = make(map[string]bool)
}

// ToggleFocusedID toggles the focus state of the node with the given ID.
func (t *Tree[T]) ToggleFocusedID(ctx context.Context, id string) error {
	t.mu.RLock()
	isFocused := t.focusedIDs[id]
	t.mu.RUnlock()

	if isFocused {
		return t.RemoveFocusedID(ctx, id)
	} else {
		return t.AddFocusedID(ctx, id)
	}
}

// MoveExtend moves the primary focus by offset while preserving the existing multi-focus.
// If there's no existing multi-focus, it behaves like regular Move.
// The bool result reports whether focus moved.
func (t *Tree[T]) MoveExtend(ctx context.Context, offset int) (bool, error) {
	// Get all currently visible nodes first (before locking)
	visible, err := t.visibleNodes(ctx)
	if err != nil {
		return false, err
	}

	// Lock for reading and updating focus
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get the current primary focused node
	var currentFocus *Node[T]
	if len(t.focusedNodes) > 0 {
		currentFocus = t.focusedNodes[0]
	}

	// Use the configured focus policy to determine next node
	next, err := t.focusPol(ctx, visible, currentFocus, offset)
	if err != nil {
		return false, err
	}

	// If no change in focus, return false
	if next == currentFocus {
		return false, nil
	}

	// If we have no existing multi-focus, start with the current node
	if len(t.focusedNodes) == 0 {
		if next != nil {
			t.focusedNodes = []*Node[T]{next}
			t.focusedIDs = map[string]bool{next.ID(): true}
		}
		return true, nil
	}

	// Find the range between current primary focus and the new target
	rangeNodes, err := t.findNodeRange(visible, currentFocus, next)
	if err != nil {
		return false, err
	}

	// Add all nodes in the range to focus (if not already focused)
	for _, node := range rangeNodes {
		if !t.focusedIDs[node.ID()] {
			t.focusedNodes = append(t.focusedNodes, node)
			t.focusedIDs[node.ID()] = true
		}
	}

	// Update the primary focus to the new target by moving it to the front
	// Remove the new target from its current position and add to front
	for i, node := range t.focusedNodes {
		if node == next {
			// Move to front
			t.focusedNodes = append([]*Node[T]{next}, append(t.focusedNodes[:i], t.focusedNodes[i+1:]...)...)
			break
		}
	}

	return true, nil
}

// findNodeRange finds all visible nodes between start and end (inclusive).
func (t *Tree[T]) findNodeRange(visible []*Node[T], start, end *Node[T]) ([]*Node[T], error) {
	if start == nil || end == nil {
		return nil, nil
	}

	// Find indices of start and end nodes in visible list
	startIdx, endIdx := -1, -1
	for i, node := range visible {
		if node == start {
			startIdx = i
		}
		if node == end {
			endIdx = i
		}
	}

	if startIdx == -1 || endIdx == -1 {
		return nil, nil
	}

	// Ensure we have the range in the correct order
	if startIdx > endIdx {
		startIdx, endIdx = endIdx, startIdx
	}

	// Return the range of nodes (inclusive)
	return visible[startIdx : endIdx+1], nil
}
