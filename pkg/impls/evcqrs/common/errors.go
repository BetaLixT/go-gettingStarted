// All of the implementation level errors
// first digit identifies the layer (3 = infra)
// the first two digit identify the domain the error
// was created for 99 refers to a non domain specific error

package common

import "github.com/betalixt/gorr"

const (
	FailedToAssertContextTypeErrorCode    = 3_99_000
	FailedToAssertContextTypeErrorMessage = "FailedToAssertContextTypeError"

	FailedToAssertDatabaseCtxTypeErrorCode    = 3_99_001
	FailedToAssertDatabaseCtxTypeErrorMessage = "FailedToAssertDatabaseCtxTypeError"

	HexStringGenerationFailedErrorCode    = 3_99_002
	HexStringGenerationFailedErrorMessage = "HexStringGenerationFailedError"

	UnevenKeyValueCountProvidedErrorCode    = 3_99_003
	UnevenKeyValueCountProvidedErrorMessage = "UnevenKeyValueCountProvidedError"

	NonStringKeyProvidedErrorCode    = 3_99_004
	NonStringKeyProvidedErrorMessage = "NonStringKeyProvidedError"

	NoValuesBeingUpdatedErrorCode    = 3_99_005
	NoValuesBeingUpdatedErrorMessage = "NoValuesBeingUpdatedError"
)

func NewFailedToAssertContextTypeError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    FailedToAssertContextTypeErrorCode,
			Message: FailedToAssertContextTypeErrorMessage,
		},
		403,
		"",
	)
}

func NewFailedToAssertDatabaseCtxTypeError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    FailedToAssertDatabaseCtxTypeErrorCode,
			Message: FailedToAssertDatabaseCtxTypeErrorMessage,
		},
		403,
		"",
	)
}

func NewHexStringGenerationFailedError(err error) *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    HexStringGenerationFailedErrorCode,
			Message: HexStringGenerationFailedErrorMessage,
		},
		500,
		err.Error(),
	)
}

func NewUnevenKeyValueCountProvidedError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    UnevenKeyValueCountProvidedErrorCode,
			Message: UnevenKeyValueCountProvidedErrorMessage,
		},
		500,
		"",
	)
}

func NewNonStringKeyProvidedError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    NonStringKeyProvidedErrorCode,
			Message: NonStringKeyProvidedErrorMessage,
		},
		500,
		"",
	)
}

func NewNoValuesBeingUpdatedError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    NoValuesBeingUpdatedErrorCode,
			Message: NoValuesBeingUpdatedErrorMessage,
		},
		400,
		"",
	)
}
