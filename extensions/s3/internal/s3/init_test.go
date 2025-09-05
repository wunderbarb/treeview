// v0.2.4
// Author: wunderbarb
// Sep 2025

package s3

import (
	"path/filepath"
	"testing"

	"github.com/wunderbarb/test"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

const (
	_myTestBucket = "wunderbarb.example.com"
	_cS3          = _cS3URI + _myTestBucket
	_cs3Testdata  = _cS3 + "/testdata"
	_c100K        = "sample100K.golden"
	_c1M          = "sample1M.golden"
	_cGolden100K  = _cs3Testdata + "/golden/" + _c100K
	_cGolden1M    = _cs3Testdata + "/golden/" + _c1M
	K             = 1024
)

var _goldenDirPath string

func TestMain(m *testing.M) {
	ok := localstack.IsRunning()
	if !ok {
		panic("localstack not running")
	}
	isPanic(localstack.Use())
	defer func() { _ = localstack.UseNot() }()

	_goldenDirPath = filepath.Join("internal", "testfixtures")
	_ = localstack.CreateBucket(_myTestBucket, localstack.WithNoErrorIfExist())
	_ = localstack.PutObject(_cGolden100K, filepath.Join(_goldenDirPath, _c100K))
	_ = localstack.PutObject(_cGolden1M, filepath.Join(_goldenDirPath, _c1M))
	_ = localstack.PutObject(_cs3Testdata+"/golden/recurse/"+_c100K, filepath.Join(_goldenDirPath, _c100K))
	m.Run()
}

func Test_parse(t *testing.T) {
	_, assert := test.Describe(t)

	tests := []struct {
		path string
		bkt  string
		key  string
	}{
		{"s3://sony.com/place/l2", "sony.com", "place/l2"},
		{"sony.com/place/l2", "sony.com", "place/l2"},
		{"s3://sony.com/", "sony.com", ""},
		{"sony.com", "sony.com", ""},
	}
	for _, tt := range tests {
		g1, g2 := Parse(tt.path)
		assert.Equal(tt.bkt, g1)
		assert.Equal(tt.key, g2)
	}
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}
