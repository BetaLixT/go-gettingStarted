// Package impl containing implementation interface
package impl

import "context"

// IImplementation interface for domain implementation
type IImplementation interface {
	Start(context.Context) error
	Stop(context.Context) error
	StatusCheck(context.Context) error
}
