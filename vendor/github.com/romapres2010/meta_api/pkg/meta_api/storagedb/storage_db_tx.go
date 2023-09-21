package storagedb

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_sql "github.com/romapres2010/meta_api/pkg/common/sqlxx"
)

// уникальный номер транзакции
var txGlobalID uint64

func getNextTxID() uint64 {
	return atomic.AddUint64(&txGlobalID, 1)
}

type txCacheEntry struct {
	txId    uint64
	tx      *_sql.Tx
	storage *Storage
}

type txCache struct {
	c  map[uint64]*txCacheEntry
	mx sync.RWMutex
}

func newTxCache() (cache *txCache) {
	return &txCache{
		c:  make(map[uint64]*txCacheEntry),
		mx: sync.RWMutex{},
	}
}

func (cache *txCache) Get(txId uint64) (tx *_sql.Tx) {
	if cache != nil && cache.c != nil {
		cache.mx.RLock()
		defer cache.mx.RUnlock()

		if cacheEntry, ok := cache.c[txId]; ok {
			return cacheEntry.tx
		}
		return nil
	}
	return nil
}

func (cache *txCache) Set(txId uint64, tx *_sql.Tx, storage *Storage) {
	if cache != nil && cache.c != nil && tx != nil {
		cache.mx.Lock()
		defer cache.mx.Unlock()

		cache.c[txId] = &txCacheEntry{
			txId:    txId,
			tx:      tx,
			storage: storage,
		}
	}
}

func (cache *txCache) Del(txId uint64) {
	if cache != nil && cache.c != nil {
		cache.mx.Lock()
		defer cache.mx.Unlock()

		delete(cache.c, txId)
	}
}

func (s *Storage) GetTxFromCache(txId uint64) (tx *_sql.Tx, err error) {
	if s != nil && s.db != nil {
		if s.txCache != nil {
			tx = s.txCache.Get(txId)
			if tx != nil {
				return tx, nil
			} else {
				return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("DB '%s' - transaction ID='%v' does not exists", s.db.Name, txId))
			}
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}

func (s *Storage) Begin(ctx context.Context, requestID uint64) (txId uint64, err error) {
	if s != nil && s.db != nil {
		if s.txCache != nil {
			if tx, err := s.db.BeginTxx(ctx, requestID, nil); err != nil {
				return 0, err
			} else {
				txId = getNextTxID()
				s.txCache.Set(txId, tx, s)
				return txId, err
			}
		}
		return 0, nil
	}
	return 0, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}

func (s *Storage) Commit(requestID uint64, txId uint64) (err error) {
	if s != nil && s.db != nil {
		if s.txCache != nil {
			if tx := s.txCache.Get(txId); tx != nil {
				err = s.db.Commit(requestID, tx)
				s.txCache.Del(txId)
				return err
			} else {
				return _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("DB '%s' - transaction ID='%v' does not exists", s.db.Name, txId))
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}

func (s *Storage) Rollback(requestID uint64, txId uint64) (err error) {
	if s != nil && s.db != nil {
		if s.txCache != nil {
			if tx := s.txCache.Get(txId); tx != nil {
				err = s.db.Rollback(requestID, tx)
				s.txCache.Del(txId)
				return err
			} else {
				return _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("DB '%s' - transaction ID='%v' does not exists", s.db.Name, txId))
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}
