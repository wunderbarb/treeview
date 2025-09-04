// v0.8.1
// Author: wunderbarb
//  Feb 2025

package localstack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack/exec"
)

const (
	// DefaultRegion is the default region used by localstack.
	DefaultRegion = "us-west-2"
)

const _localstackURL = "https://localhost.localstack.cloud:4566"

var _endpointURL = "--endpoint-url=" + _localstackURL

// ErrInvalidBucketName is returned when the bucket name is invalid.
var ErrInvalidBucketName = errors.New("invalid bucket name")

// Call is the equivalent to call `awslocal` from the command line.
func Call(args ...string) ([]byte, error) {
	a := []string{_endpointURL}
	return exec.Run("aws", exec.WithArgs(append(a, args...)...), exec.WithVerbose())
}

// CallSilent is the equivalent to call `awslocal` from the command line.
func CallSilent(args ...string) ([]byte, error) {
	a := []string{_endpointURL}
	return exec.Run("aws", exec.WithArgs(append(a, args...)...))
}

// CreateBucket creates the bucket `bktName` on the localstack.
// For testing purpose exclusively.
// It supports option WithNoErrorIfExist.
func CreateBucket(bktName string, opt ...Option) error {
	return CreateBucketAt(bktName, DefaultRegion, opt...)
}

// CreateBucketAt creates the bucket `bktName` on the localstack in the AWS region.
// For testing purpose exclusively.
// It supports option WithNoErrorIfExist.
func CreateBucketAt(bktName string, region string, opt ...Option) error {
	b := parse(bktName)
	if !IsBucketNameValid(b) {
		return ErrInvalidBucketName
	}
	if collectOptions(opt...).noErrIfExist {
		_, _ = CallSilent("s3api", "create-bucket",
			"--bucket", b,
			"--region", region,
			"--create-bucket-configuration", "{\"LocationConstraint\": \""+region+"\"}")
		return nil
	}
	_, err := Call("s3api", "create-bucket",
		"--bucket", b,
		"--region", region,
		"--create-bucket-configuration", "{\"LocationConstraint\": \""+region+"\"}")

	return err
}

type funcAnswer struct {
	FunctionName string `json:"FunctionName"`
	FunctionArn  string `json:"FunctionArn"`
	// Runtime      string `json:"Runtime"`
	// Role         string `json:"Role"`
	// Handler      string `json:"Handler"`
	// CodeSize     int    `json:"CodeSize"`
	// Description  string `json:"Description"`
	// Timeout      int    `json:"Timeout"`
	// LastModified string `json:"LastModified"`
	// CodeSha256   string `json:"CodeSha256"`
	// Version      string `json:"Version"`
	// VpcConfig    struct {
	// } `json:"VpcConfig"`
	// TracingConfig struct {
	// 	Mode string `json:"Mode"`
	// } `json:"TracingConfig"`
	// RevisionId       string `json:"RevisionId"`
	// State            string `json:"State"`
	// LastUpdateStatus string `json:"LastUpdateStatus"`
	// PackageType      string `json:"PackageType"`
}

type answer struct {
	Functions []funcAnswer `json:"Functions"`
}

// CreateLambdaFunction creates the lambda function `functionName` with the runtime provided.al2023 on the `localstack` using
// the zipped file `zipFilePath`.  It returns the address of the function.  Furthermore, it sets the function to use
// localstack automatically. The timeout is set to 900 seconds.
// If the function already exists, it is updated with the provided parameters.
// Exclusively for testing purpose.
// It supports option environmental WithEnvironmentVariable.
func CreateLambdaFunction(functionName string, zipFilePath string, opts ...Option) (string, error) {
	v := collectOptions(opts...).getEnv()
	abs, err := filepath.Abs(zipFilePath)
	if err != nil {
		return "", err
	}
	data, err := Call("lambda", "list-functions")
	if err != nil {
		return "", err
	}
	var a answer
	err = json.Unmarshal(data, &a)
	if err != nil {
		return "", err
	}
	var found bool
	for _, f := range a.Functions {
		if f.FunctionName == functionName {
			found = true
			break
		}
	}
	switch found {
	case false:
		data, err = Call("lambda", "create-function", "--function-name", functionName,
			"--environment", v,
			"--runtime", "provided.al2023",
			"--timeout", "900",
			"--role", "arn:aws:iam::123456789012:role/fake-role",
			"--handler", "bootstrap", "--zip-file", "fileb://"+abs)
		if err != nil {
			return "", err
			// The error was that the function already existed.
		}
	case true:
		data, err = Call("lambda", "update-function-code", "--function-name", functionName, "--zip-file",
			"fileb://"+abs)
		if err != nil {
			return "", err
		}
		if len(v) != 0 {
			_, err = Call("lambda", "update-function-configuration", "--function-name", functionName,
				"--environment", v)
			if err != nil {
				return "", err
			}
		}
	}
	var af funcAnswer
	err = json.Unmarshal(data, &af)
	if err != nil {
		return "", err
	}

	// https://docs.localstack.cloud/references/lambda-provider-v2/#function-in-pending-state
	_, err = Call("lambda", "wait", "function-active-v2", "--function-name", functionName)
	if err != nil {
		return "", err
	}
	return af.FunctionArn, nil
}

