// All of the domain level errors (errors used by and known to the domain)
// first digit identifies the layer (2 = domain)
// the first two digit identify the domain the error was created for 00 refers
// to the acl domain, 01 to the foreigns domain, 02 to the uniques domain and
// 99 refers to a non domain specific error,

package common

import "github.com/betalixt/gorr"

const (
	UserACLCheckFailedErrorCode    = 2_00_000
	UserACLCheckFailedErrorMessage = "UserACLCheckFailedError"

	InvalidUserTypeForTaskErrorCode    = 2_03_000
	InvalidUserTypeForTaskErrorMessage = "InvalidUserTypeForTaskError"

	TaskMissingErrorCode    = 2_03_001
	TaskMissingErrorMessage = "TaskMissingError"

	InvalidTaskStatusErrorCode    = 2_03_002
	InvalidTaskStatusErrorMessage = "InvalidTaskStatusError"

	NotPendingTaskErrorCode    = 2_03_003
	NotPendingTaskErrorMessage = "NotPendingTaskError"

	NotProgressTaskErrorCode    = 2_03_004
	NotProgressTaskErrorMessage = "NotPendingTaskError"

	NoTaskUpdatesErrorCode    = 2_03_005
	NoTaskUpdatesErrorMessage = "NoTaskUpdatesError"
)

func NewUserACLCheckFailedError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    UserACLCheckFailedErrorCode,
			Message: UserACLCheckFailedErrorMessage,
		},
		403,
		"",
	)
}

func NewInvalidUserTypeForTaskError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    InvalidUserTypeForTaskErrorCode,
			Message: InvalidUserTypeForTaskErrorMessage,
		},
		403,
		"only users are allowed to create task",
	)
}

func NewTaskMissingError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    TaskMissingErrorCode,
			Message: TaskMissingErrorMessage,
		},
		403,
		"only users are allowed to create task",
	)
}

func NewInvalidTaskStatusError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    InvalidTaskStatusErrorCode,
			Message: InvalidTaskStatusErrorMessage,
		},
		500,
		"only users are allowed to create task",
	)
}

func NewNotPendingTaskError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    NotPendingTaskErrorCode,
			Message: NotPendingTaskErrorMessage,
		},
		400,
		"only users are allowed to create task",
	)
}

// NewNotProgressTaskError provides an error for when the task was expected to
// be in progress but wasn't
func NewNotProgressTaskError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    NotProgressTaskErrorCode,
			Message: NotProgressTaskErrorMessage,
		},
		400,
		"only users are allowed to create task",
	)
}

// NewNoTaskUpdatesError returns error for when no fields are being provided for
// an update
func NewNoTaskUpdatesError() *gorr.Error {
	return gorr.NewError(
		gorr.ErrorCode{
			Code:    NoTaskUpdatesErrorCode,
			Message: NoTaskUpdatesErrorMessage,
		},
		400,
		"",
	)
}
