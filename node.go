package treeview

import (
	"os"
)

// Node represents a single element in a tree. It stores an arbitrary payload
// of type T along with helper metadata used by the renderer and traversal
// helpers. A Node is not safe for concurrent mutation; callers must
// synchronise if they modify Nodes from multiple goroutines.
//
// ID and Name
//
//	ID   – stable, unique identifier used for look-ups and focus handling.
//	Name – optional human-readable label shown by default renderers.
//
// The zero value of Node is NOT ready for use; always create nodes via
// NewNode, NewNodeSimple, NewNodeClone, or NewFileSystemNode so the invariants are set up correctly.
type Node[T any] struct {
	id       string
	name     string
	data     T
	children []*Node[T]
	parent   *Node[T]
	expanded bool
	visible  bool
}

// NewNode constructs a Node with the supplied name and payload. Children are
// initially empty, the node is visible and collapsed.
func NewNode[T any](id, name string, data T) *Node[T] {
	return &Node[T]{
		id:       id,
		name:     name,
		data:     data,
		expanded: false,
		visible:  true,
	}
}

// NewNodeSimple constructs a Node using the same string for both ID and name.
// This is a convenience method for cases where the identifier and display name
// are identical. The node is initially visible and collapsed with empty children.
func NewNodeSimple[T any](idAndName string, data T) *Node[T] {
	return NewNode(idAndName, idAndName, data)
}

// NewNodeClone creates a copy of an existing node preserving its state (visible, expanded)
// but not its children or parent. This is useful when building filtered or depth-limited
// versions of trees where you need to preserve the original node's visual state.
func NewNodeClone[T any](original *Node[T]) *Node[T] {
	clone := NewNode(original.id, original.name, original.data)
	clone.visible = original.visible
	clone.expanded = original.expanded
	return clone
}

// ID returns the stable identifier of the node.
//
// The identifier is unique within a single tree and is the first port of call
// for operations like SetFocusedID, searches, and serialisation. Treat it as
// an immutable primary key.
func (n *Node[T]) ID() string {
	return n.id
}

// Name returns the human-readable label of the node. If you never set a
// distinct name the constructor falls back to using the ID so renderers can
// still show something sensible.
func (n *Node[T]) Name() string {
	if n.name == "" {
		return n.id
	}
	return n.name
}

// Data exposes the payload you supplied when creating the node. As with any
// generic accessor, callers should assert the concrete type they expect.
func (n *Node[T]) Data() T {
	return n.data
}

// SetData replaces the payload stored inside the node.
func (n *Node[T]) SetData(data T) {
	n.data = data
}

// Expand marks the node as expanded so its children become traversable and
// visible to renderers. No-op for leaf nodes.
func (n *Node[T]) Expand() {
	n.expanded = true
}

// Collapse hides all descendants of the node by flipping the expanded flag
// to false. Renderers and iterators will treat the node as a black box until
// you call Expand again.
func (n *Node[T]) Collapse() {
	n.expanded = false
}

// Toggle is a convenience wrapper that switches between Expand and Collapse
// depending on the current state.
func (n *Node[T]) Toggle() {
	n.expanded = !n.expanded
}

// SetExpanded sets the expanded state of the node.
func (n *Node[T]) SetExpanded(expanded bool) {
	n.expanded = expanded
}

// IsExpanded reports whether Expand was called without a matching Collapse.
func (n *Node[T]) IsExpanded() bool {
	return n.expanded
}

// SetVisible lets search and filter operations include or exclude the node
// from subsequent traversals without physically removing it from the tree.
func (n *Node[T]) SetVisible(visible bool) {
	n.visible = visible
}

// IsVisible returns the last value set via SetVisible.
func (n *Node[T]) IsVisible() bool {
	return n.visible
}

// HasChildren is a cheap helper for renderers and UIs that need to know if
// they should draw an expand/collapse icon next to the node.
func (n *Node[T]) HasChildren() bool {
	return len(n.children) > 0
}

// SetName updates the display label shown by renderers.
func (n *Node[T]) SetName(name string) {
	n.name = name
}

// Children returns the direct descendants of the node.
func (n *Node[T]) Children() []*Node[T] {
	return n.children
}

// SetChildren replaces the entire child slice and wires up the parent pointers.
func (n *Node[T]) SetChildren(children []*Node[T]) {
	for _, child := range children {
		child.parent = n
	}
	n.children = children
}

// AddChild appends a single child and sets the reciprocal parent pointer.
func (n *Node[T]) AddChild(child *Node[T]) {
	if child == nil {
		return
	}
	n.children = append(n.children, child)
	child.parent = n
}

// Parent returns the immediate ancestor or nil for root nodes.
func (n *Node[T]) Parent() *Node[T] {
	return n.parent
}

// FileInfo embeds os.FileInfo and adds the absolute Path so callers no longer
// need to juggle both pieces of information after a stat.
type FileInfo struct {
	os.FileInfo
	Path string
}

// NewFileSystemNode builds a FileSystemNode from a path and the corresponding
// os.FileInfo result. The function panics if info is nil because that would
// violate the contract of FileSystemNode.
func NewFileSystemNode(path string, info os.FileInfo) *Node[FileInfo] {
	fi := FileInfo{
		FileInfo: info,
		Path:     path,
	}
	return NewNode(path, info.Name(), fi)
}
