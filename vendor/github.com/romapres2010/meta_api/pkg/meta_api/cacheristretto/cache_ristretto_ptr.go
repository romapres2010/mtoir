package cacheristretto

//func (s *Service) SetPtr(entity *_meta.Entity, key *_meta.Key, val interface{}, keyArgs ...interface{}) bool {
//    if s.cache != nil && entity != nil && key != nil {
//        s.Lock(entity)
//        defer s.Unlock(entity)
//        return s.SetPtrUnsafe(entity, key, val, keyArgs...)
//    }
//    return false
//}
//
//func (s *Service) GetPtr(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) (interface{}, bool) {
//    if s.cache != nil && entity != nil && key != nil {
//        s.RLock(entity)
//        defer s.RUnlock(entity)
//        return s.GetPtrUnsafe(entity, key, keyArgs...)
//    }
//    return nil, false
//}
//
//func (s *Service) DelPtr(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) {
//    if s.cache != nil && entity != nil && key != nil {
//        s.Lock(entity)
//        defer s.Unlock(entity)
//        s.DelPtrUnsafe(entity, key, keyArgs...)
//    }
//}
//
//func (s *Service) SetPtrUnsafe(entity *_meta.Entity, key *_meta.Key, val interface{}, keyArgs ...interface{}) bool {
//    if s.cache != nil && entity != nil && key != nil {
//        cacheKey := formatCacheKeyStr(entity, key, "__ALL__", keyArgs...)
//        entry := newCacheEntryPtr(entity, key, cacheKey, val)
//
//        return s.setPtrUnsafe(entity, cacheKey, entry)
//    }
//    return false
//}
//func (s *Service) GetPtrUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) (interface{}, bool) {
//    if s.cache != nil && entity != nil && key != nil {
//        cacheKey := formatCacheKeyStr(entity, key, "__ALL__", keyArgs...)
//
//        entry, exists := s.getPtrUnsafe(entity, cacheKey)
//        if exists {
//            return entry.ptr, exists
//        }
//    }
//    return nil, false
//}
//
//func (s *Service) DelPtrUnsafe(entity *_meta.Entity, key *_meta.Key, keyArgs ...interface{}) {
//    if s.cache != nil && entity != nil && key != nil {
//        cacheKey := formatCacheKeyStr(entity, key, "__ALL__", keyArgs...)
//        s.delPtrUnsafe(entity, cacheKey)
//    }
//}
//
//func (s *Service) setPtrUnsafe(entity *_meta.Entity, cacheKey string, entry *CacheEntryPtr) bool {
//    if entity != nil && entry != nil {
//        // Найдем структуру по ключу, если указатель отличается от нашего, то поставить признак Invalid и сбросить указатель
//        if cacheVal, exists := s.cache.Get(cacheKey); exists {
//            if cacheEntry, ok := cacheVal.(*CacheEntryPtr); ok {
//                if cacheEntry.ptr != entry.ptr {
//                    _log.Debug("RESET cache entry - set others entry as Invalid: entityName, key, cacheKey", entity.Name, cacheEntry.key.Name, cacheKey)
//                    if err := entity.SetCacheInvalidValuePtr(reflect.ValueOf(cacheEntry.ptr), true); err == nil {
//                        cacheEntry.clear()
//                    } else {
//                        _ = _err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err, "if err := entity.SetCacheInvalidValuePtr(reflect.ValueOf(cacheEntry.ptr), true); err == nil {}", []interface{}{cacheEntry}).PrintfError()
//                    }
//                }
//            } else {
//                _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cacheEntry, ok := cacheVal.(CacheEntryPtr); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
//            }
//        }
//
//        return s.cache.Set(cacheKey, entry, 0)
//    }
//    return false
//}
//
//func (s *Service) getPtrUnsafe(entity *_meta.Entity, cacheKey string) (*CacheEntryPtr, bool) {
//    if entity != nil && cacheKey != "" {
//        // Найдем структуру по ключу, проверим признак валидности Cache
//        if cacheVal, exists := s.cache.Get(cacheKey); exists {
//            if cacheEntry, ok := cacheVal.(*CacheEntryPtr); ok {
//                if cacheEntry.ptr != nil {
//                    if invalid, err := entity.CacheInvalidValuePtr(reflect.ValueOf(cacheEntry.ptr)); err == nil {
//                        if !invalid {
//                            return cacheEntry, true
//                        } else { // удалить невалидную запись
//                            _log.Debug("FOUND invalid cache CacheEntryPtr - Delete it: entityName, key, keyArgs", entity.Name, cacheEntry.key.Name, cacheEntry.cacheKey)
//                            cacheEntry.clear()
//                            s.cache.Del(cacheKey)
//                        }
//                    } else {
//                        _ = _err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err, "if invalid, err := entity.CacheInvalidValuePtr(reflect.ValueOf(CacheEntryPtr.ptr)); err == nil {}", []interface{}{cacheEntry}).PrintfError()
//                    }
//                } else {
//                    _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntryPtr.ptr != nil {}", []interface{}{cacheEntry}).PrintfError()
//                }
//            } else {
//                _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntryPtr, ok := cacheVal.(CacheEntryPtr); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
//            }
//        }
//    }
//    return nil, false
//}
//
//func (s *Service) delPtrUnsafe(entity *_meta.Entity, cacheKey string) {
//    if entity != nil && cacheKey != "" {
//        // Найдем структуру по ключу, проверим признак валидности Cache
//        if cacheVal, exists := s.cache.Get(cacheKey); exists {
//            if cacheEntry, ok := cacheVal.(*CacheEntryPtr); ok {
//                if cacheEntry.ptr != nil {
//                    _log.Debug("DELETE cache entry - set others entry as Invalid: entityName, key, cacheKey", entity.Name, cacheEntry.key.Name, cacheKey)
//                    if err := entity.SetCacheInvalidValuePtr(reflect.ValueOf(cacheEntry.ptr), true); err == nil {
//                        cacheEntry.clear()
//                        s.cache.Del(cacheKey)
//                    } else {
//                        _ = _err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err, "if err := entity.SetCacheInvalidValuePtr(reflect.ValueOf(cacheEntry.ptr), true); err == nil {}", []interface{}{cacheEntry}).PrintfError()
//                    }
//                } else {
//                    _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntryPtr.ptr != nil {}", []interface{}{cacheEntry}).PrintfError()
//                }
//            } else {
//                _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if CacheEntryPtr, ok := cacheVal.(CacheEntryPtr); ok {}", []interface{}{reflect.ValueOf(cacheVal).String()}).PrintfError()
//            }
//		}
//	}
//}
