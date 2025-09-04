// v0.1.0
// Author: wunderbarb
// Aug 2024

// Package exec provides a set of system-related functions such a Run that runs an application, grpc settings, or
// universal that operates files and objects seamlessly.
package exec

// options is the structure managing the options of the function Run.
type options struct {
	args        []string
	verbose     bool
	testValue   bool
	retry       bool
	retryNumber int
}

// Option allows parameterizing the Run function.
type Option func(opts *options)

func collectOptions(opts ...Option) options {
	appOpts := options{verbose: false}
	for _, option := range opts {
		option(&appOpts)
	}
	return appOpts
}

// WithArgs sets the arguments `args` used by the application.
func WithArgs(args ...string) Option {
	return func(ap *options) {
		ap.args = append(ap.args, args...)
	}
}

// WithVerbose requests to print on stdout the command issued if an error occurred.
func WithVerbose() Option {
	return func(ap *options) {
		ap.verbose = true
	}
}
