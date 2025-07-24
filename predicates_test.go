package treeview

import (
	"regexp"
	"testing"
)

// Mock types for testing
type mockDirData struct {
	isDir bool
}

func (m mockDirData) IsDir() bool {
	return m.isDir
}

type mockFileData struct{}

// Helper to create test nodes
func createTestNode[T any](name string, data T, expanded bool) *Node[T] {
	n := &Node[T]{
		name:     name,
		data:     data,
		expanded: expanded,
	}
	return n
}

func TestPredHasExtension(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		nodeName   string
		want       bool
	}{
		{"single_match", []string{".txt"}, "file.txt", true},
		{"single_no_match", []string{".txt"}, "file.md", false},
		{"multiple_match_first", []string{".txt", ".md"}, "file.txt", true},
		{"multiple_match_second", []string{".txt", ".md"}, "file.md", true},
		{"multiple_no_match", []string{".txt", ".md"}, "file.go", false},
		{"case_insensitive", []string{".txt"}, "FILE.TXT", true},
		{"no_extension", []string{".txt"}, "file", false},
		{"empty_extensions", []string{}, "file.txt", false},
		{"dot_file", []string{""}, ".gitignore", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredHasExtension[struct{}](test.extensions...)
			got := pred(node)
			if got != test.want {
				t.Errorf("PredHasExtension(%v)(&Node{name:%q}) = %v, want %v",
					test.extensions, test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredIsDir(t *testing.T) {
	tests := []struct {
		name     string
		nodeData any
		want     bool
	}{
		{"dir_true", mockDirData{isDir: true}, true},
		{"dir_false", mockDirData{isDir: false}, false},
		{"non_dir_type", mockFileData{}, false},
		{"string_data", "some string", false},
		{"nil_data", nil, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode("test", test.nodeData, false)
			pred := PredIsDir[any]()
			got := pred(node)
			if got != test.want {
				t.Errorf("PredIsDir()(&Node{data:%T}) = %v, want %v",
					test.nodeData, got, test.want)
			}
		})
	}
}

func TestPredIsFile(t *testing.T) {
	tests := []struct {
		name     string
		nodeData any
		want     bool
	}{
		{"dir_true", mockDirData{isDir: true}, false},
		{"dir_false", mockDirData{isDir: false}, true},
		{"non_dir_type", mockFileData{}, true},
		{"string_data", "some string", true},
		{"nil_data", nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode("test", test.nodeData, false)
			pred := PredIsFile[any]()
			got := pred(node)
			if got != test.want {
				t.Errorf("PredIsFile()(&Node{data:%T}) = %v, want %v",
					test.nodeData, got, test.want)
			}
		})
	}
}

func TestPredIsHidden(t *testing.T) {
	tests := []struct {
		name     string
		nodeName string
		want     bool
	}{
		{"hidden_file", ".gitignore", true},
		{"hidden_dir", ".config", true},
		{"visible_file", "file.txt", false},
		{"visible_dir", "src", false},
		{"empty_name", "", false},
		{"just_dot", ".", true},
		{"double_dot", "..", true},
		{"dot_in_middle", "file.name", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredIsHidden[struct{}]()
			got := pred(node)
			if got != test.want {
				t.Errorf("PredIsHidden()(&Node{name:%q}) = %v, want %v",
					test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredIsExpanded(t *testing.T) {
	tests := []struct {
		name     string
		expanded bool
		want     bool
	}{
		{"expanded", true, true},
		{"collapsed", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode("test", struct{}{}, test.expanded)
			pred := PredIsExpanded[struct{}]()
			got := pred(node)
			if got != test.want {
				t.Errorf("PredIsExpanded()(&Node{expanded:%v}) = %v, want %v",
					test.expanded, got, test.want)
			}
		})
	}
}

func TestPredIsCollapsed(t *testing.T) {
	tests := []struct {
		name     string
		expanded bool
		want     bool
	}{
		{"expanded", true, false},
		{"collapsed", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode("test", struct{}{}, test.expanded)
			pred := PredIsCollapsed[struct{}]()
			got := pred(node)
			if got != test.want {
				t.Errorf("PredIsCollapsed()(&Node{expanded:%v}) = %v, want %v",
					test.expanded, got, test.want)
			}
		})
	}
}

func TestPredHasName(t *testing.T) {
	tests := []struct {
		name       string
		targetName string
		nodeName   string
		want       bool
	}{
		{"exact_match", "file.txt", "file.txt", true},
		{"case_sensitive", "file.txt", "File.txt", false},
		{"no_match", "file.txt", "other.txt", false},
		{"empty_target", "", "", true},
		{"empty_node", "file.txt", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredHasName[struct{}](test.targetName)
			got := pred(node)
			if got != test.want {
				t.Errorf("PredHasName(%q)(&Node{name:%q}) = %v, want %v",
					test.targetName, test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredHasNameIgnoreCase(t *testing.T) {
	tests := []struct {
		name       string
		targetName string
		nodeName   string
		want       bool
	}{
		{"exact_match", "file.txt", "file.txt", true},
		{"case_insensitive", "file.txt", "File.TXT", true},
		{"mixed_case", "FiLe.TxT", "file.txt", true},
		{"no_match", "file.txt", "other.txt", false},
		{"empty_both", "", "", true},
		{"unicode", "café", "CAFÉ", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredHasNameIgnoreCase[struct{}](test.targetName)
			got := pred(node)
			if got != test.want {
				t.Errorf("PredHasNameIgnoreCase(%q)(&Node{name:%q}) = %v, want %v",
					test.targetName, test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredContainsText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		nodeName string
		want     bool
	}{
		{"contains", "test", "test_file.txt", true},
		{"case_insensitive", "TEST", "test_file.txt", true},
		{"partial_match", "file", "test_file.txt", true},
		{"no_match", "xyz", "test_file.txt", false},
		{"empty_text", "", "test_file.txt", true},
		{"empty_node", "test", "", false},
		{"both_empty", "", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredContainsText[struct{}](test.text)
			got := pred(node)
			if got != test.want {
				t.Errorf("PredContainsText(%q)(&Node{name:%q}) = %v, want %v",
					test.text, test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredMatchesRegex(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		want     bool
	}{
		{"simple_match", "^test", "test_file.txt", true},
		{"no_match", "^test", "file_test.txt", false},
		{"extension_match", `\.go$`, "main.go", true},
		{"extension_no_match", `\.go$`, "main.js", false},
		{"complex_pattern", `^[a-z]+_\d+\.txt$`, "file_123.txt", true},
		{"empty_pattern", "", "anything", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pattern := regexp.MustCompile(test.pattern)
			node := createTestNode(test.nodeName, struct{}{}, false)
			pred := PredMatchesRegex[struct{}](pattern)
			got := pred(node)
			if got != test.want {
				t.Errorf("PredMatchesRegex(%q)(&Node{name:%q}) = %v, want %v",
					test.pattern, test.nodeName, got, test.want)
			}
		})
	}
}

func TestPredAny(t *testing.T) {
	hiddenNode := createTestNode(".hidden", struct{}{}, false)
	goFileNode := createTestNode("main.go", struct{}{}, false)
	txtFileNode := createTestNode("file.txt", struct{}{}, false)

	tests := []struct {
		name       string
		predicates []func(*Node[struct{}]) bool
		node       *Node[struct{}]
		want       bool
	}{
		{
			"all_true",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".hidden"),
			},
			hiddenNode,
			true,
		},
		{
			"one_true",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			},
			goFileNode,
			true,
		},
		{
			"none_true",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			},
			txtFileNode,
			false,
		},
		{
			"empty_predicates",
			[]func(*Node[struct{}]) bool{},
			txtFileNode,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pred := PredAny(test.predicates...)
			got := pred(test.node)
			if got != test.want {
				t.Errorf("PredAny(%d predicates)(&Node{name:%q}) = %v, want %v",
					len(test.predicates), test.node.Name(), got, test.want)
			}
		})
	}
}

func TestPredAll(t *testing.T) {
	hiddenGoNode := createTestNode(".config.go", struct{}{}, false)
	goFileNode := createTestNode("main.go", struct{}{}, false)
	hiddenTxtNode := createTestNode(".hidden.txt", struct{}{}, false)

	tests := []struct {
		name       string
		predicates []func(*Node[struct{}]) bool
		node       *Node[struct{}]
		want       bool
	}{
		{
			"all_true",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			},
			hiddenGoNode,
			true,
		},
		{
			"one_false",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			},
			goFileNode,
			false,
		},
		{
			"all_false",
			[]func(*Node[struct{}]) bool{
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			},
			hiddenTxtNode,
			false,
		},
		{
			"empty_predicates",
			[]func(*Node[struct{}]) bool{},
			goFileNode,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pred := PredAll(test.predicates...)
			got := pred(test.node)
			if got != test.want {
				t.Errorf("PredAll(%d predicates)(&Node{name:%q}) = %v, want %v",
					len(test.predicates), test.node.Name(), got, test.want)
			}
		})
	}
}

func TestPredNot(t *testing.T) {
	hiddenNode := createTestNode(".hidden", struct{}{}, false)
	visibleNode := createTestNode("visible", struct{}{}, false)

	tests := []struct {
		name      string
		predicate func(*Node[struct{}]) bool
		node      *Node[struct{}]
		want      bool
	}{
		{
			"not_hidden_on_hidden",
			PredIsHidden[struct{}](),
			hiddenNode,
			false,
		},
		{
			"not_hidden_on_visible",
			PredIsHidden[struct{}](),
			visibleNode,
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pred := PredNot(test.predicate)
			got := pred(test.node)
			if got != test.want {
				t.Errorf("PredNot(predicate)(&Node{name:%q}) = %v, want %v",
					test.node.Name(), got, test.want)
			}
		})
	}
}

// Test edge cases and combinations
func TestPredicateCombinations(t *testing.T) {
	t.Run("complex_combination", func(t *testing.T) {
		// Create a predicate that matches hidden Go files or visible test files
		pred := PredAny(
			PredAll(
				PredIsHidden[struct{}](),
				PredHasExtension[struct{}](".go"),
			),
			PredAll(
				PredNot(PredIsHidden[struct{}]()),
				PredContainsText[struct{}]("test"),
			),
		)

		tests := []struct {
			nodeName string
			want     bool
		}{
			{".hidden.go", true},       // hidden go file
			{"visible_test.txt", true}, // visible test file
			{"visible.go", false},      // visible go file (no "test")
			{".hidden.txt", false},     // hidden non-go file
		}

		for _, test := range tests {
			node := createTestNode(test.nodeName, struct{}{}, false)
			got := pred(node)
			if got != test.want {
				t.Errorf("complex predicate(&Node{name:%q}) = %v, want %v",
					test.nodeName, got, test.want)
			}
		}
	})
}

// Test with different generic types
func TestPredicatesWithDifferentTypes(t *testing.T) {
	t.Run("string_type", func(t *testing.T) {
		node := createTestNode("test.go", "string data", false)
		pred := PredHasExtension[string](".go")
		got := pred(node)
		if !got {
			t.Errorf("PredHasExtension[string](\".go\")(&Node{name:\"test.go\"}) = false, want true")
		}
	})

	t.Run("int_type", func(t *testing.T) {
		node := createTestNode(".hidden", 42, false)
		pred := PredIsHidden[int]()
		got := pred(node)
		if !got {
			t.Errorf("PredIsHidden[int]()(&Node{name:\".hidden\"}) = false, want true")
		}
	})
}

// Verify Node implementation matches test expectations
func TestNodeImplementation(t *testing.T) {
	t.Run("node_name", func(t *testing.T) {
		node := createTestNode("test.txt", struct{}{}, false)
		got := node.Name()
		want := "test.txt"
		if got != want {
			t.Errorf("Node.Name() = %q, want %q", got, want)
		}
	})

	t.Run("node_expanded", func(t *testing.T) {
		node := createTestNode("test", struct{}{}, true)
		got := node.IsExpanded()
		want := true
		if got != want {
			t.Errorf("Node.IsExpanded() = %v, want %v", got, want)
		}
	})

	t.Run("node_data", func(t *testing.T) {
		data := mockDirData{isDir: true}
		node := createTestNode("test", data, false)
		got := node.Data()
		// Compare the actual data
		if got != data {
			t.Errorf("Node.Data() = %v, want %v", got, data)
		}
	})
}
