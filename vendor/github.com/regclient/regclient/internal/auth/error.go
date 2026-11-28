// Copyright the regclient contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"github.com/regclient/regclient/types/errs"
)

var (
	// ErrEmptyChallenge indicates an issue with the received challenge in the WWW-Authenticate header
	//
	// Deprecated: replace with [errs.ErrEmptyChallenge].
	//go:fix inline
	ErrEmptyChallenge = errs.ErrEmptyChallenge
	// ErrInvalidChallenge indicates an issue with the received challenge in the WWW-Authenticate header
	//
	// Deprecated: replace with [errs.ErrInvalidChallenge].
	//go:fix inline
	ErrInvalidChallenge = errs.ErrInvalidChallenge
	// ErrNoNewChallenge indicates a challenge update did not result in any change
	//
	// Deprecated: replace with [errs.ErrNoNewChallenge].
	//go:fix inline
	ErrNoNewChallenge = errs.ErrNoNewChallenge
	// ErrNotFound indicates no credentials found for basic auth
	//
	// Deprecated: replace with [errs.ErrNotFound].
	//go:fix inline
	ErrNotFound = errs.ErrNotFound
	// ErrNotImplemented returned when method has not been implemented yet
	//
	// Deprecated: replace with [errs.ErrNotImplemented].
	//go:fix inline
	ErrNotImplemented = errs.ErrNotImplemented
	// ErrParseFailure indicates the WWW-Authenticate header could not be parsed
	//
	// Deprecated: replace with [errs.ErrParseFailure].
	//go:fix inline
	ErrParseFailure = errs.ErrParsingFailed
	// ErrUnauthorized request was not authorized
	//
	// Deprecated: replace with [errs.ErrUnauthorized].
	//go:fix inline
	ErrUnauthorized = errs.ErrHTTPUnauthorized
	// ErrUnsupported indicates the request was unsupported
	//
	// Deprecated: replace with [errs.ErrUnsupported].
	//go:fix inline
	ErrUnsupported = errs.ErrUnsupported
)
