// v0.1.0
// Author: wunderbarb
//  Mar 2023

package localstack

import (
	"testing"

	"github.com/wunderbarb/test"
)

func TestWithEnvironmentVariable(t *testing.T) {
	require, assert := test.Describe(t)

	o := collectOptions(WithEnvironmentVariable("Test", "true"))
	require.NotNil(o)
	s := o.getEnv()
	assert.NotEmpty(s)
	assert.Equal(`{"Variables":{"Test":"true","USE_LOCALSTACK":"true"}}`, s)
}
