package storageservice

import (
	"context"

	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// Service - интерфейс обработки данных на стороне хранения (реляционная БД, ...)
type Service interface {
	Begin(ctx context.Context, requestID uint64) (txId uint64, err error)
	Commit(requestID uint64, txId uint64) (err error)
	Rollback(requestID uint64, txId uint64) (err error)
	Create(ctx context.Context, requestID uint64, txId uint64, rowIn *_meta.Object, rowOut *_meta.Object) (err error)
	Update(ctx context.Context, requestID uint64, txId uint64, rowIn *_meta.Object, rowOut *_meta.Object) (err error)
	Execute(ctx context.Context, requestID uint64, txId uint64, command string, args ...interface{}) (err error)
	ExecuteScan(ctx context.Context, requestID uint64, txId uint64, dest []interface{}, command string, args ...interface{}) (exists bool, err error)
	Get(ctx context.Context, requestID uint64, txId uint64, rowOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (exists bool, err error)
	Select(ctx context.Context, requestID uint64, txId uint64, rowsOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (exists bool, err error)
}
