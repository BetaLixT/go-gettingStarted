// Package acl defining functionalities and models around managing the Access
// Control List of resources
package acl

import "context"

// IRepository repo interface for managing ACL
type IRepository interface {
	CreateACLEntry(
		ctx context.Context,
		stream string,
		streamID string,
		userType string,
		userID string,
		permissions int,
	) error
	DeleteACLEntry(
		ctx context.Context,
		stream string,
		streamID string,
		userType string,
		userID string,
	) error
	DeleteACLEntries(
		ctx context.Context,
		stream string,
		streamID string,
	) error

	CanRead(
		ctx context.Context,
		stream string,
		streamIds []string,
		userType string,
		userID string,
	) error
	CanWrite(
		ctx context.Context,
		stream string,
		streamIds []string,
		userType string,
		userID string,
	) error
}
