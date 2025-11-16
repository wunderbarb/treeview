package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	_cS3URI = "s3://"
)

// Base returns the last element of `path`.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns "."
func Base(path string) string {
	if path == "" {
		return "."
	}
	_, p := Parse(path)
	l := strings.Split(strings.TrimRight(p, "/"), "/")
	return l[len(l)-1]
}

// IsDir informs whether the path is a key with objects.  The path may end with "/".
// The root of an accessible bucket is a directory.
//
//	If the bucket is not accessible, it returns false.
func IsDir(ctx context.Context, path string, opts ...Option) bool {
	b, p := parsePtr(path)
	cl, err := newClientForBucket(*b, opts...)
	if err != nil {
		return false
	}
	var token *string
	token = nil
	for {
		lov2i := &awsS3.ListObjectsV2Input{
			Bucket:            b,
			Prefix:            p,
			ContinuationToken: token,
		}
		lov2o, err := cl.ListObjectsV2(ctx, lov2i)
		if err != nil {
			return false
		}
		if strings.HasSuffix(path, "/") {
			return true // It exists, and it ends with /, thus it is a dir.
		}
		for _, o := range lov2o.Contents {
			oWithoutPrefix := strings.TrimPrefix(*o.Key, *p)
			if strings.HasPrefix(oWithoutPrefix, "/") {
				return true
			}
		}
		if !*lov2o.IsTruncated {
			break
		}
		token = lov2o.NextContinuationToken
	}
	return false
}

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

// -------------------------------------------

func parsePtr(path string) (*string, *string) {
	a, b := Parse(path)
	return aws.String(a), aws.String(b)
}
