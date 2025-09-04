// v0.1.16
// Author: wunderbarb
//  Aug 2024

package s3

import (
	"context"
	"io/fs"
	"testing"

	"github.com/wunderbarb/test"
)

func Test_ReadDir1(t *testing.T) {
	require, assert := test.Describe(t)

	a, err := ReadDir1(context.Background(), _cs3Testdata+"/golden")
	require.NoError(err)
	require.LessOrEqual(2, len(a))
	var as []string
	for _, dir := range a {
		as = append(as, dir.Name())
	}
	assert.Contains(as, "sample100K.golden")
	assert.Contains(as, "sample1M.golden")
	a, err = ReadDir1(context.Background(), _cs3Testdata+"/golden/")
	require.NoError(err)
	require.LessOrEqual(2, len(a))
	var as4 []string
	for _, dir := range a {
		as4 = append(as4, dir.Name())
	}
	assert.Contains(as4, "sample100K.golden")
	assert.Contains(as4, "sample1M.golden")

	a, err = ReadDir1(context.Background(), _cs3Testdata)
	require.NoError(err)
	require.LessOrEqual(1, len(a))
	var as2 []string
	for _, dir := range a {
		as2 = append(as2, dir.Name())
	}
	assert.Contains(as2, "golden")

	_, err = ReadDir1(context.Background(), "badBucket/golden")
	assert.Error(err)

	// localstack.UseNot()
	// defer localstack.Use()
	// a, err = ReadDir1(context.Background(), "s3://4test.ed.techdev.spe.sony.com/sample/BladeRunnr2049")
	// require.NoError(err)
	// assert.Len(a, 10)
	// var as3 []string
	// for _, dir := range a {
	// 	as3 = append(as3, dir.Name())
	// }
	// assert.Contains(as3, "BladeRunnr2049_FTR-2D-DVis_S_EN-ES_ES_71-Atmos_4K_SPE_20170904_DGB_SMPTE_OV")
}

func Test_DirEntry_is_interface(t *testing.T) {
	_, _ = test.Describe(t)

	var _ fs.DirEntry = (*DirEntry)(nil)
}

func Test_IsDir(t *testing.T) {
	_, assert := test.Describe(t)

	tests := []struct {
		path       string
		expSuccess bool
	}{
		{_cs3Testdata + "/golden", true},
		{_cs3Testdata + "/golden/", true},
		{_cs3Testdata + "/golden/sample", false},
		{_cs3Testdata, true},
		{"s3://bad", false},
	}
	for i, tt := range tests {
		assert.Equal(tt.expSuccess, IsDir1(context.Background(), tt.path), "sample %d", i+1)
	}
}

func Test_Base(t *testing.T) {
	_, assert := test.Describe(t)

	tests := []struct {
		path   string
		result string
	}{
		{"s3://bucket/prefix/object/", "object"},
		{"s3://bucket/prefix", "prefix"},
		{"s3://bucket", ""},
		{"", "."},
	}
	for _, tt := range tests {
		assert.Equal(tt.result, Base(tt.path))
	}
}
