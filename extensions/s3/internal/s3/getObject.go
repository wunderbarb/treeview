// v0.2.13
// Author: wunderbarb
//  Jul 2024

package s3

import (
	"context"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/pkg/errors"
)

// GetClient returns a client for the bucket of `path`.  It handles the option WithProfile.
func GetClient(path string, opts ...Option) (*awsS3.Client, error) {
	b, _ := Parse(path)
	return newClientForBucket(b, opts...)
}

type getObjecter interface {
	GetObject(ctx context.Context, params *awsS3.GetObjectInput,
		optFns ...func(*awsS3.Options)) (*awsS3.GetObjectOutput, error)
}

// getObject returns a GetObjectOutput from the object `path`.  It handles the option WithProfile and passes
// the other ones.
//
// It supports the option WithProfile.
func getObject(ctx context.Context, path string, opts ...Option) (*awsS3.GetObjectOutput, error) {
	c, err := GetClient(path, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	return getObject1(ctx, c, path, opts...)
}

// getObject1 returns a GetObjectOutput from the object `path` using the client `c`. It handles the options
// WithRange, and WithRetry.
// If WithRetry, the default triggered error is "slow down".  The option may add other ones.
func getObject1(ctx context.Context, c getObjecter, path string, opts ...Option) (*awsS3.GetObjectOutput, error) {
	o := collectOptions(opts...)
	b, p := parsePtr(path)

	goi := &awsS3.GetObjectInput{
		Bucket: b,
		Key:    p,
	}
	r1 := retrypolicy.Builder[any]().WithMaxRetries(o.retryMaxAttempt).WithDelay(o.retryLapse).
		HandleIf(isRetryableError).Build()
	a, err := failsafe.Get(func() (any, error) {
		return c.GetObject(ctx, goi)
	}, r1)
	if err != nil {
		return nil, errors.WithMessagef(err, "key %s", path)
	}
	return a.(*awsS3.GetObjectOutput), nil
}

func isRetryableError(_ any, err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() { // Corrected dereferencing
		case "InternalError", "OperationAborted", "RequestTimeout", "ServiceUnavailable", "SlowDown":
			return true
		}
	}
	return false
}

//   - Code: AccessDenied
//   - Code: AllAccessDisabled
//   - Code: AmbiguousGrantByEmailAddress
//   - Code: AuthorizationHeaderMalformed
//   - Code: BadDigest
//   - Code: BucketAlreadyExists
//   - Code: BucketAlreadyOwnedByYou
//   - Code: BucketNotEmpty
//   - Code: CredentialsNotSupported
//   - Code: CrossLocationLoggingProhibited
//   - Code: EntityTooSmall
//   - Code: EntityTooLarge
//   - Code: ExpiredToken
//   - Code: IllegalVersioningConfigurationException
//   - Code: IncompleteBody
//   - Code: IncorrectNumberOfFilesInPostRequest
//   - Code: InlineDataTooLarge
//   - Code: InternalError
//   - Code: InvalidAccessKeyId
//   - Code: InvalidAddressingHeader
//   - Code: InvalidArgument
//   - Code: InvalidBucketName
//   - Code: InvalidBucketState
//   - Code: InvalidDigest
//   - Code: InvalidEncryptionAlgorithmError
//   - Code: InvalidLocationConstraint
//   - Code: InvalidObjectState
//   - Code: InvalidPart
//   - Code: InvalidPartOrder
//   - Code: InvalidPayer
//   - Code: InvalidPolicyDocument
//   - Code: InvalidRange
//   - Code: InvalidRequest
//   - Code: InvalidSecurity
//   - Code: InvalidSOAPRequest
//   - Code: InvalidStorageClass
//   - Code: InvalidTargetBucketForLogging
//   - Code: InvalidToken
//   - Code: InvalidURI
//   - Code: KeyTooLongError
//   - Code: MalformedACLError
//   - Code: MalformedPOSTRequest
//   - Code: MalformedXML
//   - Code: MaxMessageLengthExceeded
//   - Code: MaxPostPreDataLengthExceededError
//   - Code: MetadataTooLarge
//   - Code: MethodNotAllowed
//   - Code: MissingAttachment
//   - Code: MissingContentLength
//   - Code: MissingRequestBodyError
//   - Code: MissingSecurityElement
//   - Code: MissingSecurityHeader
//   - Code: NoLoggingStatusForKey
//   - Code: NoSuchBucket
//   - Code: NoSuchBucketPolicy
//   - Code: NoSuchKey
//   - Code: NoSuchLifecycleConfiguration
//   - Code: NoSuchUpload
//   - Code: NoSuchVersion
//   - Code: NotImplemented
//   - Code: NotSignedUp
//   - Code: OperationAborted
//   - Code: PermanentRedirect
//   - Code: PreconditionFailed
//   - Code: Redirect
//   - Code: RestoreAlreadyInProgress
//   - Code: RequestIsNotMultiPartContent
//   - Code: RequestTimeout
//   - Code: RequestTimeTooSkewed
//   - Code: RequestTorrentOfBucketError
//   - Code: SignatureDoesNotMatch
//   - Code: ServiceUnavailable
//   - Code: SlowDown
//   - Code: TemporaryRedirect
//   - Code: TokenRefreshRequired
//   - Code: TooManyBuckets
//   - Code: UnexpectedContent
//   - Code: UnresolvableGrantByEmailAddress
//   - Code: UserKeyMustBeSpecified
