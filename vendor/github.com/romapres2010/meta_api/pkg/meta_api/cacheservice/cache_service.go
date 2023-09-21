package cacheservice

import _meta "github.com/romapres2010/meta_api/pkg/common/meta"

// CacheService - интерфейс для работы с кэшем
type CacheService interface {
	CloseAll()
	ClearAll()
	Clear(entity *_meta.Entity)
	Lock(entity *_meta.Entity)
	Unlock(entity *_meta.Entity)
	RLock(entity *_meta.Entity)
	RUnlock(entity *_meta.Entity)

	//SetPtr(entity *_meta.Entity, key *_meta.Key, v interface{}, keyArgs ...interface{}) bool
	//SetPtrUnsafe(entity *_meta.Entity, key *_meta.Key, v interface{}, keyArgs ...interface{}) bool
	//GetPtr(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) (interface{}, bool)
	//GetPtrUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) (interface{}, bool)
	//DelPtr(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{})
	//DelPtrUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{})

	Set(key *_meta.Key, object *_meta.Object, keyArgs ...interface{}) bool
	SetUnsafe(key *_meta.Key, object *_meta.Object, keyArgs ...interface{}) bool
	Get(entity *_meta.Entity, key *_meta.Key, keepRLock bool, keyArgs ...interface{}) (*_meta.Object, bool)
	GetUnsafe(entity *_meta.Entity, key *_meta.Key, keepRLock bool, keyArgs ...interface{}) (*_meta.Object, bool)
	Del(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{})
	DelUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{})
}