// DeleteBucket deletes the bucket `bktName` on the localstack.
func DeleteBucket(bktName string) error {
	return DeleteBucketAt(bktName, DefaultRegion)
}

// DeleteBucketAt deletes the bucket `bktName` on the localstack.  The bucket does not need to be empty.
func DeleteBucketAt(bktName string, region string) error {
	type answer struct {
		Contents []struct {
			Key          string    `json:"Key"`
			LastModified time.Time `json:"LastModified"`
			ETag         string    `json:"ETag"`
			Size         int       `json:"Size"`
			StorageClass string    `json:"StorageClass"`
		} `json:"Contents"`
	}
	b := parse(bktName)
	// list all potential objects
	data, err := Call("s3api", "list-objects-v2", "--bucket", b, "--region", region)
	if err != nil {
		return err
	}
	if len(data) != 0 { // if there are objects
		var a answer
		err = json.Unmarshal(data, &a)
		if err != nil {
			return err
		}
		for _, c := range a.Contents { // delete all objects
			_, err = Call("s3api", "delete-object", "--bucket", b, "--key", c.Key, "--region", region)
			if err != nil {
				return err
			}
		}
	}
	// delete the emptied bucket
	_, err = Call("s3api", "delete-bucket", "--bucket", b, "--region", region)
	return err
}

// DeleteLambdaFunction deletes the lambda function `functionName` on the localstack.
func DeleteLambdaFunction(functionName string) error {
	_, err := Call("lambda", "delete-function", "--function-name", functionName)
	if err != nil {
		return err
	}
	return nil
}

// GetEndPointURL returns the current endpoint URL of the localstack.
func GetEndPointURL() string {
	return strings.TrimPrefix(_endpointURL, "--endpoint-url=")
}

// IsBucketNameValid returns true if the bucket name is syntactically valid.
func IsBucketNameValid(b string) bool {
	return regexp.MustCompile(`^[a-z0-9.\-]+$`).MatchString(b) && !(len(b) < 1 || len(b) > 63)
}

// IsExist checks whether the object `objectPath` exists on the localstack.
func IsExist(objectPath string) bool {
	a, _ := exec.Run("aws", exec.WithArgs(_endpointURL, "s3", "ls", objectPath))
	return len(a) > 0
}

// IsRunning checks if the localstack is running.
func IsRunning() bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(_localstackURL)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}

// PutObject pushes the file `source` on the localstack as object `objectPath`.
// For testing purpose exclusively.  If it is already in the bucket, the operation is skipped.
func PutObject(objectPath string, source string) error {
	if IsExist(objectPath) {
		return nil
	}
	_, err := Call("s3", "cp", "--region", DefaultRegion, source, objectPath)
	return err
}

// SetEndPointURLForLambda sets the endpoint URL for the localstack to a different value than
// "--endpoint-url=http://localhost:4566" to be used in localstack lambda.  It is necessary to
// set the endpoint URL for the lambda function to be able to use the localstack S3.
//
// See https://levelup.gitconnected.com/aws-run-an-s3-triggered-lambda-locally-using-localstack-ac05f03dc896
func SetEndPointURLForLambda() {
	_endpointURL = fmt.Sprintf("--endpoint-url=http://%s:4566", os.Getenv("LOCALSTACK_HOSTNAME"))
}

// parse returns the bucket name
func parse(path string) string {
	const (
		cNumEle = 2
	)
	path = strings.TrimPrefix(path, "s3://")
	p := strings.SplitAfterN(path, "/", cNumEle)
	p[0] = strings.TrimSuffix(p[0], "/")
	if len(p) == 1 {
		return p[0]
	}
	return p[0]
}
