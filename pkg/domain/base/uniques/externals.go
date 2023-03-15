package uniques

import "context"

// The uniques repository can be used to a unique constraint, tying a property
// and value to a particular entity in a stream, this aims to bring unique
// constraints to the business layer (a unique constraint is now business logic)
type IRepository interface {
	RegisterConstraint(
		ctx context.Context,
		stream string,
		streamId string,
		sagaId *string,
		property string,
		value string,
	) error
	RemoveConstraint(
		ctx context.Context,
		stream string,
		property string,
		value string,
	) error
}
