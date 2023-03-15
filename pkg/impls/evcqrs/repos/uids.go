package repos

import (
	"context"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/uids"

	"github.com/bwmarrin/snowflake"
)

// UIDRepository for generating unique ids
type UIDRepository struct {
	sf *snowflake.Node
}

var _ uids.IRepository = (*UIDRepository)(nil)

// NewUIDRepository Constructs new UUIDRepository
func NewUIDRepository(
	sf *snowflake.Node,
) *UIDRepository {
	return &UIDRepository{
		sf: sf,
	}
}

// GetID generates and returns a unique id
func (r *UIDRepository) GetID(
	ctx context.Context,
) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return r.sf.Generate().String(), nil
}
