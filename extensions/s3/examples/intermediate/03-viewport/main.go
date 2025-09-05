package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/extensions/s3"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

func main() {
	initEnv()
	defer func() { _ = localstack.UseNot() }()
	provider := createViewportProvider()
	tree, _ := s3.NewTreeFromS3(context.Background(), &s3.InputTreeFromS3{Path: _cs3Testdata},
		treeview.WithProvider[treeview.FileInfo](provider))
	model := treeview.NewTuiTreeModel(
		tree,
		treeview.WithTuiWidth[treeview.FileInfo](80),
		treeview.WithTuiHeight[treeview.FileInfo](20),
	)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// viewportFormatter formats files with size information
func viewportFormatter(node *treeview.Node[treeview.FileInfo]) (string, bool) {
	data := *node.Data()
	if data.Size() > 0 {
		return fmt.Sprintf("%s (%s)", data.Name(), formatFileSize(data.Size())), true
	}
	return data.Name(), true
}

// createViewportProvider creates a DefaultNodeProvider with file size formatting
func createViewportProvider() *treeview.DefaultNodeProvider[treeview.FileInfo] {
	// Start with default file node provider options
	baseOptions := []treeview.ProviderOption[treeview.FileInfo]{
		treeview.WithDefaultFolderRules[treeview.FileInfo](),
		treeview.WithDefaultFileRules[treeview.FileInfo](),
		treeview.WithFileExtensionRules[treeview.FileInfo](),
		treeview.WithDefaultIcon[treeview.FileInfo]("•"),
		treeview.WithFormatter[treeview.FileInfo](viewportFormatter),
	}

	return treeview.NewDefaultNodeProvider(baseOptions...)
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * KB
		GB = MB * KB
	)
	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < MB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	}
	if bytes < GB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
}

const (
	_cS3URI       = "s3://"
	_myTestBucket = "wunderbarb.example.com"
	_cS3          = _cS3URI + _myTestBucket
	_cs3Testdata  = _cS3 + "/testdata"
	_c100K        = "sample100K.golden"
	_c1M          = "sample1M.golden"
	_cGolden100K  = _cs3Testdata + "/golden/" + _c100K
	_cGolden1M    = _cs3Testdata + "/golden/" + _c1M
)

// initEnv initializes and populates a localstack bucket.
func initEnv() {
	ok := localstack.IsRunning()
	if !ok {
		panic("localstack not running")
	}
	_ = localstack.Use()

	goldenDirPath := filepath.Join("..", "..", "internal", "testfixtures")
	_ = localstack.CreateBucket(_myTestBucket, localstack.WithNoErrorIfExist())
	_ = localstack.PutObject(_cGolden100K, filepath.Join(goldenDirPath, _c100K))
	_ = localstack.PutObject(_cGolden1M, filepath.Join(goldenDirPath, _c1M))
	_ = localstack.PutObject(_cs3Testdata+"/golden/recurse/"+_c100K, filepath.Join(goldenDirPath, _c100K))
}
