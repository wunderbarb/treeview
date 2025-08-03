package treeview

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Helper function to create a test tree structure for iterators
func createTestTreeForIterators[T any](t *testing.T) *Tree[T] {
	t.Helper()
	tree := &Tree[T]{}
	return tree
}

// Helper function to create a simple tree with string data
func createSimpleTree(t *testing.T) *Tree[string] {
	t.Helper()
	tree := &Tree[string]{}

	// Create nodes
	root1 := NewNode("1", "Node 1", "data1")
	root2 := NewNode("2", "Node 2", "data2")

	child1_1 := NewNode("1.1", "Node 1.1", "data1.1")
	child1_2 := NewNode("1.2", "Node 1.2", "data1.2")
	child1_1_1 := NewNode("1.1.1", "Node 1.1.1", "data1.1.1")

	child2_1 := NewNode("2.1", "Node 2.1", "data2.1")

	// Build tree structure
	root1.AddChild(child1_1)
	root1.AddChild(child1_2)
	child1_1.AddChild(child1_1_1)
	root2.AddChild(child2_1)

	// Expand some nodes
	root1.Expand()
	child1_1.Expand()

	tree.nodes = []*Node[string]{root1, root2}
	return tree
}

// Helper to collect NodeInfo from an iterator
func collectNodes[T any](t *testing.T, iter iter.Seq2[NodeInfo[T], error]) ([]NodeInfo[T], error) {
	t.Helper()
	var results []NodeInfo[T]
	for info, err := range iter {
		if err != nil {
			return results, err
		}
		results = append(results, info)
	}
	return results, nil
}

// Helper to extract IDs from NodeInfo slice
func extractIDs[T any](infos []NodeInfo[T]) []string {
	ids := make([]string, len(infos))
	for i, info := range infos {
		ids[i] = info.Node.ID()
	}
	return ids
}

