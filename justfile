# ==============================================================================
# Treeview Development & Demo Management Justfile
# ==============================================================================
# This justfile automates development tasks and demo GIF generation for the
# treeview library.
# ==============================================================================

# Show help
default:
    @just --list --unsorted --list-heading $'Development automation and demo management for treeview\n'

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run benchmarks (skip regular tests)
bench:
    go test -run=^$ -bench=. -benchmem ./...

# Generate test coverage report and open in browser
test-coverage:
    #!/usr/bin/env bash
    set -euo pipefail

    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"
    
    # Open in browser (macOS/Linux)
    if command -v open >/dev/null 2>&1; then
        open coverage.html
    elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open coverage.html
    else
        echo "Open coverage.html in your browser to view the report"
    fi

# ==============================================================================
# Demo Management Commands  
# ==============================================================================

# Run a demo using fuzzy search. `just run basic/01` or `just run b/01`
run path:
    #!/usr/bin/env bash
    set -euo pipefail
    
    found_dir=$(just _find_example "{{path}}")
    echo "Running demo: $found_dir"
    cd "$found_dir"
    go run main.go

# Generate and publish demo gif for an example. `just update-gif basic/01` or `just update-gif b/01`
update-gif path:
    #!/usr/bin/env bash
    set -euo pipefail
    
    found_dir=$(just _find_example "{{path}}")
    echo "Processing: $found_dir"
    just _process_example "$found_dir"

# Generate demo gif for an example without publishing. `just create-gif basic/01` or `just create-gif b/01`
create-gif path:
    #!/usr/bin/env bash
    set -euo pipefail
    
    found_dir=$(just _find_example "{{path}}")
    echo "Creating GIF for: $found_dir"
    just _create_gif_only "$found_dir"

# Update demo gifs for all examples
update-all:
    #!/usr/bin/env bash
    set -euo pipefail
    
    echo "Creating demos for all examples..."
    
    # Find all example directories dynamically
    examples=($(find examples -type d -name "*-*" | sort))
    
    if [[ ${#examples[@]} -eq 0 ]]; then
        echo "No examples found"
        exit 1
    fi
    
    echo "Found ${#examples[@]} examples:"
    printf '  • %s\n' "${examples[@]#examples/}"
    echo
    
    # Process each example
    for example_dir in "${examples[@]}"; do
        echo "📁 Processing: $example_dir"
        just _process_example "$example_dir"
        echo
    done
    
    say "All demos created successfully!"
    echo "All demos created successfully!"

# Internal: Find example directory using fuzzy matching
_find_example path:
    #!/usr/bin/env bash
    set -euo pipefail
    
    # Find the example directory using fuzzy matching
    found_dir=""
    path="{{path}}"
    
    # Handle shorthand category prefixes (b/ -> basic/, i/ -> intermediate/)
    if [[ "$path" == "b/"* ]]; then
        path="basic/${path#b/}"
    elif [[ "$path" == "i/"* ]]; then
        path="intermediate/${path#i/}"
    fi
    
    # Try exact path first
    if [[ -d "examples/$path" ]]; then
        found_dir="examples/$path"
    elif [[ "$path" == *"/"* ]]; then
        # Category-aware fuzzy matching (basic/01 -> basic/01-*)
        category=$(echo "$path" | cut -d'/' -f1)
        number=$(echo "$path" | cut -d'/' -f2)
        found_dir=$(find examples -type d -path "examples/$category/$number-*" | head -1)
    else
        # General fuzzy matching
        found_dir=$(find examples -type d -name "*$path*" | head -1)
    fi
    
    if [[ -z "$found_dir" ]]; then
        echo "No example found matching '{{path}}'" >&2
        echo "Available examples:" >&2
        find examples -type d -name "*-*" | sed 's|examples/||' | sort | sed 's/^/  • /' >&2
        exit 1
    fi
    
    echo "$found_dir"

# Internal: Process a specific example directory
_process_example example_dir:
    #!/usr/bin/env bash
    set -euo pipefail
    
    cd "{{example_dir}}"
    
    # Generate GIF if needed
    if [[ ! -f "demo.gif" ]]; then
        just _create_gif_only "{{example_dir}}"
    fi
    
    # Publish to vhs.charm.sh
    echo "Publishing..."
    new_url=$(vhs publish demo.gif -q | tr -d '\n')
    echo "Published: $new_url"
    
    # Update local README
    if [[ -f "README.md" ]]; then
        sed -i '' -E "s|!\[[^]]*\]\(https://vhs\.charm\.sh/[^)]+\.gif\)|![The Demo]($new_url)|g" README.md
    fi
    
    # Update root README
    example_path=$(pwd | sed 's|.*/examples/||')
    cd - > /dev/null
    just _update_root_readme "$example_path" "$new_url"
    cd "examples/$example_path"
    
    # Cleanup
    rm -f demo.gif
    echo "Done!"

# Internal: Create GIF for a specific example directory without publishing
_create_gif_only example_dir:
    #!/usr/bin/env bash
    set -euo pipefail
    
    cd "{{example_dir}}"
    
    # Generate GIF
    echo "Generating demo.gif..."
    vhs demo.tape
    echo "GIF created at: $(pwd)/demo.gif"

# Internal: Update root README.md with new demo URL
_update_root_readme example_path new_url:
    #!/usr/bin/env bash
    set -euo pipefail
    
    [[ ! -f "README.md" ]] && return
    
    # Escape slashes in example_path for sed
    escaped_path=$(echo "{{example_path}}" | sed 's/\//\\\//g')
    
    # Update the demo URL for the matching example path
    sed -i '' -E "/examples\/.*$escaped_path/,/^<img/ s|src=\"[^\"]*\"|src=\"{{new_url}}\"|" README.md

