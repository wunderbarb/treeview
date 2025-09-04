// v0.1.11
// Author: wunderbarb
// Sep 2025

package localstack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wunderbarb/test"
)

func TestMain(m *testing.M) {
	isPanic(Use())
	defer func() { _ = UseNot() }()
	_ = m.Run()
}

func TestIsRunning(t *testing.T) {
	require, _ := test.Describe(t)

	require.True(IsRunning())
}

func TestCreateBucket(t *testing.T) {
	require, assert := test.Describe(t)

	bkt := "tst" + test.RandomAlphaString(16, test.Small)
	require.NoError(CreateBucket(bkt))
	assert.Error(CreateBucket(bkt))
	assert.NoError(CreateBucket(bkt, WithNoErrorIfExist()))
	assert.NoError(DeleteBucket(bkt))
	assert.Error(DeleteBucket("s3://bad"))
	assert.Error(CreateBucket("tst" + test.RandomAlphaString(64, test.Small)))
}

func TestCreateBucketAt(t *testing.T) {
	require, assert := test.Describe(t)

	bkt := "tst" + test.RandomAlphaString(16, test.Small)
	const cRegion = "us-west-2"
	require.NoError(CreateBucketAt(bkt, cRegion))
	assert.NoError(DeleteBucketAt(bkt, cRegion))
}

func TestPutObject(t *testing.T) {
	require, assert := test.Describe(t)

	bkt := "tst" + test.RandomAlphaString(16, test.Small)
	isPanic(CreateBucket(bkt))
	defer func() { _ = DeleteBucket(bkt) }()

	hDir, err := os.UserHomeDir()
	isPanic(err)
	require.NoError(PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden")))
	assert.NoError(PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden")))
}

func TestDeleteBucket(t *testing.T) {
	require, _ := test.Describe(t)

	bkt := "tst" + test.RandomAlphaString(16, test.Small)
	isPanic(CreateBucket(bkt))
	hDir, _ := os.UserHomeDir()
	_ = PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden"))

	require.NoError(DeleteBucket(bkt))
}

func TestSetEndPointURLForLambda(t *testing.T) {
	_, assert := test.Describe(t)
	const key = "LOCALSTACK_HOSTNAME"

	adr := "127.27.0.2"
	isPanic(os.Setenv(key, adr))
	defer func() {
		_ = os.Unsetenv(key)
		_endpointURL = "--endpoint-url=http://localhost:4566"
	}()

	SetEndPointURLForLambda()
	assert.Equal("--endpoint-url=http://"+adr+":4566", _endpointURL)
	assert.Equal("http://"+adr+":4566", GetEndPointURL())
}

func TestCreateLambdaFunction(t *testing.T) {
	require, assert := test.Describe(t)

	functionARN, err := CreateLambdaFunction("hello", filepath.Join("..", "lambda", "handler.zip"))
	require.NoError(err)
	assert.NotEmpty(functionARN)
	require.NoError(DeleteLambdaFunction("hello"))
	_, err = CreateLambdaFunction("hello1", "bad.zip")
	assert.Error(err)
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}
