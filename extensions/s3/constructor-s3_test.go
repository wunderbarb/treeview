// v0.1.1
// Author: wunderbarb
//  Sep 2025

package s3

import (
	"context"
	"strings"
	"testing"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/extensions/s3/internal/s3"
)

func TestNewTreeFromS3(t *testing.T) {
	tests := []struct {
		path              string
		symLink           bool
		opt               []treeview.Option[treeview.FileInfo]
		expSuccess        bool
		numberNodes       int
		numberNodesLevel1 int
		numberNodesLevel2 int
	}{
		{_cs3Testdata, false, nil, true,
			1, 1, 3},
		{s3.Join(_cs3Testdata, "golden", "recurse"), false, nil, true,
			1, 1, 0},
		{_cS3, false, []treeview.Option[treeview.FileInfo]{treeview.WithMaxDepth[treeview.FileInfo](2)},
			true, 1, 1, 1},
		{_cs3Testdata, false,
			[]treeview.Option[treeview.FileInfo]{treeview.WithFilterFunc[treeview.FileInfo](func(f treeview.FileInfo) bool {
				return !strings.HasSuffix(f.Path, "golden1")
			})},
			true, 1, 1, 2},
		{"bad", false, nil, false,
			0, 0, 0},
		{_cs3Testdata, true, nil, false,
			0, 0, 0},
		{_cs3Testdata, false, []treeview.Option[treeview.FileInfo]{treeview.WithTraversalCap[treeview.FileInfo](3)},
			false, 0, 0, 0},
	}
	for ii, tt := range tests {
		var opts []treeview.Option[treeview.FileInfo]
		if tt.opt != nil {
			opts = tt.opt
		}
		tr, err := NewTreeFromS3(context.Background(), &InputTreeFromS3{Path: tt.path, FollowSymlinks: tt.symLink},
			opts...)
		if tt.expSuccess != (err == nil) {
			t.Fatalf("expected success %v, got %v", tt.expSuccess, err)
		}
		if err == nil {
			if tr == nil {
				t.Fatal("expected non-nil tree from NewTreeFromS3")
			}
			if len(tr.Nodes()) != tt.numberNodes {
				t.Errorf("expected %d nodes, got %d", tt.numberNodes, len(tr.Nodes()))
			}
			if len(tr.Nodes()[0].Children()) != tt.numberNodesLevel1 {
				t.Errorf("expected %d nodes, got %d", tt.numberNodesLevel1, len(tr.Nodes()[0].Children()))
			}
			if len(tr.Nodes()[0].Children()[0].Children()) != tt.numberNodesLevel2 {
				t.Errorf("expected %d nodes, got %d sample %d", tt.numberNodesLevel2,
					len(tr.Nodes()[0].Children()), ii+1)
			}
		}

	}
}
