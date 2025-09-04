// v0.2.5
// Author: wunderbarb
// Sep 2025

package s3

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	_cS3URI = "s3://"
)

// Parse returns the bucket name and the key of `path`
func Parse(path string) (string, string) {
	const (
		cNumEle = 2
	)
	path = strings.TrimPrefix(path, _cS3URI)
	p := strings.SplitAfterN(path, "/", cNumEle)
	p[0] = strings.TrimSuffix(p[0], "/")
	if len(p) == 1 {
		return p[0], ""
	}
	return p[0], p[1]
}

// UnParse returns the s3URI of the object defined by `bucket` and `key`.
func UnParse(bucket string, key string) string {
	return _cS3URI + bucket + "/" + key
}

func parse1(path string) (*string, *string) {
	a, b := Parse(path)
	return aws.String(a), aws.String(b)
}
