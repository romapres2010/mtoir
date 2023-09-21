package storagedb

import (
	"context"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// Create создать строку в БД
func (s *Storage) Create(ctx context.Context, requestID uint64, txId uint64, rowIn *_meta.Object, rowOut *_meta.Object) (err error) {
	if s != nil && s.db != nil && rowIn != nil && rowOut != nil {
		return s.persist(ctx, requestID, PERSIST_ACTION_CREATE, txId, rowIn, rowOut)
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil && rowIn != nil && rowOut != nil {}", []interface{}{s, rowIn, rowOut}).PrintfError()
}
