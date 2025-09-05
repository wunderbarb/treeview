// v0.1.1
// Author: wunderbarb
//  Sep 2025

package s3

import (
	"context"
	"testing"

	"github.com/wunderbarb/test"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/s3"
)

func TestNewTreeFromS3(t *testing.T) {
	require, assert := test.Describe(t)

	tr, err := NewTreeFromS3(context.Background(), &InputTreeFromS3{Path: _cs3Testdata})
	require.NoError(err)
	require.NotNil(tr)
	assert.Len(tr.Nodes(), 1)
	assert.Len(tr.Nodes()[0].Children(), 1)
	assert.Len(tr.Nodes()[0].Children()[0].Children(), 3)

	tr, err = NewTreeFromS3(context.Background(), &InputTreeFromS3{Path: s3.Join(_cs3Testdata, "golden", "recurse")})
	require.NoError(err)
	require.NotNil(tr)
	assert.Len(tr.Nodes(), 1)
	assert.Len(tr.Nodes()[0].Children(), 1)

	// errors
	_, err = NewTreeFromS3(context.Background(), &InputTreeFromS3{Path: "bad"})
	assert.Error(err)
	_, err = NewTreeFromS3(context.Background(), &InputTreeFromS3{Path: _cs3Testdata, FollowSymlinks: true})
	assert.ErrorIs(err, ErrNotYetSupported)
}
