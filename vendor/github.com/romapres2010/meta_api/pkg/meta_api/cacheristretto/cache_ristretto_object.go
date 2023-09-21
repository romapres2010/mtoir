package cacheristretto

import (
	"reflect"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

func (s *Service) Set(key *_meta.Key, object *_meta.Object, keyArgs ...interface{}) bool {
	if s.cache != nil && object != nil && key != nil && object.Entity != nil {
		s.RLock(object.Entity)
		defer s.RUnlock(object.Entity)
		//s.Lock(object.Entity)
		//defer s.Unlock(object.Entity)
		return s.SetUnsafe(key, object, keyArgs...)
	}
	return false
}

func (s *Service) Get(entity *_meta.Entity, key *_meta.Key, keepRLock bool, keyArgs ...interface{}) (*_meta.Object, bool) {
	if s.cache != nil && entity != nil && key != nil && len(keyArgs) > 0 {
		s.RLock(entity)
		defer s.RUnlock(entity)
		return s.GetUnsafe(entity, key, keepRLock, keyArgs...)
	}
	return nil, false
}

func (s *Service) Del(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) {
	if s.cache != nil && entity != nil && key != nil {
		s.RLock(entity)
		defer s.RUnlock(entity)
		//s.Lock(entity)
		//defer s.Unlock(entity)
		s.DelUnsafe(entity, key, keyArgs...)
	}
}

func (s *Service) SetUnsafe(key *_meta.Key, object *_meta.Object, keyArgs ...interface{}) bool {
	if s.cache != nil && object != nil && key != nil && object.Entity != nil {
		cacheKey := formatCacheKeyStr(object.Entity, key, "__ALL__", keyArgs...)
		entry := newCacheEntry(key, cacheKey, object)

		return s.setUnsafe(object.Entity, cacheKey, entry)
	}
	return false
}
func (s *Service) GetUnsafe(entity *_meta.Entity, key *_meta.Key, keepRLock bool, keyArgs ...interface{}) (*_meta.Object, bool) {
	if s.cache != nil && entity != nil && key != nil && len(keyArgs) > 0 {
		cacheKey := formatCacheKeyStr(entity, key, "__ALL__", keyArgs...)

		entry, exists := s.getUnsafe(entity, cacheKey, keepRLock)
		if exists {
			return entry.object, true
		}
	}
	return nil, false
}

func (s *Service) DelUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) {
	if s.cache != nil && entity != nil && key != nil {
		cacheKey := formatCacheKeyStr(entity, key, "__ALL__", keyArgs...)
		s.delUnsafe(entity, cacheKey)
	}
}

func (s *Service) setUnsafe(entity *_meta.Entity, cacheKey string, entry *CacheEntry) bool {
	if entity != nil && entry != nil {
		//// Найдем структуру по ключу, если указатель отличается от нашего, то поставить признак Invalid и сбросить указатель
		//if cacheVal, exists := s.cache.Get(cacheKey); exists {
		//	if cacheEntry, ok := cacheVal.(*CacheEntry); ok {
		//		if cacheEntry.object.Value != entry.object.Value {
		//			_log.Info("RESET cache entry - set object as Invalid: entityName, key, cacheKey, cacheEntry.Value, new.Value", entity.Name, cacheEntry.key.Name, cacheKey, cacheEntry.object.Value, entry.object.Value)
		//			cacheEntry.object.SetCacheInvalid(true)
		//			cacheEntry.clear()
		//		}
		//	} else {
		//		_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cacheEntry, ok := cacheVal.(CacheEntry); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
		//	}
		//}

		return s.cache.Set(cacheKey, entry, 0)
	}
	return false
}

func (s *Service) getUnsafe(entity *_meta.Entity, cacheKey string, keepRLock bool) (*CacheEntry, bool) {
	if entity != nil && cacheKey != "" {
		// Найдем структуру по ключу, проверим признак валидности Cache
		if cacheVal, exists := s.cache.Get(cacheKey); exists {
			if cacheEntry, ok := cacheVal.(*CacheEntry); ok {
				if cacheEntry.object != nil {
					return cacheEntry, true
					////cacheEntry.object.RLock() // один объект может кешироваться по нескольким ключам
					//if !cacheEntry.object.CacheInvalid() {
					//    if keepRLock {
					//        // Оставить блокировку на чтение - вызывающая функция должна сама решить, когда снять блокировку
					//        return cacheEntry, true
					//    } else {
					//        //cacheEntry.object.RUnlock()
					//        return cacheEntry, true
					//    }
					//} else {
					//    _log.Info("GET cache entry - entry is Invalid: entityName, key, cacheKey", entity.Name, cacheEntry.key.Name, cacheKey)
					//    //cacheEntry.object.RUnlock()
					//    cacheEntry.clear()
					//    s.cache.Del(cacheKey)
					//}
				} else {
					_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cacheEntry.object != nil {}", []interface{}{cacheEntry}).PrintfError()
				}
			} else {
				_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntry, ok := cacheVal.(CacheEntry); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
			}
		}
	}
	return nil, false
}

func (s *Service) delUnsafe(entity *_meta.Entity, cacheKey string) {
	if entity != nil && cacheKey != "" {
		// Найдем структуру по ключу, проверим признак валидности Cache
		if cacheVal, exists := s.cache.Get(cacheKey); exists {
			if cacheEntry, ok := cacheVal.(*CacheEntry); ok {
				if cacheEntry.object != nil {
					_log.Info("DELETE cache entry - entry as Invalid: entityName, key, cacheKey", entity.Name, cacheEntry.key.Name, cacheKey)
					cacheEntry.object.SetCacheInvalidUnsafe(true)
					cacheEntry.clear()
					s.cache.Del(cacheKey)
				} else {
					_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntry.object != nil {}", []interface{}{cacheEntry}).PrintfError()
				}
			} else {
				_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntry, ok := cacheVal.(CacheEntry); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
			}
		}
	}
}
