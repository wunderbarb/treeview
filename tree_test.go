package treeview

import (
	"context"
	"errors"
	"testing"
)

func TestTree_GetFocusedID_NoFocus(t *testing.T) {
	tree := NewTree([]*Node[string]{})

	got := tree.GetFocusedID()
	want := ""

	if got != want {
		t.Errorf("GetFocusedID() = %q, want %q", got, want)
	}
}

func TestTree_SetFocusedID_EdgeCases(t *testing.T) {
	root := NewNode("root", "root", "root")
	child := NewNode("child", "child", "child")
	root.AddChild(child)
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	tests := []struct {
		name      string
		id        string
		wantErr   bool
		errIs     error
		wantFocus bool
	}{
		{
			name:      "clear_focus_empty_id",
			id:        "",
			wantErr:   false,
			wantFocus: false,
		},
		{
			name:    "nonexistent_id",
			id:      "nonexistent",
			wantErr: true,
			errIs:   ErrNodeNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed, err := tree.SetFocusedID(ctx, test.id)

			if test.wantErr {
				if err == nil {
					t.Errorf("SetFocusedID(%q) = %v, %v, want error", test.id, changed, err)
					return
				}
				if test.errIs != nil && !errors.Is(err, test.errIs) {
					t.Errorf("SetFocusedID(%q) error = %v, want %v", test.id, err, test.errIs)
				}
				return
			}

			if err != nil {
				t.Errorf("SetFocusedID(%q) = %v, %v, want no error", test.id, changed, err)
				return
			}

			gotFocus := tree.GetFocusedID()
			wantFocusID := test.id
			if test.wantFocus && gotFocus != wantFocusID {
				t.Errorf("SetFocusedID(%q) focused = %q, want %q", test.id, gotFocus, wantFocusID)
			}
		})
	}
}

func TestTree_SetFocusedID_AlreadyFocused(t *testing.T) {
	root := NewNode("root", "root", "root")
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	// Root should already be focused by NewTree, so try to focus on same node again
	changed, err := tree.SetFocusedID(ctx, "root")
	if err != nil {
		t.Errorf("SetFocusedID(root) already focused = %v, %v, want no error", changed, err)
	}
	if changed {
		t.Errorf("SetFocusedID(root) already focused = %v, want false", changed)
	}
}

func TestTree_FindByID_ContextCancelled(t *testing.T) {
	root := NewNode("root", "root", "root")
	tree := NewTree([]*Node[string]{root})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := tree.FindByID(ctx, "root")
	if err == nil {
		t.Errorf("FindByID(cancelled_context, root) = _, nil, want context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("FindByID(cancelled_context, root) error = %v, want context.Canceled", err)
	}
}

func TestTree_Move_EdgeCases(t *testing.T) {
	// Create a simple tree structure
	root := NewNode("root", "root", "root")
	child1 := NewNode("child1", "child1", "child1")
	child2 := NewNode("child2", "child2", "child2")
	root.AddChild(child1)
	root.AddChild(child2)
	root.Expand()

	tree := NewTree([]*Node[string]{root}, WithFocusPolicy[string](wrapPolicy))
	ctx := context.Background()

	tests := []struct {
		name       string
		setupFocus string
		offset     int
		wantErr    bool
		wantChange bool
	}{
		{
			name:       "no_focus_move_down",
			setupFocus: "",
			offset:     1,
			wantErr:    false,
			wantChange: true,
		},
		{
			name:       "move_zero_offset",
			setupFocus: "root",
			offset:     0,
			wantErr:    false,
			wantChange: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup focus
			if test.setupFocus != "" {
				_, err := tree.SetFocusedID(ctx, test.setupFocus)
				if err != nil {
					t.Fatalf("Setup focus failed: %v", err)
				}
			} else {
				_, _ = tree.SetFocusedID(ctx, "") // Clear focus
			}

			changed, err := tree.Move(ctx, test.offset)

			if test.wantErr && err == nil {
				t.Errorf("Move(%d) = %v, nil, want error", test.offset, changed)
			} else if !test.wantErr && err != nil {
				t.Errorf("Move(%d) = %v, %v, want no error", test.offset, changed, err)
			}

			if changed != test.wantChange {
				t.Errorf("Move(%d) changed = %v, want %v", test.offset, changed, test.wantChange)
			}
		})
	}
}

