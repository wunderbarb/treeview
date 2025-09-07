// v0.1.2
// Author: wunderbarb
// Sep 2025

package s3

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

const (
	_myTestBucket = "wunderbarb.example.com"
	_cS3          = _cS3URI + _myTestBucket
	_cS3URI       = "s3://"
	_cs3Testdata  = _cS3 + "/testdata"
	_c100K        = "sample100K.golden"
	_c1M          = "sample1M.golden"
	_cGolden100K  = _cs3Testdata + "/golden/" + _c100K
	_cGolden1M    = _cs3Testdata + "/golden/" + _c1M + "1"
)

func TestMain(m *testing.M) {
	ok := localstack.IsRunning()
	if !ok {
		panic("localstack not running")
	}
	isPanic(localstack.Use())
	defer func() { _ = localstack.UseNot() }()

	dir, err := os.UserHomeDir()
	isPanic(err)

	// s3://<_myTestBucket>
	//    |- testdata
	//       |-   golden
	//          |- recurse
	//          |   |- <_c100K>
	//          |- <_c100K>
	//          |- <_c1M>
	_goldenDirPath := filepath.Join(dir, "Dev", "golden")
	_ = localstack.CreateBucket(_myTestBucket, localstack.WithNoErrorIfExist())
	_ = localstack.PutObject(_cGolden100K, filepath.Join(_goldenDirPath, _c100K))
	_ = localstack.PutObject(_cGolden1M, filepath.Join(_goldenDirPath, _c1M))
	_ = localstack.PutObject(_cs3Testdata+"/golden/recurse/"+_c100K, filepath.Join(_goldenDirPath, _c100K))
	m.Run()
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}
