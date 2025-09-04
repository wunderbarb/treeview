// v0.2.3
// Author: wunderbarb
//  Sep 2025

// Package localstack manages the localstack emulator and provides some helper functions.
// It is automatically used by the github.com/TechDev-SPE/go-aws library if the environment variable `USE_LOCALSTACK`
// is set.  The preferred way to set this variable is to use `localstack.Use()` and `localstack.UseNot()`.
package localstack

import (
	"os"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/config"
)

const (
	// Env is the environment variable name for the localstack use.  If set, the aws
	// library uses the localstack emulator.  The preferred way to set this variable is to use
	// `localstack.Use()` and `localstack.UseNot()`.
	Env = config.Env
)

// InUse returns true if the application uses the localstack emulator.  LocalStack may run but not used by
// the application.
func InUse() bool {
	return os.Getenv(Env) != ""
}

// Use indicates that the localstack emulator should be used.
func Use() error {
	config.ClearCached()
	return os.Setenv(Env, "true")
}

// UseNot indicates that the localstack emulator should not be used anymore.
func UseNot() error {
	config.ClearCached()
	return os.Unsetenv(Env)
}