// wrapPolicy is a simple focus policy that wraps around
func wrapPolicy[T any](ctx context.Context, visible []*Node[T], focused *Node[T], offset int) (*Node[T], error) {
	if len(visible) == 0 {
		return nil, nil
	}
	if offset == 0 {
		return focused, nil
	}
	if focused == nil {
		return visible[0], nil
	}

	// Find current index
	currentIdx := -1
	for i, node := range visible {
		if node == focused {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		return visible[0], nil
	}

	// Calculate new index with wrapping
	newIdx := (currentIdx + offset) % len(visible)
	if newIdx < 0 {
		newIdx += len(visible)
	}

	return visible[newIdx], nil
}

func TestTree_SetExpanded_NotFound(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})
	ctx := context.Background()

	changed, err := tree.SetExpanded(ctx, "nonexistent", true)

	if err == nil {
		t.Errorf("SetExpanded(nonexistent, true) = %v, nil, want error", changed)
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("SetExpanded(nonexistent, true) error = %v, want ErrNodeNotFound", err)
	}
	if changed {
		t.Errorf("SetExpanded(nonexistent, true) changed = %v, want false", changed)
	}
}

func TestTree_ToggleFocused(t *testing.T) {
	root := NewNode("root", "root", "root")
	child := NewNode("child", "child", "child")
	root.AddChild(child)
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	tests := []struct {
		name         string
		setupFocus   string
		wantExpanded bool
	}{
		{
			name:         "no_focus",
			setupFocus:   "",
			wantExpanded: false, // Root starts collapsed
		},
		{
			name:         "toggle_focused_node",
			setupFocus:   "root",
			wantExpanded: true, // Should expand from collapsed
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset root to collapsed state
			root.Collapse()

			// Setup focus
			if test.setupFocus != "" {
				_, err := tree.SetFocusedID(ctx, test.setupFocus)
				if err != nil {
					t.Fatalf("Setup focus failed: %v", err)
				}
			} else {
				_, _ = tree.SetFocusedID(ctx, "")
			}

			tree.ToggleFocused(ctx)

			if test.setupFocus != "" {
				got := root.IsExpanded()
				if got != test.wantExpanded {
					t.Errorf("ToggleFocused() root expanded = %v, want %v", got, test.wantExpanded)
				}
			}
		})
	}
}

func TestTree_Search_EmptyTerm(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})
	ctx := context.Background()

	results, err := tree.Search(ctx, "")

	if err != nil {
		t.Errorf("Search(\"\") = %v, %v, want nil error", results, err)
	}
	if results != nil {
		t.Errorf("Search(\"\") = %v, want nil", results)
	}
}

func TestTree_Search_ContextError(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := tree.Search(ctx, "term")

	if err == nil {
		t.Errorf("Search(cancelled_context, term) = %v, nil, want context error", results)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Search(cancelled_context, term) error = %v, want context.Canceled", err)
	}
}

func TestTree_SearchAndExpand_EmptyTerm(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})
	ctx := context.Background()

	results, err := tree.SearchAndExpand(ctx, "")

	if err != nil {
		t.Errorf("SearchAndExpand(\"\") = %v, %v, want nil error", results, err)
	}
	if results != nil {
		t.Errorf("SearchAndExpand(\"\") = %v, want nil", results)
	}
}

func TestTree_SearchAndExpand_NoMatches(t *testing.T) {
	root := NewNode("root", "root", "root")
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	results, err := tree.SearchAndExpand(ctx, "nonexistent")

	if err != nil {
		t.Errorf("SearchAndExpand(nonexistent) = %v, %v, want nil error", results, err)
	}
	if len(results) != 0 {
		t.Errorf("SearchAndExpand(nonexistent) = %v, want empty slice", results)
	}

	// Verify that ExpandAll and ShowAll were called
	if !root.IsExpanded() {
		t.Errorf("SearchAndExpand(nonexistent) root expanded = %v, want true", root.IsExpanded())
	}
	if !root.IsVisible() {
		t.Errorf("SearchAndExpand(nonexistent) root visible = %v, want true", root.IsVisible())
	}
}

func TestTree_SearchAndExpand_ContextErrors(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := tree.SearchAndExpand(ctx, "term")

	if err == nil {
		t.Errorf("SearchAndExpand(cancelled_context, term) = %v, nil, want context error", results)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("SearchAndExpand(cancelled_context, term) error = %v, want context.Canceled", err)
	}
}

func TestTree_Render(t *testing.T) {
	root := NewNode("root", "root", "root")
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	output, err := tree.Render(ctx)

	if err != nil {
		t.Errorf("Render() = %q, %v, want no error", output, err)
	}
	if output == "" {
		t.Errorf("Render() = %q, want non-empty output", output)
	}
}

func TestTree_Render_ContextCancelled(t *testing.T) {
	root := NewNode("root", "root", "root")
	tree := NewTree([]*Node[string]{root})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	output, err := tree.Render(ctx)

	if err == nil {
		t.Errorf("Render(cancelled_context) = %q, nil, want context error", output)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Render(cancelled_context) error = %v, want context.Canceled", err)
	}
}

