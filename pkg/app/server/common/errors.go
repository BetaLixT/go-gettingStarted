// All of the app level errors
// first digit identifies the layer (3 = infra)
// the first two digit identify the domain the error
// was created for 99 refers to a non domain specific error

package common

import "github.com/betalixt/gorr"

const (
	InvalidContextProvidedToHandlerErrorCode    = 1_99_000
	InvalidContextProvidedToHandlerErrorMessage = "InvalidContextProvidedToHandlerError"

	UserContextMissingErrorCode    = 1_99_001
	UserContextMissingErrorMessage = "UserContextMissingError"
)

// NewInvalidContextProvidedToHandlerError creates new error
func NewInvalidContextProvidedToHandlerError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    InvalidContextProvidedToHandlerErrorCode,
			Message: InvalidContextProvidedToHandlerErrorMessage,
		},
		500,
		"",
	)
}

// NewUserContextMissingError creates a new user context missing error
func NewUserContextMissingError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    UserContextMissingErrorCode,
			Message: UserContextMissingErrorMessage,
		},
		500,
		"",
	)
}
