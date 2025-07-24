package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Digital-Shane/treeview"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Start in current directory
	initialPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Create file browser model
	model := NewFileBrowserModel(initialPath)

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// FileBrowserModel represents the complete file browser state
type FileBrowserModel struct {
	currentPath  string
	treeModel    *treeview.TuiTreeModel[treeview.FileInfo]
	showMetadata bool
	width        int
	height       int
	provider     *treeview.DefaultNodeProvider[treeview.FileInfo]
}

// NewFileBrowserModel creates a new file browser model
func NewFileBrowserModel(initialPath string) *FileBrowserModel {
	provider := treeview.NewFileNodeProvider[treeview.FileInfo]()

	m := &FileBrowserModel{
		currentPath:  initialPath,
		showMetadata: true,
		width:        120,
		height:       30,
		provider:     provider,
	}

	tree, err := loadDirectory(initialPath, provider)
	if err != nil {
		log.Printf("Failed to load directory %s: %v", initialPath, err)
		tree = treeview.NewTree([]*treeview.Node[treeview.FileInfo]{}, treeview.WithProvider[treeview.FileInfo](provider))
	}

	m.treeModel = m.newTuiTreeModel(tree)
	return m
}

// newTuiTreeModel creates a new TUI tree model with the current dimensions.
func (m *FileBrowserModel) newTuiTreeModel(tree *treeview.Tree[treeview.FileInfo]) *treeview.TuiTreeModel[treeview.FileInfo] {
	metadataWidth := m.width * 4 / 10
	treeWidth := m.width - metadataWidth - 3
	if !m.showMetadata {
		treeWidth = m.width
	}

	// Create custom key map to avoid Enter key conflict
	keyMap := treeview.DefaultKeyMap()
	keyMap.SearchStart = []string{"/"}

	return treeview.NewTuiTreeModel(
		tree,
		treeview.WithTuiWidth[treeview.FileInfo](treeWidth),
		treeview.WithTuiHeight[treeview.FileInfo](m.height-2),
		treeview.WithTuiKeyMap[treeview.FileInfo](keyMap),
		treeview.WithTuiDisableNavBar[treeview.FileInfo](true),
	)
}

// loadDirectory loads a directory tree from the filesystem
func loadDirectory(path string, provider treeview.NodeProvider[treeview.FileInfo]) (*treeview.Tree[treeview.FileInfo], error) {
	// Create context with timeout for large directories
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tree, err := treeview.NewTreeFromFileSystem(
		ctx,
		path, false,
		treeview.WithMaxDepth[treeview.FileInfo](4),
		treeview.WithExpandAll[treeview.FileInfo](),
		treeview.WithFilterFunc[treeview.FileInfo](func(item treeview.FileInfo) bool {
			return !strings.HasPrefix(item.Name(), ".")
		}),
		treeview.WithProvider[treeview.FileInfo](provider),
	)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// refreshCurrentDirectory reloads the current directory
func (m *FileBrowserModel) refreshCurrentDirectory() tea.Cmd {
	return func() tea.Msg {
		tree, err := loadDirectory(m.currentPath, m.provider)
		if err != nil {
			return errorMsg{err}
		}
		return directoryLoadedMsg{tree}
	}
}

// navigateToDirectory changes to a new directory
func (m *FileBrowserModel) navigateToDirectory(path string) tea.Cmd {
	return func() tea.Msg {
		// Resolve the path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return errorMsg{err}
		}

		// Check if it's a directory
		info, err := os.Stat(absPath)
		if err != nil {
			return errorMsg{err}
		}
		if !info.IsDir() {
			return errorMsg{fmt.Errorf("not a directory: %s", absPath)}
		}

		tree, err := loadDirectory(absPath, m.provider)
		if err != nil {
			return errorMsg{err}
		}

		return directoryChangedMsg{
			path: absPath,
			tree: tree,
		}
	}
}

// getCurrentSelectedNode returns the currently selected file system node
func (m *FileBrowserModel) getCurrentSelectedNode() *treeview.Node[treeview.FileInfo] {
	if m.treeModel == nil {
		return nil
	}
	return m.treeModel.GetFocusedNode()
}

// Message types for the file browser
type directoryLoadedMsg struct {
	tree *treeview.Tree[treeview.FileInfo]
}

type directoryChangedMsg struct {
	path string
	tree *treeview.Tree[treeview.FileInfo]
}

type errorMsg struct {
	err error
}

