package lyvecloud

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	ErrBadRequest                          = "BadRequest"
	ErrCodeBucketNotEmpty                  = "BucketNotEmpty"
	ErrCodeObjectLockConfigurationNotFound = "ObjectLockConfigurationNotFoundError"
	ErrCodeOperationAborted                = "OperationAborted"
	ErrCodeUnauthorized                    = 401
)
