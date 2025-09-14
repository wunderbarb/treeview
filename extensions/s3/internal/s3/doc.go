// Package s3 is a wrapper around aws-sdk-go-v2/s3.
package s3

import (
	"errors"
)

var (
	// ErrMajorFailure is returned when an unexpected serious error occurred.  Contact the development team.
	ErrMajorFailure = errors.New("major failure")
	// ErrNoAccess is returned when the access to the object or bucket is not available.
	ErrNoAccess = errors.New("no access")
)
