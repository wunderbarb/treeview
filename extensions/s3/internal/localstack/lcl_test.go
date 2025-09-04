// v0.1.0
// Author: wunderbarb
//  Jan 2023

package localstack

import (
	"testing"

	"github.com/wunderbarb/test"
)

func TestUseNot(t *testing.T) {
	require, assert := test.Describe(t)
	defer func() { _ = Use() }()

	require.True(InUse())
	require.NoError(UseNot())
	assert.False(InUse())
	require.NoError(Use())
	assert.True(InUse())
}
