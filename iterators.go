package treeview

import (
	"context"
	"iter"
)

// NodeInfo is returned by iters and contains metadata about a node during iteration.
type NodeInfo[T any] struct {
	Node   *Node[T]
	Depth  int
	IsLast bool
}

// All returns an iterator that yields to every node using depth-first traversal.
// Context errors are returned unwrapped.
func (t *Tree[T]) All(ctx context.Context) iter.Seq2[NodeInfo[T], error] {
	return dfsSeq(ctx, t.Nodes(), true)
}

// All (From Node) - Iterate from specific Node using depth-first traversal.
// Context errors are returned unwrapped.
func (n *Node[T]) All(ctx context.Context) iter.Seq2[NodeInfo[T], error] {
	return dfsSeq(ctx, []*Node[T]{n}, true)
}

// AllVisible returns an iterator over visible nodes using depth-first traversal.
// Context errors are returned unwrapped.
func (t *Tree[T]) AllVisible(ctx context.Context) iter.Seq2[NodeInfo[T], error] {
	return func(yield func(NodeInfo[T], error) bool) {
		for info, err := range dfsSeq(ctx, t.Nodes(), false) {
			if err != nil {
				yield(NodeInfo[T]{}, err)
				return
			}
			if info.Node.IsVisible() {
				if !yield(info, nil) {
					return
				}
			}
		}
	}
}

// AllFocused returns an iterator over all focused nodes in the tree.
// Context errors are returned unwrapped.
func (t *Tree[T]) AllFocused(ctx context.Context) iter.Seq2[NodeInfo[T], error] {
	return func(yield func(NodeInfo[T], error) bool) {
		for info, err := range t.All(ctx) {
			if err != nil {
				yield(NodeInfo[T]{}, err)
				return
			}
			if t.IsFocused(info.Node.ID()) {
				if !yield(info, nil) {
					return
				}
			}
		}
	}
}

// BreadthFirst returns an iterator over nodes using breadth first traversal.
// Context errors are returned unwrapped.
func (t *Tree[T]) BreadthFirst(ctx context.Context) iter.Seq2[NodeInfo[T], error] {
	return bfsSeq(ctx, t.Nodes(), true)
}

// dfsSeq produces a depth-first iterator that yields NodeInfo values and context errors.
func dfsSeq[T any](ctx context.Context, roots []*Node[T], followUnexpanded bool) iter.Seq2[NodeInfo[T], error] {
	return func(yield func(NodeInfo[T], error) bool) {
		// Push roots in reverse order so we pop left-to-right.
		stack := make([]NodeInfo[T], 0, len(roots))
		for i := len(roots) - 1; i >= 0; i-- {
			stack = append(stack, NodeInfo[T]{Node: roots[i], Depth: 0, IsLast: i == len(roots)-1})
		}

		for len(stack) > 0 {
			// Check for context cancellation
			if err := ctx.Err(); err != nil {
				yield(NodeInfo[T]{}, err)
				return
			}

			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if !yield(NodeInfo[T]{Node: f.Node, Depth: f.Depth, IsLast: f.IsLast}, nil) {
				return
			}

			if f.Node.HasChildren() && (followUnexpanded || f.Node.IsExpanded()) {
				children := f.Node.Children()
				for i := len(children) - 1; i >= 0; i-- {
					stack = append(stack, NodeInfo[T]{Node: children[i], Depth: f.Depth + 1, IsLast: i == len(children)-1})
				}
			}
		}
	}
}

// bfsSeq produces a breadth-first iterator that yields NodeInfo values and context errors.
func bfsSeq[T any](ctx context.Context, roots []*Node[T], followUnexpanded bool) iter.Seq2[NodeInfo[T], error] {
	return func(yield func(NodeInfo[T], error) bool) {
		queue := make([]NodeInfo[T], 0, len(roots))
		for i, n := range roots {
			queue = append(queue, NodeInfo[T]{Node: n, Depth: 0, IsLast: i == len(roots)-1})
		}

		for len(queue) > 0 {
			if err := ctx.Err(); err != nil {
				yield(NodeInfo[T]{}, err)
				return
			}

			cur := queue[0]
			queue = queue[1:]

			if !yield(NodeInfo[T]{Node: cur.Node, Depth: cur.Depth, IsLast: cur.IsLast}, nil) {
				return
			}

			if cur.Node.HasChildren() && (followUnexpanded || cur.Node.IsExpanded()) {
				children := cur.Node.Children()
				for i, child := range children {
					queue = append(queue, NodeInfo[T]{Node: child, Depth: cur.Depth + 1, IsLast: i == len(children)-1})
				}
			}
		}
	}
}
