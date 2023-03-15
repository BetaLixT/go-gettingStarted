package repos

import (
	"context"
	"database/sql"
	"math"
	"strconv"
	"strings"
	"time"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/acl"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	domcom "techunicorn.com/udc-core/gettingStarted/pkg/domain/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/entities"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type ACLRepository struct {
	*BaseDataRepository
	rctx    *redis.Client
	lgrf    logger.IFactory
	keySffx string
}

var _ acl.IRepository = (*ACLRepository)(nil)

func NewACLRepository(
	base *BaseDataRepository,
	rctx *redis.Client,
	lgrf logger.IFactory,
) *ACLRepository {
	return &ACLRepository{
		BaseDataRepository: base,
		rctx:               rctx,
		lgrf:               lgrf,
		keySffx:            common.ACLCacheSuffix + domcom.ServiceName + ":",
	}
}

func (r *ACLRepository) CreateACLEntry(
	c context.Context,
	stream string,
	streamID string,
	userType string,
	userID string,
	permissions int,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	entry := entities.ACL{}
	err = dbtx.Get(
		ctx,
		&entry,
		InsertACLQuery,
		stream,
		streamID,
		userType,
		userID,
		permissions,
	)
	if err != nil {
		lgr.Error("failed to create ACL entry")
		return err
	}

	return nil
}

func (r *ACLRepository) DeleteACLEntry(
	c context.Context,
	stream string,
	streamID string,
	userType string,
	userID string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	var entry entities.ACL
	err = dbtx.Get(
		ctx,
		&entry,
		DeleteACLEntryQuery,
		userType,
		userID,
		stream,
		streamID,
	)
	if err != nil {
		lgr.Error("failed to delete entry")
		return err
	}
	err = r.rctx.ZRem(
		ctx,
		r.keySffx+userType+":"+userID,
		generateACLSetMember(entry.Stream, entry.StreamID, entry.Permissions),
	).Err()
	if err != nil {
		lgr.Error("failed to delete cache entry", zap.Error(err))
		return err
	}
	return nil
}

func (r *ACLRepository) DeleteACLEntries(
	c context.Context,
	stream string,
	streamID string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	var entry []entities.ACL
	err = dbtx.Select(
		ctx,
		&entry,
		DeleteACLEntriesQuery,
		stream,
		streamID,
	)
	if err != nil {
		lgr.Error("failed to delete entries", zap.Error(err))
		return err
	}

	rpipe := r.rctx.Pipeline()
	for idx := range entry {
		rpipe.ZRem(
			ctx,
			r.keySffx+entry[idx].UserType+":"+entry[idx].UserId,
			generateACLSetMember(stream, streamID, entry[idx].Permissions),
		)
	}

	_, err = rpipe.Exec(ctx)
	if err != nil {
		lgr.Error("failed to delete cache entries")
		return err
	}
	return nil
}

func (r *ACLRepository) CanRead(
	ctx context.Context,
	stream string,
	streamIDs []string,
	userType string,
	userID string,
) error {
	perm := 0
	var err error
	if len(streamIDs) == 1 {
		perm, err = r.getEntry(
			ctx,
			stream,
			streamIDs[0],
			userType,
			userID,
		)
	} else {
		perm, err = r.getEntries(
			ctx,
			stream,
			streamIDs,
			userType,
			userID,
		)
	}

	if err != nil {
		return err
	}
	if (perm & acl.Read) != 0 {
		return nil
	}
	return domcom.NewUserACLCheckFailedError()
}

func (r *ACLRepository) CanWrite(
	ctx context.Context,
	stream string,
	streamIDs []string,
	userType string,
	userID string,
) error {
	perm := 0
	var err error
	if len(streamIDs) == 1 {
		perm, err = r.getEntry(
			ctx,
			stream,
			streamIDs[0],
			userType,
			userID,
		)
	} else {
		perm, err = r.getEntries(
			ctx,
			stream,
			streamIDs,
			userType,
			userID,
		)
	}

	if err != nil {
		return err
	}
	if (perm & acl.Write) != 0 {
		return nil
	}
	return domcom.NewUserACLCheckFailedError()
}

