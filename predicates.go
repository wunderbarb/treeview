package treeview

import (
	"regexp"
	"strings"
)

// Common predicate functions for use with DefaultNodeProvider rules.
// These functions return predicates that can be used with WithIconRule,
// WithStyleRule, and other provider options.

// PredHasExtension returns a predicate that checks if a node's name ends with any of the given extensions.
func PredHasExtension[T any](extensions ...string) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		name := strings.ToLower(n.Name())
		for _, ext := range extensions {
			if strings.HasSuffix(name, ext) {
				return true
			}
		}
		return false
	}
}

// PredIsDir returns a predicate that checks if a node represents a directory.
// This works by checking if the node's data implements the IsDir() bool method.
func PredIsDir[T any]() func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		if dirNode, ok := any(n.Data()).(interface{ IsDir() bool }); ok {
			return dirNode.IsDir()
		}
		return false
	}
}

// PredIsFile returns a predicate that checks if a node represents a file.
// This is the inverse of PredIsDir.
func PredIsFile[T any]() func(*Node[T]) bool {
	return PredNot(PredIsDir[T]())
}

// PredIsHidden returns a predicate that checks if a node's name starts with a dot,
// indicating it's a hidden file or directory.
func PredIsHidden[T any]() func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return strings.HasPrefix(n.Name(), ".")
	}
}

// PredIsExpanded returns a predicate that checks if a node is currently expanded.
func PredIsExpanded[T any]() func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return n.IsExpanded()
	}
}

// PredIsCollapsed returns a predicate that checks if a node is currently collapsed.
func PredIsCollapsed[T any]() func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return !n.IsExpanded()
	}
}

// PredHasName returns a predicate that checks if a node's name equals the given name.
// The comparison is case-sensitive.
func PredHasName[T any](name string) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return n.Name() == name
	}
}

// PredHasNameIgnoreCase returns a predicate that checks if a node's name equals the given name.
// The comparison is case-insensitive.
func PredHasNameIgnoreCase[T any](name string) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return strings.EqualFold(n.Name(), name)
	}
}

// PredContainsText returns a predicate that checks if a node's name contains the given text.
// The comparison is case-insensitive.
func PredContainsText[T any](text string) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return strings.Contains(strings.ToLower(n.Name()), strings.ToLower(text))
	}
}

// PredMatchesRegex returns a predicate that checks if a node's name matches the compiled regex pattern.
// Pass a compiled regexp.Regexp for efficient repeated matching.
func PredMatchesRegex[T any](pattern *regexp.Regexp) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return pattern.MatchString(n.Name())
	}
}

// PredAny combines multiple predicates with logical OR.
// At least one predicate must return true for the combined predicate to return true.
func PredAny[T any](predicates ...func(*Node[T]) bool) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		for _, pred := range predicates {
			if pred(n) {
				return true
			}
		}
		return false
	}
}

// PredAll combines multiple predicates with logical AND.
// All predicates must return true for the combined predicate to return true.
func PredAll[T any](predicates ...func(*Node[T]) bool) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		for _, pred := range predicates {
			if !pred(n) {
				return false
			}
		}
		return true
	}
}

// PredNot negates a predicate.
func PredNot[T any](predicate func(*Node[T]) bool) func(*Node[T]) bool {
	return func(n *Node[T]) bool {
		return !predicate(n)
	}
}
