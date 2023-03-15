package foreigns

import "context"

// The foreigns repository can be used to register foreign entities and create
// constraints for entities that are managed by the service, this aims to bring
// foreign constraints to the business logic layer (a foreign constraint is now
// business logic)
type IRepository interface {

  // Registers a new foreign item for the service to track
	RegisterForeignItem(
		ctx context.Context,
		sagaId *string,
		foreignStream string,
		foreignStreamId string,
	) error

	// Removes a foreign item, will fail if the item is registered in a constraint
	RemoveForeignItem(
		ctx context.Context,
		foreignStream string,
		foreignStreamId string,
	) error

	// Registers a new foreign constraint
	RegisterConstraint(
		ctx context.Context,
		sagaId *string,
		foreignStream string,
		foreignStreamId string,
		stream string,
		streamId string,
	) error

	// Removes a foreign constraint
	RemoveConstraint(
		ctx context.Context,
		foreignStream string,
		foreignStreamId string,
		stream string,
		streamId string,
	) error

	// List objects tied to a foreign object
	ListAssociatedObjects(
		ctx context.Context,
		foreignStream string,
		foreignStreamId string,
	) ([]Object, error)
}