func TestTree_VisibleNodes_ContextError(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	nodes, err := tree.visibleNodes(ctx)

	if err == nil {
		t.Errorf("visibleNodes(cancelled_context) = %v, nil, want context error", nodes)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("visibleNodes(cancelled_context) error = %v, want context.Canceled", err)
	}
}

func TestTree_SetExpandedState_ContextError(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := tree.setExpandedState(ctx, true)

	if err == nil {
		t.Errorf("setExpandedState(cancelled_context, true) = nil, want context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("setExpandedState(cancelled_context, true) error = %v, want context.Canceled", err)
	}
}

func TestTree_SetVisibleState_ContextError(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := tree.setVisibleState(ctx, true)

	if err == nil {
		t.Errorf("setVisibleState(cancelled_context, true) = nil, want context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("setVisibleState(cancelled_context, true) error = %v, want context.Canceled", err)
	}
}

// Test concurrent access to demonstrate thread safety
func TestTree_ConcurrentAccess(t *testing.T) {
	root := NewNode("root", "root", "root")
	for i := 0; i < 10; i++ {
		child := NewNode("child", "child", "child")
		root.AddChild(child)
	}
	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	// Run multiple goroutines accessing tree methods
	done := make(chan bool, 4)

	// Goroutine 1: Read operations
	go func() {
		for i := 0; i < 100; i++ {
			_ = tree.Nodes()
			_ = tree.GetFocusedID()
			_ = tree.GetFocusedNode()
		}
		done <- true
	}()

	// Goroutine 2: Focus operations
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = tree.SetFocusedID(ctx, "root")
			_, _ = tree.SetFocusedID(ctx, "")
		}
		done <- true
	}()

	// Goroutine 3: Expansion operations
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = tree.SetExpanded(ctx, "root", true)
			_, _ = tree.SetExpanded(ctx, "root", false)
		}
		done <- true
	}()

	// Goroutine 4: Toggle operations
	go func() {
		for i := 0; i < 100; i++ {
			tree.ToggleFocused(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// If we get here without data races, the test passes
}

func TestTree_Move_VisibleNodesError(t *testing.T) {
	tree := NewTree([]*Node[string]{NewNode("root", "root", "root")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	changed, err := tree.Move(ctx, 1)

	if err == nil {
		t.Errorf("Move(cancelled_context, 1) = %v, nil, want context error", changed)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Move(cancelled_context, 1) error = %v, want context.Canceled", err)
	}
}

func TestTree_Move_FocusPolicyError(t *testing.T) {
	root := NewNode("root", "root", "root")

	// Create a focus policy that returns an error
	errorPolicy := func(ctx context.Context, visible []*Node[string], focused *Node[string], offset int) (*Node[string], error) {
		return nil, errors.New("focus policy error")
	}

	tree := NewTree([]*Node[string]{root}, WithFocusPolicy[string](errorPolicy))
	ctx := context.Background()

	changed, err := tree.Move(ctx, 1)

	if err == nil {
		t.Errorf("Move(error_policy, 1) = %v, nil, want focus policy error", changed)
	}
	if err.Error() != "focus policy error" {
		t.Errorf("Move(error_policy, 1) error = %q, want %q", err.Error(), "focus policy error")
	}
}

func TestTree_SearchAndExpand_WithMatches(t *testing.T) {
	root := NewNode("root", "root", "root")
	child1 := NewNode("child1", "match", "match")
	child2 := NewNode("child2", "child2", "child2")
	root.AddChild(child1)
	root.AddChild(child2)

	tree := NewTree([]*Node[string]{root})
	ctx := context.Background()

	matches, err := tree.SearchAndExpand(ctx, "match")

	if err != nil {
		t.Errorf("SearchAndExpand(match) = %v, %v, want no error", matches, err)
	}
	if len(matches) != 1 {
		t.Errorf("SearchAndExpand(match) matches = %v, want 1 match", len(matches))
	}
	if len(matches) > 0 && matches[0] != child1 {
		t.Errorf("SearchAndExpand(match) first match = %v, want child1", matches[0])
	}

	// Verify focus is on first match
	focusedID := tree.GetFocusedID()
	if focusedID != "child1" {
		t.Errorf("SearchAndExpand(match) focused = %q, want %q", focusedID, "child1")
	}

	// Verify ancestors are expanded and visible
	if !root.IsExpanded() {
		t.Errorf("SearchAndExpand(match) root expanded = %v, want true", root.IsExpanded())
	}
	if !child1.IsVisible() {
		t.Errorf("SearchAndExpand(match) child1 visible = %v, want true", child1.IsVisible())
	}
}
