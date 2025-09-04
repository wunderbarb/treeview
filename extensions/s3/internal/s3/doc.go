// v0.3.0
// Author: wunderbarb
//  Aug 2025

package s3

import (
	"errors"
)

var (
	// ErrInvalidBucketName is returned when the bucket name is invalid.
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html
	ErrInvalidBucketName = errors.New("invalid bucket name")
	// ErrInvalidACL is returned when the ACL passed via WithACL is invalid.
	ErrInvalidACL = errors.New("invalid ACL")
	// ErrObjectNotGlacier is returned when the object queried for restoration is not in Glacier storage.
	ErrObjectNotGlacier = errors.New("object is not in glacier")
	// ErrMajorFailure is returned when an unexpected serious error occurred.  Contact the development team.
	ErrMajorFailure = errors.New("major failure")
	// ErrNoAccess is returned when the access to the object or bucket is not available.
	ErrNoAccess = errors.New("no access")
)
