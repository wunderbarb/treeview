package s3

import (
	"context"
	"io/fs"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
)

const defaultMode = 0o755

// ObjectInfo holds the information related to an S3 object.  It implements the interface io/fs/FileInfo.
type ObjectInfo struct {
	// bucket provides the name of the bucket holding the object.
	bucket       string
	key          string
	lastModified time.Time
	size         int64
	storageClass string
	isDir        bool
}

// Info returns the ObjectInfo of the object at `path`.  If it is a directory, then IsDir() is true and Size() and
// ModTime() is meaningless.
func Info(ctx context.Context, path string, opts ...Option) (*ObjectInfo, error) {
	if !HasAccess(ctx, path, opts...) {
		return nil, ErrNoAccess
	}
	b, p := Parse(path)
	oi := &ObjectInfo{
		bucket:       b,
		key:          p,
		lastModified: time.Time{},
		size:         0,
		storageClass: "STANDARD",
		isDir:        true,
	}
	if p == "" {
		return oi, nil
	}
	oi.isDir = IsDir(ctx, path, opts...)
	if !oi.isDir {
		hoOut, err := getHeadObject(ctx, path, opts...)
		if err != nil {
			return nil, err
		}
		oi.lastModified = *hoOut.LastModified
		oi.size = *hoOut.ContentLength
		oi.storageClass = string(hoOut.StorageClass)
	}
	if oi.storageClass == "" {
		oi.storageClass = "STANDARD" // Provides storage class information of the object. Amazon S3 returns this header
		// for all objects except for S3 Standard storage class objects.
	}
	return oi, nil
}

// Name returns the base name of the object as defined by fs.FileInfo.
func (oi *ObjectInfo) Name() string {
	return Base(Join(oi.bucket, oi.key))
}

// Mode is mandatory for the fs.FileInfo interface.
func (oi *ObjectInfo) Mode() fs.FileMode {

	if oi.isDir {
		return fs.ModeDir | defaultMode
	}
	return fs.ModeIrregular
}

// ModTime returns the last modification time of the object.
func (oi *ObjectInfo) ModTime() time.Time {
	return oi.lastModified
}

// Info implements the interface fs.FileInfo.
func (oi *ObjectInfo) Info() (fs.FileInfo, error) {
	if oi == nil {
		return nil, fs.ErrNotExist
	}
	return oi, nil
}

// IsDir determines if the given ObjectInfo represents a directory based on its bucket and key attributes.
func (oi *ObjectInfo) IsDir() bool {
	return oi.isDir
}

// Sys is a placeholder method to satisfy the fs.FileInfo interface, always returning nil.
func (oi *ObjectInfo) Sys() any {
	return nil
}

// Bucket returns the name of the bucket holding the object.
func (oi *ObjectInfo) Bucket() string {
	return oi.bucket
}

// LastModified returns the last modification date of the object.
func (oi *ObjectInfo) LastModified() time.Time {
	return oi.lastModified
}

// Path returns the path of the object without the bucket.
func (oi *ObjectInfo) Path() string {
	return oi.key
}

// Size returns the size of the object.
func (oi *ObjectInfo) Size() int64 {
	return oi.size
}

// StorageClass returns the storage class of the object.  It has one of the following values: "STANDARD",
// "REDUCED_REDUNDANCY", "GLACIER", "STANDARD_IA", "ONEZONE_IA", "INTELLIGENT_TIERING", "DEEP_ARCHIVE", "OUTPOSTS",
// or "GLACIER_IR".
func (oi *ObjectInfo) StorageClass() string {
	return oi.storageClass
}

// ListAllObjectsAndPrefixes recursively lists all objects and prefixes in the specified path and returns a slice of
// ObjectInfo.  It does not list the prefixes.
// It hides the pagination, i.e., it may return more than 1,000 objects.
func ListAllObjectsAndPrefixes(ctx context.Context, path string, opts ...Option) ([]ObjectInfo, error) {
	return getListObjectsRecurse(ctx, path, false, opts...)
}

func getListObjectsRecurse(ctx context.Context, path string, withoutDir bool, opts ...Option) ([]ObjectInfo, error) {
	b, p := parsePtr(path)
	c, err := newClientForBucket(*b, opts...)
	if err != nil {
		return nil, err
	}
	var ois []ObjectInfo
	commonPrefixes := map[string]bool{}
	prefix := strings.TrimSuffix(*p, "/") + "/"
	if prefix == "/" { // when starting from the root of the bucket
		prefix = ""
	}
	var token *string
	token = nil
	for {
		lov2i := &awsS3.ListObjectsV2Input{
			Bucket:            b,
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		}
		lov2o, err := c.ListObjectsV2(ctx, lov2i)
		if err != nil {
			return nil, errors.Wrap(err, "ListObjectsV2")
		}
		for _, content := range lov2o.Contents {
			if !strings.HasSuffix(*content.Key, "/") {
				ois = append(ois, typesObjectToObjectInfo(&content, *b))
			}
			if withoutDir {
				continue
			}
			key := strings.TrimPrefix(*content.Key, prefix)
			l := strings.Split(key, "/")
			if len(l) == 1 {
				continue
			}
			l0 := ""
			for _, l1 := range l[:len(l)-1] {
				l0 += l1
				_, ok := commonPrefixes[l0]
				if !ok {
					commonPrefixes[l0] = true
					ois = append(ois, ObjectInfo{
						bucket:       *b,
						key:          prefix + l0 + "/",
						lastModified: time.Time{},
						size:         0,
						storageClass: "",
						isDir:        true,
					})
				}
				l0 += "/"
			}
		}
		if !*lov2o.IsTruncated {
			break
		}
		token = lov2o.NextContinuationToken
	}
	return ois, nil
}

func typesObjectToObjectInfo(to *types.Object, bucket string) ObjectInfo {
	return ObjectInfo{
		bucket:       bucket,
		key:          *to.Key,
		lastModified: *to.LastModified,
		size:         *to.Size,
		storageClass: string(to.StorageClass),
		isDir:        false,
	}
}