// getEntries gets acl entries, optimized to get a multiple records
func (r *ACLRepository) getEntries(
	ctx context.Context,
	stream string,
	streamIDs []string,
	userType string,
	userID string,
) (int, error) {
	lgr := r.lgrf.Create(ctx)

	rkey := r.keySffx + userType + ":" + userID
	entries, err := r.rctx.ZRange(
		ctx,
		rkey,
		0,
		100,
	).Result()
	if err != nil {
		// only should happen if the key has been hijacked by something else
		lgr.Error("failed to range through set", zap.Error(err))
	}

	// looping through redis entries, parsing and filtering out required entries
	parsedRes := map[string]int{}
	for idx := range entries {
		split := strings.Split(entries[idx], ":")
		if len(split) != 3 {
			lgr.Warn(
				"invalid split while parsing redis cached acl entry",
				zap.String("entry", entries[idx]),
			)
			continue
		}
		if split[0] == stream {
			per, err := strconv.Atoi(split[2])
			if err != nil {
				lgr.Warn(
					"unabled to parse acl entry's permission field to int",
					zap.String("entry", entries[idx]),
				)
			}
			parsedRes[split[1]] = per
		}
	}

	perm := 0b11
	rpipe := r.rctx.Pipeline()
	defer func() {
		// Keeping the entire LRU cache alive if it's being used, if nmt used for
		// more than two hours for this stream and this user, the cache will be
		// deleted
		rpipe.ExpireAt(
			ctx,
			rkey,
			time.Now().Add(2*time.Hour),
		)
		rpipe.Exec(ctx)
	}()

	nfidx := 0
	notFound := make([]string, len(streamIDs))
	for idx := range streamIDs {
		val, ok := parsedRes[streamIDs[idx]]
		if ok {
			// ACL is implemented as an LRU cache, we increase the score as we get
			// hits and cache values with low hits will be cleaned up
			rpipe.ZIncrBy(
				ctx,
				rkey,
				1,
				generateACLSetMember(stream, streamIDs[idx], val),
			)
			perm = perm & val
		} else {
			notFound[nfidx] = streamIDs[idx]
			nfidx++
		}
		if perm == 0 {
			return perm, nil
		}
	}

	if nfidx == 0 {
		return perm, nil
	}

	// Finding uncached ACL entries
	var dbEntries []entities.ACL
	err = r.dbctx.Select(
		ctx,
		&dbEntries,
		SelectACLEntriesQuery,
		userType,
		userID,
		stream,
		pq.StringArray(notFound[:nfidx]),
	)
	if err != nil {
		lgr.Error("failure while quering database", zap.Error(err))
		return 0, err
	}

	if len(dbEntries) != nfidx {
		lgr.Warn("some acl entries were not present")
		return 0, nil
	}

	mems := make([]*redis.Z, nfidx)
	for idx := range dbEntries {
		perm = perm & dbEntries[idx].Permissions
		mems[idx] = &redis.Z{
			Member: generateACLSetMember(
				stream,
				dbEntries[idx].StreamID,
				dbEntries[idx].Permissions,
			),
			Score: -math.MaxFloat64,
		}
	}
	rpipe.ZAdd(ctx, rkey, mems...)
	return perm, nil
}

// getEntry gets acl entry, optimized to get a single record
func (r *ACLRepository) getEntry(
	ctx context.Context,
	stream string,
	streamID string,
	userType string,
	userID string,
) (int, error) {
	lgr := r.lgrf.Create(ctx)
	rkey := r.keySffx + userType + ":" + userID

	entries, err := r.rctx.ZRange(
		ctx,
		rkey,
		0,
		100,
	).Result()
	if err != nil {
		// only should happen if the key has been hijacked by something else
		lgr.Error("failed to range through set", zap.Error(err))
	}

	perm := 0
	rpipe := r.rctx.Pipeline()
	defer func() {
		// Keeping the entire LRU cache alive if it's being used, if nmt used for
		// more than two hours for this stream and this user, the cache will be
		// deleted
		rpipe.ExpireAt(
			ctx,
			rkey,
			time.Now().Add(2*time.Hour),
		)
		rpipe.Exec(ctx)
	}()

	for idx := range entries {
		if strings.HasPrefix(entries[idx], stream+":"+streamID+":") {
			split := strings.Split(entries[idx], ":")
			if len(split) != 3 {
				lgr.Warn(
					"invalid split while parsing redis cached acl entry",
					zap.String("entry", entries[idx]),
				)
				break
			}
			perm, err = strconv.Atoi(split[2])
			if err != nil {
				lgr.Warn(
					"unabled to parse acl entry's permission field to int",
					zap.String("entry", entries[idx]),
				)
				break
			}
			rpipe.ZIncrBy(
				ctx,
				rkey,
				1,
				generateACLSetMember(stream, streamID, perm),
			)
			break
		}
	}

	if perm != 0 {
		return perm, nil
	}

	var entry entities.ACL
	err = r.dbctx.Get(
		ctx,
		&entry,
		SelectACLEntryQuery,
		userType,
		userID,
		stream,
		streamID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			lgr.Warn("no acl entry found")
			return 0, nil
		}
		lgr.Error("failed to fetch acl entry")
		return 0, err
	}

	rpipe.ZAdd(ctx, rkey, &redis.Z{
		Member: generateACLSetMember(
			stream,
			streamID,
			entry.Permissions,
		),
		Score: -math.MaxFloat64,
	})

	return entry.Permissions, nil
}

func generateACLSetMember(stream string, id string, perm int) string {
	return stream + ":" + id + ":" + strconv.Itoa(perm)
}

// - Queries
const (
	InsertACLQuery = `
	INSERT INTO acl (
		stream,
		stream_id,
		user_type,
		user_id,
	  permissions
	) VALUES (
		$1, $2, $3, $4, $5
	) RETURNING *
	`

	SelectACLEntriesQuery = `
  SELECT * FROM acl
  WHERE user_type = $1 AND user_id = $2 AND stream = $3 AND stream_id = ANY($4)
	`

	SelectACLEntryQuery = `
  SELECT * FROM acl
  WHERE user_type = $1 AND user_id = $2 AND stream = $3 AND stream_id = $4
	`

	DeleteACLEntryQuery = `
  DELETE FROM acl
  WHERE user_type = $1 AND user_id = $2 AND stream = $3 AND stream_id = $4
  RETURNING *
	`

	DeleteACLEntriesQuery = `
  DELETE FROM acl
  WHERE stream = $1 AND stream_id = $2
  RETURNING *
	`
)
