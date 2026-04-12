package types

import "github.com/regclient/regclient/types/errs"

var (
	// ErrAllRequestsFailed when there are no mirrors left to try
	//
	// Deprecated: replace with [errs.ErrAllRequestsFailed].
	//go:fix inline
	ErrAllRequestsFailed = errs.ErrAllRequestsFailed
	// ErrAPINotFound if an api is not available for the host
	//
	// Deprecated: replace with [errs.ErrAPINotFound].
	//go:fix inline
	ErrAPINotFound = errs.ErrAPINotFound
	// ErrBackoffLimit maximum backoff attempts reached
	//
	// Deprecated: replace with [errs.ErrBackoffLimit].
	//go:fix inline
	ErrBackoffLimit = errs.ErrBackoffLimit
	// ErrCanceled if the context was canceled
	//
	// Deprecated: replace with [errs.ErrCanceled].
	//go:fix inline
	ErrCanceled = errs.ErrCanceled
	// ErrDigestMismatch if the expected digest wasn't received
	//
	// Deprecated: replace with [errs.ErrDigestMismatch].
	//go:fix inline
	ErrDigestMismatch = errs.ErrDigestMismatch
	// ErrEmptyChallenge indicates an issue with the received challenge in the WWW-Authenticate header
	//
	// Deprecated: replace with [errs.ErrEmptyChallenge].
	//go:fix inline
	ErrEmptyChallenge = errs.ErrEmptyChallenge
	// ErrFileDeleted indicates a requested file has been deleted
	//
	// Deprecated: replace with [errs.ErrFileDeleted].
	//go:fix inline
	ErrFileDeleted = errs.ErrFileDeleted
	// ErrFileNotFound indicates a requested file is not found
	//
	// Deprecated: replace with [errs.ErrFileNotFound].
	//go:fix inline
	ErrFileNotFound = errs.ErrFileNotFound
	// ErrHTTPStatus if the http status code was unexpected
	//
	// Deprecated: replace with [errs.ErrHTTPStatus].
	//go:fix inline
	ErrHTTPStatus = errs.ErrHTTPStatus
	// ErrInvalidChallenge indicates an issue with the received challenge in the WWW-Authenticate header
	//
	// Deprecated: replace with [errs.ErrInvalidChallenge].
	//go:fix inline
	ErrInvalidChallenge = errs.ErrInvalidChallenge
	// ErrInvalidReference indicates the reference to an image is has an invalid syntax
	//
	// Deprecated: replace with [errs.ErrInvalidReference].
	//go:fix inline
	ErrInvalidReference = errs.ErrInvalidReference
	// ErrLoopDetected indicates a child node points back to the parent
	//
	// Deprecated: replace with [errs.ErrLoopDetected].
	//go:fix inline
	ErrLoopDetected = errs.ErrLoopDetected
	// ErrManifestNotSet indicates the manifest is not set, it must be pulled with a ManifestGet first
	//
	// Deprecated: replace with [errs.ErrManifestNotSet].
	//go:fix inline
	ErrManifestNotSet = errs.ErrManifestNotSet
	// ErrMissingAnnotation returned when a needed annotation is not found
	//
	// Deprecated: replace with [errs.ErrMissingAnnotation].
	//go:fix inline
	ErrMissingAnnotation = errs.ErrMissingAnnotation
	// ErrMissingDigest returned when image reference does not include a digest
	//
	// Deprecated: replace with [errs.ErrMissingDigest].
	//go:fix inline
	ErrMissingDigest = errs.ErrMissingDigest
	// ErrMissingLocation returned when the location header is missing
	//
	// Deprecated: replace with [errs.ErrMissingLocation].
	//go:fix inline
	ErrMissingLocation = errs.ErrMissingLocation
	// ErrMissingName returned when name missing for host
	//
	// Deprecated: replace with [errs.ErrMissingName].
	//go:fix inline
	ErrMissingName = errs.ErrMissingName
	// ErrMissingTag returned when image reference does not include a tag
	//
	// Deprecated: replace with [errs.ErrMissingTag].
	//go:fix inline
	ErrMissingTag = errs.ErrMissingTag
	// ErrMissingTagOrDigest returned when image reference does not include a tag or digest
	//
	// Deprecated: replace with [errs.ErrMissingTagOrDigest].
	//go:fix inline
	ErrMissingTagOrDigest = errs.ErrMissingTagOrDigest
	// ErrMismatch returned when a comparison detects a difference
	//
	// Deprecated: replace with [errs.ErrMismatch].
	//go:fix inline
	ErrMismatch = errs.ErrMismatch
	// ErrMountReturnedLocation when a blob mount fails but a location header is received
	//
	// Deprecated: replace with [errs.ErrMountReturnedLocation].
	//go:fix inline
	ErrMountReturnedLocation = errs.ErrMountReturnedLocation
	// ErrNoNewChallenge indicates a challenge update did not result in any change
	//
	// Deprecated: replace with [errs.ErrNoNewChallenge].
	//go:fix inline
	ErrNoNewChallenge = errs.ErrNoNewChallenge
	// ErrNotFound isn't there, search for your value elsewhere
	//
	// Deprecated: replace with [errs.ErrNotFound].
	//go:fix inline
	ErrNotFound = errs.ErrNotFound
	// ErrNotImplemented returned when method has not been implemented yet
	//
	// Deprecated: replace with [errs.ErrNotImplemented].
	//go:fix inline
	ErrNotImplemented = errs.ErrNotImplemented
	// ErrNotRetryable indicates the process cannot be retried
	//
	// Deprecated: replace with [errs.ErrNotRetryable].
	//go:fix inline
	ErrNotRetryable = errs.ErrNotRetryable
	// ErrParsingFailed when a string cannot be parsed
	//
	// Deprecated: replace with [errs.ErrParsingFailed].
	//go:fix inline
	ErrParsingFailed = errs.ErrParsingFailed
	// ErrRetryNeeded indicates a request needs to be retried
	//
	// Deprecated: replace with [errs.ErrRetryNeeded].
	//go:fix inline
	ErrRetryNeeded = errs.ErrRetryNeeded
	// ErrShortRead if contents are less than expected the size
	//
	// Deprecated: replace with [errs.ErrShortRead].
	//go:fix inline
	ErrShortRead = errs.ErrShortRead
	// ErrSizeLimitExceeded if contents exceed the size limit
	//
	// Deprecated: replace with [errs.ErrSizeLimitExceeded].
	//go:fix inline
	ErrSizeLimitExceeded = errs.ErrSizeLimitExceeded
	// ErrUnavailable when a requested value is not available
	//
	// Deprecated: replace with [errs.ErrUnavailable].
	//go:fix inline
	ErrUnavailable = errs.ErrUnavailable
	// ErrUnsupported indicates the request was unsupported
	//
	// Deprecated: replace with [errs.ErrUnsupported].
	//go:fix inline
	ErrUnsupported = errs.ErrUnsupported
	// ErrUnsupportedAPI happens when an API is not supported on a registry
	//
	// Deprecated: replace with [errs.ErrUnsupportedAPI].
	//go:fix inline
	ErrUnsupportedAPI = errs.ErrUnsupportedAPI
	// ErrUnsupportedConfigVersion happens when config file version is greater than this command supports
	//
	// Deprecated: replace with [errs.ErrUnsupportedConfigVersion].
	//go:fix inline
	ErrUnsupportedConfigVersion = errs.ErrUnsupportedConfigVersion
	// ErrUnsupportedMediaType returned when media type is unknown or unsupported
	//
	// Deprecated: replace with [errs.ErrUnsupportedMediaType].
	//go:fix inline
	ErrUnsupportedMediaType = errs.ErrUnsupportedMediaType
	// ErrHTTPRateLimit when requests exceed server rate limit
	//
	// Deprecated: replace with [errs.ErrHTTPRateLimit].
	//go:fix inline
	ErrHTTPRateLimit = errs.ErrHTTPRateLimit
	// ErrHTTPUnauthorized when authentication fails
	//
	// Deprecated: replace with [errs.ErrHTTPUnauthorized].
	//go:fix inline
	ErrHTTPUnauthorized = errs.ErrHTTPUnauthorized
)
