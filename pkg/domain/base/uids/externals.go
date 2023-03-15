// Package uids defining functionality for generation of unique identifiers
package uids

import "context"

// IRepository repo interface for generating unique ids
type IRepository interface {
	GetID(ctx context.Context) (string, error)
}
