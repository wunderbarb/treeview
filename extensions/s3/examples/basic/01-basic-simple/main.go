package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/examples/shared"
	"github.com/Digital-Shane/treeview/extensions/s3"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

// //////////////////////////////////////////////////////////////////
//   Simple Tree Example
// //////////////////////////////////////////////////////////////////

func main() {
	initEnv()
	defer func() { _ = localstack.UseNot() }()
	shared.ClearTerminal()
	fmt.Println("Basic Tree With Default Formatting")
	// Create nodes representing a basic tree structure
	root := shared.CreateBasicTreeNodes()
	tree, err := s3.NewTreeFromS3(context.Background(), &s3.InputTreeFromS3{Path: _cs3Testdata},
		treeview.WithExpandAll[treeview.FileInfo]())
	if err != nil {
		panic(err)
	}
	// Render the tree to a string & print it
	output, _ := tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	// Move focus to the second child and re-render
	fmt.Println("Focus moved to second child")
	_, _ = tree.SetFocusedID(context.Background(), root.Children()[0].Children()[0].ID())
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	// Delete focus and re-render
	fmt.Println("Focus deleted")
	_, _ = tree.SetFocusedID(context.Background(), "")
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelay(3.5)
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

func initEnv() {
	ok := localstack.IsRunning()
	if !ok {
		panic("localstack not running")
	}
	_ = localstack.Use()
	defer func() { _ = localstack.UseNot() }()

	goldenDirPath := filepath.Join("..", "..", "..", "internal", "testfixtures")
	_ = localstack.CreateBucket(_myTestBucket, localstack.WithNoErrorIfExist())
	_ = localstack.PutObject(_cGolden100K, filepath.Join(goldenDirPath, _c100K))
	_ = localstack.PutObject(_cGolden1M, filepath.Join(goldenDirPath, _c1M))
	_ = localstack.PutObject(_cs3Testdata+"/golden/recurse/"+_c100K, filepath.Join(goldenDirPath, _c100K))
}
