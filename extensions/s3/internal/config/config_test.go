// v0.1.0
// Author: wunderbarb
//  Jul 2023

package config

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/wunderbarb/test"
)

func TestMain(m *testing.M) {
	_ = os.Setenv(Env, "true")
	defer func() { _ = os.Unsetenv(Env) }()
	m.Run()
}

func TestNewConfig(t *testing.T) {
	require, assert := test.Describe(t)

	c, err := NewConfig()
	require.NoError(err)
	assert.NotEqual(aws.Config{}, c)
	_, err = NewConfig()
	require.NoError(err)
	// Test cache
	_, ok := _cached.get()
	require.True(ok)
	ClearCached()
	_, ok = _cached.get()
	require.False(ok)
}
