// v0.4.2
// Author: wunderbarb
//  Apr 2023

package s3

import (
	"testing"
	"time"

	"github.com/wunderbarb/test"
)

func TestIsWithRetry(t *testing.T) {
	_, assert := test.Describe(t)

	var opts []Option
	assert.False(IsWithRetry(opts...))
	opts = append(opts, WithRetry(10, time.Second))
	assert.True(IsWithRetry(opts...))
}