func TestTree_All(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) *Tree[string]
		wantIDs    []string
		wantDepths []int
	}{
		{
			name: "empty_tree",
			setup: func(t *testing.T) *Tree[string] {
				return createTestTreeForIterators[string](t)
			},
			wantIDs:    []string{},
			wantDepths: []int{},
		},
		{
			name: "single_root",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				tree.nodes = []*Node[string]{NewNode("1", "Root", "data")}
				return tree
			},
			wantIDs:    []string{"1"},
			wantDepths: []int{0},
		},
		{
			name: "multiple_roots",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				tree.nodes = []*Node[string]{
					NewNode("1", "Root 1", "data1"),
					NewNode("2", "Root 2", "data2"),
					NewNode("3", "Root 3", "data3"),
				}
				return tree
			},
			wantIDs:    []string{"1", "2", "3"},
			wantDepths: []int{0, 0, 0},
		},
		{
			name:       "depth_first_traversal",
			setup:      createSimpleTree,
			wantIDs:    []string{"1", "1.1", "1.1.1", "1.2", "2", "2.1"},
			wantDepths: []int{0, 1, 2, 1, 0, 1},
		},
		{
			name: "unexpanded_nodes_still_visited",
			setup: func(t *testing.T) *Tree[string] {
				tree := createSimpleTree(t)
				// Collapse all nodes
				for info := range tree.All(context.Background()) {
					info.Node.Collapse()
				}
				return tree
			},
			wantIDs:    []string{"1", "1.1", "1.1.1", "1.2", "2", "2.1"},
			wantDepths: []int{0, 1, 2, 1, 0, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := test.setup(t)
			ctx := context.Background()

			infos, err := collectNodes(t, tree.All(ctx))
			if err != nil {
				t.Errorf("Tree.All() error = %v", err)
				return
			}

			gotIDs := extractIDs(infos)
			if diff := cmp.Diff(test.wantIDs, gotIDs); diff != "" {
				t.Errorf("Tree.All() IDs mismatch (-want +got):\n%s", diff)
			}

			// Check depths
			gotDepths := make([]int, len(infos))
			for i, info := range infos {
				gotDepths[i] = info.Depth
			}
			if diff := cmp.Diff(test.wantDepths, gotDepths); diff != "" {
				t.Errorf("Tree.All() depths mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNode_All(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) *Node[string]
		wantIDs    []string
		wantDepths []int
	}{
		{
			name: "single_node_no_children",
			setup: func(t *testing.T) *Node[string] {
				return NewNode("1", "Node", "data")
			},
			wantIDs:    []string{"1"},
			wantDepths: []int{0},
		},
		{
			name: "node_with_children",
			setup: func(t *testing.T) *Node[string] {
				root := NewNode("1", "Root", "data")
				child1 := NewNode("1.1", "Child 1", "data1.1")
				child2 := NewNode("1.2", "Child 2", "data1.2")
				grandchild := NewNode("1.1.1", "Grandchild", "data1.1.1")

				root.AddChild(child1)
				root.AddChild(child2)
				child1.AddChild(grandchild)

				root.Expand()
				child1.Expand()

				return root
			},
			wantIDs:    []string{"1", "1.1", "1.1.1", "1.2"},
			wantDepths: []int{0, 1, 2, 1},
		},
		{
			name: "subtree_iteration",
			setup: func(t *testing.T) *Node[string] {
				// Create full tree but return a subtree node
				tree := createSimpleTree(t)
				// Find node 1.1
				for info := range tree.All(context.Background()) {
					if info.Node.ID() == "1.1" {
						return info.Node
					}
				}
				t.Fatal("Failed to find node 1.1")
				return nil
			},
			wantIDs:    []string{"1.1", "1.1.1"},
			wantDepths: []int{0, 1}, // Depths are relative to starting node
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := test.setup(t)
			ctx := context.Background()

			infos, err := collectNodes(t, node.All(ctx))
			if err != nil {
				t.Errorf("Node.All() error = %v", err)
				return
			}

			gotIDs := extractIDs(infos)
			if diff := cmp.Diff(test.wantIDs, gotIDs); diff != "" {
				t.Errorf("Node.All() IDs mismatch (-want +got):\n%s", diff)
			}

			// Check depths
			gotDepths := make([]int, len(infos))
			for i, info := range infos {
				gotDepths[i] = info.Depth
			}
			if diff := cmp.Diff(test.wantDepths, gotDepths); diff != "" {
				t.Errorf("Node.All() depths mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTree_AllVisible(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *Tree[string]
		wantIDs []string
	}{
		{
			name: "all_nodes_visible",
			setup: func(t *testing.T) *Tree[string] {
				return createSimpleTree(t)
			},
			wantIDs: []string{"1", "1.1", "1.1.1", "1.2", "2"}, // 2.1 not visible because 2 is not expanded
		},
		{
			name: "some_nodes_hidden",
			setup: func(t *testing.T) *Tree[string] {
				tree := createSimpleTree(t)
				// Hide specific nodes
				for info := range tree.All(context.Background()) {
					if info.Node.ID() == "1.1" || info.Node.ID() == "2" {
						info.Node.SetVisible(false)
					}
				}
				return tree
			},
			wantIDs: []string{"1", "1.1.1", "1.2"}, // 1.1 and 2 are hidden, 2.1 not visited because 2 is not expanded
		},
		{
			name: "all_nodes_hidden",
			setup: func(t *testing.T) *Tree[string] {
				tree := createSimpleTree(t)
				// Hide all nodes
				for info := range tree.All(context.Background()) {
					info.Node.SetVisible(false)
				}
				return tree
			},
			wantIDs: []string{},
		},
		{
			name: "hidden_parent_visible_children",
			setup: func(t *testing.T) *Tree[string] {
				tree := createSimpleTree(t)
				// Hide only root 1
				tree.nodes[0].SetVisible(false)
				return tree
			},
			wantIDs: []string{"1.1", "1.1.1", "1.2", "2"}, // Root 1 hidden but children visible, 2.1 not visited because 2 is not expanded
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := test.setup(t)
			ctx := context.Background()

			infos, err := collectNodes(t, tree.AllVisible(ctx))
			if err != nil {
				t.Errorf("Tree.AllVisible() error = %v", err)
				return
			}

			gotIDs := extractIDs(infos)
			if diff := cmp.Diff(test.wantIDs, gotIDs); diff != "" {
				t.Errorf("Tree.AllVisible() IDs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTree_BreadthFirst(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) *Tree[string]
		wantIDs    []string
		wantDepths []int
	}{
		{
			name: "empty_tree",
			setup: func(t *testing.T) *Tree[string] {
				return createTestTreeForIterators[string](t)
			},
			wantIDs:    []string{},
			wantDepths: []int{},
		},
		{
			name:       "breadth_first_order",
			setup:      createSimpleTree,
			wantIDs:    []string{"1", "2", "1.1", "1.2", "2.1", "1.1.1"},
			wantDepths: []int{0, 0, 1, 1, 1, 2},
		},
		{
			name: "single_branch",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				root := NewNode("1", "Root", "data")
				child := NewNode("2", "Child", "data")
				grandchild := NewNode("3", "Grandchild", "data")

				root.AddChild(child)
				child.AddChild(grandchild)
				root.Expand()
				child.Expand()

				tree.nodes = []*Node[string]{root}
				return tree
			},
			wantIDs:    []string{"1", "2", "3"},
			wantDepths: []int{0, 1, 2},
		},
		{
			name: "multiple_roots_breadth_first",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				root1 := NewNode("A", "Root A", "dataA")
				root2 := NewNode("B", "Root B", "dataB")
				root3 := NewNode("C", "Root C", "dataC")

				childA1 := NewNode("A1", "Child A1", "dataA1")
				childA2 := NewNode("A2", "Child A2", "dataA2")
				childB1 := NewNode("B1", "Child B1", "dataB1")

				root1.AddChild(childA1)
				root1.AddChild(childA2)
				root2.AddChild(childB1)

				root1.Expand()
				root2.Expand()

				tree.nodes = []*Node[string]{root1, root2, root3}
				return tree
			},
			wantIDs:    []string{"A", "B", "C", "A1", "A2", "B1"},
			wantDepths: []int{0, 0, 0, 1, 1, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := test.setup(t)
			ctx := context.Background()

			infos, err := collectNodes(t, tree.BreadthFirst(ctx))
			if err != nil {
				t.Errorf("Tree.BreadthFirst() error = %v", err)
				return
			}

			gotIDs := extractIDs(infos)
			if diff := cmp.Diff(test.wantIDs, gotIDs); diff != "" {
				t.Errorf("Tree.BreadthFirst() IDs mismatch (-want +got):\n%s", diff)
			}

			// Check depths
			gotDepths := make([]int, len(infos))
			for i, info := range infos {
				gotDepths[i] = info.Depth
			}
			if diff := cmp.Diff(test.wantDepths, gotDepths); diff != "" {
				t.Errorf("Tree.BreadthFirst() depths mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTree_AllBottomUp(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) *Tree[string]
		wantIDs    []string
		wantDepths []int
	}{
		{
			name: "empty_tree",
			setup: func(t *testing.T) *Tree[string] {
				return createTestTreeForIterators[string](t)
			},
			wantIDs:    []string{},
			wantDepths: []int{},
		},
		{
			name:  "bottom_up_order",
			setup: createSimpleTree,
			// Bottom-up should visit leaves first, then parents
			wantIDs:    []string{"1.1.1", "1.1", "1.2", "1", "2.1", "2"},
			wantDepths: []int{2, 1, 1, 0, 1, 0},
		},
		{
			name: "single_branch",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				root := NewNode("1", "Root", "data")
				child := NewNode("2", "Child", "data")
				grandchild := NewNode("3", "Grandchild", "data")

				root.AddChild(child)
				child.AddChild(grandchild)
				root.Expand()
				child.Expand()

				tree.nodes = []*Node[string]{root}
				return tree
			},
			wantIDs:    []string{"3", "2", "1"},
			wantDepths: []int{2, 1, 0},
		},
		{
			name: "multiple_roots_bottom_up",
			setup: func(t *testing.T) *Tree[string] {
				tree := createTestTreeForIterators[string](t)
				root1 := NewNode("A", "Root A", "dataA")
				root2 := NewNode("B", "Root B", "dataB")
				root3 := NewNode("C", "Root C", "dataC")

				childA1 := NewNode("A1", "Child A1", "dataA1")
				childA2 := NewNode("A2", "Child A2", "dataA2")
				childB1 := NewNode("B1", "Child B1", "dataB1")

				root1.AddChild(childA1)
				root1.AddChild(childA2)
				root2.AddChild(childB1)

				root1.Expand()
				root2.Expand()

				tree.nodes = []*Node[string]{root1, root2, root3}
				return tree
			},
			// Bottom-up: leaves first, then parents in order of roots
			wantIDs:    []string{"A1", "A2", "A", "B1", "B", "C"},
			wantDepths: []int{1, 1, 0, 1, 0, 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := test.setup(t)
			ctx := context.Background()

			infos, err := collectNodes(t, tree.AllBottomUp(ctx))
			if err != nil {
				t.Errorf("Tree.AllBottomUp() error = %v", err)
				return
			}

			gotIDs := extractIDs(infos)
			if diff := cmp.Diff(test.wantIDs, gotIDs); diff != "" {
				t.Errorf("Tree.AllBottomUp() IDs mismatch (-want +got):\n%s", diff)
			}

			// Check depths
			gotDepths := make([]int, len(infos))
			for i, info := range infos {
				gotDepths[i] = info.Depth
			}
			if diff := cmp.Diff(test.wantDepths, gotDepths); diff != "" {
				t.Errorf("Tree.AllBottomUp() depths mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIterators_ContextCancellation(t *testing.T) {
	tests := []struct {
		name     string
		iterator func(*Tree[string], context.Context) iter.Seq2[NodeInfo[string], error]
	}{
		{
			name: "Tree.All",
			iterator: func(tree *Tree[string], ctx context.Context) iter.Seq2[NodeInfo[string], error] {
				return tree.All(ctx)
			},
		},
		{
			name: "Tree.AllVisible",
			iterator: func(tree *Tree[string], ctx context.Context) iter.Seq2[NodeInfo[string], error] {
				return tree.AllVisible(ctx)
			},
		},
		{
			name: "Tree.BreadthFirst",
			iterator: func(tree *Tree[string], ctx context.Context) iter.Seq2[NodeInfo[string], error] {
				return tree.BreadthFirst(ctx)
			},
		},
		{
			name: "Tree.AllBottomUp",
			iterator: func(tree *Tree[string], ctx context.Context) iter.Seq2[NodeInfo[string], error] {
				return tree.AllBottomUp(ctx)
			},
		},
		{
			name: "Node.All",
			iterator: func(tree *Tree[string], ctx context.Context) iter.Seq2[NodeInfo[string], error] {
				if len(tree.nodes) > 0 {
					return tree.nodes[0].All(ctx)
				}
				// Return empty iterator if no nodes
				return func(yield func(NodeInfo[string], error) bool) {}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := createSimpleTree(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var count int
			var gotErr error

			for _, err := range test.iterator(tree, ctx) {
				if err != nil {
					gotErr = err
					break
				}
				count++
				// Cancel after first node
				if count == 1 {
					cancel()
				}
			}

			if !errors.Is(gotErr, context.Canceled) {
				t.Errorf("%s context cancellation: got error %v, want context.Canceled", test.name, gotErr)
			}

			if count > 2 {
				t.Errorf("%s processed too many nodes after cancellation: got %d, want <= 2", test.name, count)
			}
		})
	}
}

func TestIterators_IsLast(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) *Tree[string]
		iterator func(*Tree[string]) iter.Seq2[NodeInfo[string], error]
		wantLast map[string]bool // ID -> IsLast
	}{
		{
			name:  "Tree.All_IsLast",
			setup: createSimpleTree,
			iterator: func(tree *Tree[string]) iter.Seq2[NodeInfo[string], error] {
				return tree.All(context.Background())
			},
			wantLast: map[string]bool{
				"1":     false, // Not last root
				"1.1":   false, // Not last child of 1
				"1.1.1": true,  // Last (only) child of 1.1
				"1.2":   true,  // Last child of 1
				"2":     true,  // Last root
				"2.1":   true,  // Last (only) child of 2
			},
		},
		{
			name:  "Tree.BreadthFirst_IsLast",
			setup: createSimpleTree,
			iterator: func(tree *Tree[string]) iter.Seq2[NodeInfo[string], error] {
				return tree.BreadthFirst(context.Background())
			},
			wantLast: map[string]bool{
				"1":     false, // Not last root
				"2":     true,  // Last root
				"1.1":   false, // Not last child of 1
				"1.2":   true,  // Last child of 1
				"2.1":   true,  // Last (only) child of 2
				"1.1.1": true,  // Last (only) child of 1.1
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := test.setup(t)

			for info, err := range test.iterator(tree) {
				if err != nil {
					t.Errorf("%s error = %v", test.name, err)
					return
				}

				wantIsLast, ok := test.wantLast[info.Node.ID()]
				if !ok {
					t.Errorf("%s: unexpected node ID %s", test.name, info.Node.ID())
					continue
				}

				if info.IsLast != wantIsLast {
					t.Errorf("%s node %s: IsLast = %v, want %v", test.name, info.Node.ID(), info.IsLast, wantIsLast)
				}
			}
		})
	}
}

func TestIterators_EarlyTermination(t *testing.T) {
	tree := createSimpleTree(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		iterator  func() iter.Seq2[NodeInfo[string], error]
		stopAfter int
	}{
		{
			name:      "Tree.All_early_stop",
			iterator:  func() iter.Seq2[NodeInfo[string], error] { return tree.All(ctx) },
			stopAfter: 3,
		},
		{
			name:      "Tree.AllVisible_early_stop",
			iterator:  func() iter.Seq2[NodeInfo[string], error] { return tree.AllVisible(ctx) },
			stopAfter: 2,
		},
		{
			name:      "Tree.BreadthFirst_early_stop",
			iterator:  func() iter.Seq2[NodeInfo[string], error] { return tree.BreadthFirst(ctx) },
			stopAfter: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count := 0
			for _, err := range test.iterator() {
				if err != nil {
					t.Errorf("%s unexpected error: %v", test.name, err)
					return
				}
				count++
				if count >= test.stopAfter {
					break // Early termination
				}
			}

			if count != test.stopAfter {
				t.Errorf("%s: got %d iterations, want %d", test.name, count, test.stopAfter)
			}
		})
	}
}

func TestIterators_FollowUnexpanded(t *testing.T) {
	setup := func(t *testing.T) *Tree[string] {
		tree := createSimpleTree(t)
		// Collapse all nodes
		for info := range tree.All(context.Background()) {
			info.Node.Collapse()
		}
		return tree
	}

	t.Run("Tree.All_follows_unexpanded", func(t *testing.T) {
		tree := setup(t)
		ctx := context.Background()

		infos, err := collectNodes(t, tree.All(ctx))
		if err != nil {
			t.Errorf("Tree.All() error = %v", err)
			return
		}

		// Should visit all nodes even when collapsed
		wantIDs := []string{"1", "1.1", "1.1.1", "1.2", "2", "2.1"}
		gotIDs := extractIDs(infos)
		if diff := cmp.Diff(wantIDs, gotIDs); diff != "" {
			t.Errorf("Tree.All() with collapsed nodes IDs mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Tree.BreadthFirst_follows_unexpanded", func(t *testing.T) {
		tree := setup(t)
		ctx := context.Background()

		infos, err := collectNodes(t, tree.BreadthFirst(ctx))
		if err != nil {
			t.Errorf("Tree.BreadthFirst() error = %v", err)
			return
		}

		// Should visit all nodes even when collapsed
		wantIDs := []string{"1", "2", "1.1", "1.2", "2.1", "1.1.1"}
		gotIDs := extractIDs(infos)
		if diff := cmp.Diff(wantIDs, gotIDs); diff != "" {
			t.Errorf("Tree.BreadthFirst() with collapsed nodes IDs mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestIterators_LargeTree(t *testing.T) {
	// Create a larger tree to test performance characteristics
	createLargeTree := func(t *testing.T) *Tree[int] {
		tree := createTestTreeForIterators[int](t)
		roots := make([]*Node[int], 10)

		for i := 0; i < 10; i++ {
			roots[i] = NewNode(string(rune('A'+i)), string(rune('A'+i)), i)
			for j := 0; j < 5; j++ {
				child := NewNode(string(rune('A'+i))+string(rune('0'+j)), "Child", i*10+j)
				roots[i].AddChild(child)
				for k := 0; k < 3; k++ {
					grandchild := NewNode(string(rune('A'+i))+string(rune('0'+j))+string(rune('a'+k)), "Grandchild", i*100+j*10+k)
					child.AddChild(grandchild)
				}
			}
			roots[i].Expand()
		}

		tree.nodes = roots
		return tree
	}

	t.Run("large_tree_depth_first", func(t *testing.T) {
		tree := createLargeTree(t)
		ctx := context.Background()

		count := 0
		maxDepth := 0
		for info, err := range tree.All(ctx) {
			if err != nil {
				t.Errorf("Tree.All() on large tree error = %v", err)
				return
			}
			count++
			if info.Depth > maxDepth {
				maxDepth = info.Depth
			}
		}

		wantCount := 10 + 10*5 + 10*5*3 // 10 roots + 50 children + 150 grandchildren
		if count != wantCount {
			t.Errorf("Tree.All() on large tree: got %d nodes, want %d", count, wantCount)
		}

		if maxDepth != 2 {
			t.Errorf("Tree.All() on large tree: got max depth %d, want 2", maxDepth)
		}
	})

	t.Run("large_tree_breadth_first", func(t *testing.T) {
		tree := createLargeTree(t)
		ctx := context.Background()

		count := 0
		lastDepth := -1
		for info, err := range tree.BreadthFirst(ctx) {
			if err != nil {
				t.Errorf("Tree.BreadthFirst() on large tree error = %v", err)
				return
			}
			// Verify breadth-first order: depth should never decrease
			if info.Depth < lastDepth {
				t.Errorf("Tree.BreadthFirst() depth decreased from %d to %d at node %s", lastDepth, info.Depth, info.Node.ID())
			}
			lastDepth = info.Depth
			count++
		}

		wantCount := 10 + 10*5 + 10*5*3
		if count != wantCount {
			t.Errorf("Tree.BreadthFirst() on large tree: got %d nodes, want %d", count, wantCount)
		}
	})
}

func TestIterators_ContextTimeout(t *testing.T) {
	tree := createSimpleTree(t)

	// Create a context that times out quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give timeout a chance to trigger
	time.Sleep(10 * time.Millisecond)

	var gotErr error
	count := 0
	for _, err := range tree.All(ctx) {
		if err != nil {
			gotErr = err
			break
		}
		count++
	}

	if !errors.Is(gotErr, context.DeadlineExceeded) {
		t.Errorf("Tree.All() with timeout: got error %v, want DeadlineExceeded", gotErr)
	}
}
