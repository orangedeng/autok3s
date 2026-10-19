package aws

import (
	"errors"

	"github.com/aws/smithy-go"
)

// stringSlice keeps call sites readable where v1 used aws.StringSlice.
func stringSlice(in []string) []string {
	return in
}

// int32Ptr keeps call sites readable where v1 used aws.Int64 and v2 wants *int32.
func int32Ptr(v int64) *int32 {
	c := int32(v)
	return &c
}

// asAPIError extracts a smithy APIError from an SDK error chain.
func asAPIError(err error) (smithy.APIError, bool) {
	var ae smithy.APIError
	ok := errors.As(err, &ae)
	return ae, ok
}
