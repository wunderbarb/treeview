// v0.1.1
// Author: wunderbarb
//  Sep 2025

package s3

import (
	"context"
	"testing"

	"github.com/wunderbarb/test"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

func TestNewTreeFromS3(t *testing.T) {
	require, assert := test.Describe(t)

	tr, err := NewTreeFromS3(context.Background(), _cs3Testdata, false)
	require.NoError(err)
	require.NotNil(tr)
	assert.Len(tr.Nodes(), 1)
	assert.Len(tr.Nodes()[0].Children(), 1)
	assert.Len(tr.Nodes()[0].Children()[0].Children(), 3)

	tr, err = NewTreeFromS3(context.Background(), _cs3Testdata+"/golden/recurse", false)
	require.NoError(err)
	require.NotNil(tr)
	assert.Len(tr.Nodes(), 1)
	assert.Len(tr.Nodes()[0].Children(), 1)

	tr, err = NewTreeFromS3(context.Background(), _cs3Testdata+"/golden/recurse", false)
	require.NoError(err)
	require.NotNil(tr)
	assert.Len(tr.Nodes(), 1)
	assert.Len(tr.Nodes()[0].Children(), 1)

	_ = localstack.UseNot()
	defer func() {
		_ = localstack.Use()
	}()
	tr, err = NewTreeFromS3(context.Background(), "s3://4test.ed.techdev.spe.sony.com/sample/BladeRunnr2049/", false)
	require.NoError(err)
	require.NotNil(tr)
}
