// v0.7.0
// Author: wunderbarb
//  Aug 2025

package s3

import (
	"context"
	"testing"
	"time"

	"github.com/wunderbarb/test"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

func Test_Join(t *testing.T) {
	_, assert := test.Describe(t)

	tests := []struct {
		els    []string
		result string
	}{
		{[]string{}, ""},
		{[]string{"bkt", "prefix", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "..", "object"}, "s3://bkt/object"},
	}
	for i, tt := range tests {
		assert.Equal(tt.result, Join(tt.els...), "sample %d", i+1)
	}
}

func Test_Size(t *testing.T) {
	require, assert := test.Describe(t)

	s, err := Size(context.Background(), _cGolden100K)
	require.NoError(err)
	assert.Equal(int64(100*1024), s)
	_, err = Size(context.Background(), _cs3Testdata+"/bad")
	assert.Error(err)

	_ = localstack.UseNot()
	defer func() { _ = localstack.Use() }()
	const inProfile = "s3://4test2.ed.techdev.spe.sony.com/test2/sample100K.golden"
	s, err = Size(context.Background(), inProfile, WithProfile("default2"))
	require.NoError(err)
	assert.Equal(int64(102400), s)
}

func Test_bucketFinder(t *testing.T) {
	require, assert := test.Describe(t)

	r, err := bucketFinder("s3://4test.ed.techdev.spe.sony.com")
	require.NoError(err)
	assert.Equal(localstack.DefaultRegion, r)
	isPanic(localstack.UseNot())
	defer func() { _ = localstack.Use() }()
	r, err = bucketFinder("4test.ed.techdev.spe.sony.com")
	require.NoError(err)
	assert.Equal("us-west-2", r)
	r, err = bucketFinder("seff-test-thebride")
	require.NoError(err)
	assert.Equal("us-east-1", r)
}

func TestHasAccess(t *testing.T) {
	require, _ := test.Describe(t)

	require.True(HasAccess(context.Background(), _cS3))
	require.True(HasAccess(context.Background(), Join(_cS3, test.RandomID()), WithRetry(3, 500*time.Millisecond)))
	require.False(HasAccess(context.Background(), "s3://bad"))
	require.False(HasAccess(context.Background(), "s3://bad/bad", WithRetry(3, 500*time.Millisecond)))

	isPanic(localstack.UseNot())
	defer func() { _ = localstack.Use() }()

	require.True(HasAccess(context.Background(),
		"s3://4test2.ed.techdev.spe.sony.com/testdata/SONY_LOGO_SDR_239_ENG_DS_00.mxf", WithProfile("default2")))

}
