// v0.1.0
// Author: wunderbarb
//  Aug 2025

package s3

import (
	"context"
	"testing"

	"github.com/wunderbarb/test"
)

func TestInfo1(t *testing.T) {
	require, assert := test.Describe(t)

	oi, err := Info(context.Background(), _cs3Testdata)
	require.NoError(err)
	require.NotNil(oi)
	assert.True(oi.IsDir())

	oi, err = Info(context.Background(), _cS3)
	require.NoError(err)
	require.NotNil(oi)
	assert.True(oi.IsDir())

	oi, err = Info(context.Background(), _cGolden100K)
	require.NoError(err)
	require.NotNil(oi)
	assert.False(oi.IsDir())
	assert.Equal("sample100K.golden", oi.Name())
	assert.Equal(int64(100*K), oi.Size())
}