// Init initializes the file browser model
func (m *FileBrowserModel) Init() tea.Cmd {
	return m.refreshCurrentDirectory()
}

// Update handles input events and state changes
func (m *FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.treeModel != nil {
			m.treeModel = m.newTuiTreeModel(m.treeModel.Tree)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedNode := m.getCurrentSelectedNode()
			if selectedNode != nil && selectedNode.Data().IsDir() {
				return m, m.navigateToDirectory(selectedNode.Data().Path)
			}
		case "h":
			parentPath := filepath.Dir(m.currentPath)
			if parentPath != m.currentPath {
				return m, m.navigateToDirectory(parentPath)
			}
		case "r":
			return m, m.refreshCurrentDirectory()
		}

	case directoryLoadedMsg:
		m.treeModel = m.newTuiTreeModel(msg.tree)

	case directoryChangedMsg:
		m.currentPath = msg.path
		m.treeModel = m.newTuiTreeModel(msg.tree)

	case errorMsg:
		log.Printf("Error: %v", msg.err)
	}

	// Delegate all other messages to the tree model
	if m.treeModel != nil {
		updatedModel, treeCmd := m.treeModel.Update(msg)
		if newTreeModel, ok := updatedModel.(*treeview.TuiTreeModel[treeview.FileInfo]); ok {
			m.treeModel = newTreeModel
		}
		cmd = treeCmd
	}

	return m, cmd
}

// View renders the complete file browser interface
func (m *FileBrowserModel) View() string {
	if m.treeModel == nil {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteByte('\n')

	if m.showMetadata {
		b.WriteString(m.renderTwoPanelLayout())
	} else {
		b.WriteString(m.treeModel.View())
	}

	b.WriteByte('\n')
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderHeader creates the header bar
func (m *FileBrowserModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("15")).
		Width(m.width).
		Align(lipgloss.Center)

	title := fmt.Sprintf("📂 File Browser - %s", m.currentPath)
	return style.Render(title)
}

// renderTwoPanelLayout creates the two-panel layout with metadata and tree
func (m *FileBrowserModel) renderTwoPanelLayout() string {
	metadataWidth := m.width * 4 / 10
	metadataView := m.renderMetadataPanel(metadataWidth)
	treeView := m.treeModel.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, metadataView, " │ ", treeView)
}

// renderMetadataPanel creates the metadata side panel
func (m *FileBrowserModel) renderMetadataPanel(width int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1)

	selectedNode := m.getCurrentSelectedNode()
	if selectedNode == nil {
		return style.Render("No file selected")
	}

	var info []string
	data := selectedNode.Data()

	if data.IsDir() {
		info = append(info, "Directory", "")
	} else {
		info = append(info, "File", "")
	}

	info = append(info, fmt.Sprintf("Name: %s", data.Name()))

	if data.IsDir() {
		info = append(info, fmt.Sprintf("Expanded: %v", selectedNode.IsExpanded()))
		info = append(info, fmt.Sprintf("Items: %d", len(selectedNode.Children())))
	} else {
		info = append(info, fmt.Sprintf("Size: %s", formatSize(data.Size())))
	}

	modTime := data.ModTime()
	info = append(info, fmt.Sprintf("Modified: %s", modTime.Format("2006-01-02 15:04:05")))
	info = append(info, fmt.Sprintf("Ago: %s", formatTimeSince(modTime)), "")

	if osInfo, err := os.Stat(data.Path); err == nil {
		info = append(info, fmt.Sprintf("Permissions: %s", osInfo.Mode().String()))
		if !data.IsDir() {
			if ext := filepath.Ext(data.Name()); ext != "" {
				info = append(info, fmt.Sprintf("Extension: %s", ext))
			}
		}
	}

	return style.Render(strings.Join(info, "\n"))
}

// renderStatusBar creates the status bar merging tree navigation with file browser controls
func (m *FileBrowserModel) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15")).
		Width(m.width).
		Padding(0, 1)

	// Get the tree navigation controls from the tree model
	treeNav := m.treeModel.NavBar()

	// Add file browser specific controls
	fileBrowserNav := "enter: Focus  h: Parent  r: Refresh"

	// Combine both sets of controls
	var statusText string
	if treeNav != "" {
		statusText = treeNav + "  " + fileBrowserNav
	} else {
		statusText = fileBrowserNav
	}

	return style.Render(statusText)
}

// Utility functions

// formatSize formats file size in human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatTimeSince formats time duration since a given time
func formatTimeSince(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d min ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		return fmt.Sprintf("%d years ago", years)
	}
}
