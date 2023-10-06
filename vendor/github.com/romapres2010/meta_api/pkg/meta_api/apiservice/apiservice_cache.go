package apiservice

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"

	apicacheservice "github.com/romapres2010/meta_api/pkg/meta_api/cacheristretto"
	_cache "github.com/romapres2010/meta_api/pkg/meta_api/cacheservice"
)

var defLocalCacheCfg = &apicacheservice.Config{
	NumCounters:        10000,
	MaxCost:            100000,
	BufferItems:        64,
	Metrics:            false,
	IgnoreInternalCost: false,
}

func (s *Service) newLocalCache(ctx context.Context, cfg *apicacheservice.Config) (cache _cache.CacheService, err error) {
	if s != nil {
		if cfg == nil {
			cfg = defLocalCacheCfg
		}
		return apicacheservice.New(ctx, nil, cfg)
	}
	return nil, nil
}

//func (s *Service) cacheLock(entity *_meta.Entity) (err error) {
//	if s != nil && s.globalCache != nil {
//		s.globalCache.Lock(entity)
//	}
//	return nil // не используется globalCache
//}
//
//func (s *Service) cacheUnlock(entity *_meta.Entity) {
//	if s != nil && s.globalCache != nil {
//		s.globalCache.Unlock(entity)
//	}
//}
//
//func (s *Service) cacheRLock(entity *_meta.Entity) (err error) {
//	if s != nil && s.globalCache != nil {
//		s.globalCache.RLock(entity)
//	}
//	return nil // не используется globalCache
//}
//
//func (s *Service) cacheRUnlock(entity *_meta.Entity) {
//	if s != nil && s.globalCache != nil {
//		s.globalCache.RUnlock(entity)
//	}
//}

//func (s *Service) cacheSetRowUnsafePtr(entity *_meta.Entity, key *_meta.Key, rowPtr interface{}) (err error) {
//	if s != nil && s.globalCache != nil {
//		if entity != nil && rowPtr != nil {
//
//			if key == nil {
//				// Если key пустой, кешируем на все ключи, которые есть в сущности
//				for _, k := range entity.Keys {
//					if err = s.cacheSetRowUnsafePtrByKey(entity, k, rowPtr); err != nil {
//						return err
//					}
//				}
//			} else {
//				// Кешируем конкретный ключ
//				return s.cacheSetRowUnsafePtrByKey(entity, key, rowPtr)
//			}
//			return nil
//		}
//		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && rowPtr != nil {}", []interface{}{entity, rowPtr, key}).PrintfError()
//	}
//	return nil // не используется globalCache
//}

func (s *Service) cacheSetRowUnsafe(cache _cache.CacheService, key *_meta.Key, row *_meta.Object) (err error) {
	if s != nil {

		if cache == nil {
			cache = s.globalCache
		}

		if row != nil && row.Entity != nil {

			if key == nil {
				// Если key пустой, кешируем на все ключи, которые есть в сущности
				for _, k := range row.Entity.KeysUK() {
					if err = s.cacheSetRowUnsafeByKey(cache, k, row); err != nil {
						return err
					}
				}
			} else {
				// Кешируем конкретный ключ
				return s.cacheSetRowUnsafeByKey(cache, key, row)
			}
			return nil
		}
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if row != nil && row.Entity != nil {}", []interface{}{row, key}).PrintfError()
	}
	return nil // не используется globalCache
}

//func (s *Service) cacheSetRowUnsafePtrByKey(entity *_meta.Entity, key *_meta.Key, rowPtr interface{}) (err error) {
//	if s != nil && s.globalCache != nil {
//		if entity != nil && rowPtr != nil && key != nil {
//
//			// Функция восстановления после паники
//			defer func() {
//				r := recover()
//				if r != nil {
//					err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "cacheSetRowUnsafePtrByKey", entity.Name)
//				}
//			}()
//
//			keyArgs, err := entity.KeyFieldsValuePtr(key, reflect.ValueOf(rowPtr))
//			if err != nil {
//				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s' - error get fields value", entity.Name, key.Name))
//			}
//
//			if len(keyArgs) > 0 {
//				cacheHit := s.globalCache.SetPtrUnsafe(entity, key, rowPtr, keyArgs...)
//				if !cacheHit {
//					_log.Info("DO NOT add Entity in globalCache - globalCache rejected: entityName, keyArgs", entity.Name, keyArgs)
//				}
//			}
//
//			return nil
//		}
//		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && rowPtr != nil && key != nil {}", []interface{}{entity, rowPtr, key}).PrintfError()
//	}
//	return nil // не используется globalCache
//}

func (s *Service) cacheSetRowUnsafeByKey(cache _cache.CacheService, key *_meta.Key, row *_meta.Object) (err error) {
	if s != nil {

		if cache == nil {
			cache = s.globalCache
		}

		if row != nil && row.Entity != nil && key != nil {

			// Функция восстановления после паники
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "cacheSetRowUnsafeByKey", row.Entity.Name)
				}
			}()

			keyArgs, err := row.KeyFieldsValue(key)
			if err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s' - error get fields value", row.Entity.Name, key.Name))
			}

			if len(keyArgs) > 0 {
				cacheHit := cache.Set(key, row, keyArgs...)
				if !cacheHit {
					_log.Info("DO NOT add Entity in globalCache - globalCache rejected: entityName, keyArgs", row.Entity.Name, keyArgs)
				}
			}

			return nil
		}
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if row!= nil && row.Entity  != nil && key != nil {}", []interface{}{row, key}).PrintfError()
	}
	return nil // не используется globalCache
}

func (s *Service) cacheGetRowUnsafeByKey(cache _cache.CacheService, entity *_meta.Entity, key *_meta.Key, keepRLock bool, keyArgs ...interface{}) (cacheHit bool, rowOut *_meta.Object, err error) {
	if s != nil {

		if cache == nil {
			cache = s.globalCache
		}

		if entity != nil && key != nil && len(keyArgs) > 0 {

			// Функция восстановления после паники
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "cacheGetRowUnsafeByKey", entity.Name)
				}
			}()

			if rowOut, cacheHit = cache.Get(entity, key, keepRLock, keyArgs...); cacheHit {
				_log.Debug("USE globalCache - Found Entity value in globalCache: entityName, keyName, keyArgs", entity.Name, key.Name, keyArgs)

				return cacheHit, rowOut, nil
			}

			return false, nil, nil
		}
		return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && key != nil {}", []interface{}{entity, key}).PrintfError()
	}
	return false, nil, nil // не используется globalCache
}

//func (s *Service) cacheDelRowUnsafe(entity *_meta.Entity, key *_meta.Key, rowPtr interface{}) (err error) {
//	if s != nil && s.globalCache != nil {
//		if entity != nil && rowPtr != nil {
//
//			if key == nil {
//				// Если key пустой, кешируем на все ключи, которые есть в сущности
//				for _, k := range entity.Keys {
//					if err = s.cacheDelUnsafeRowByKey(entity, k, rowPtr); err != nil {
//						return err
//					}
//				}
//			} else {
//				// Кешируем конкретный ключ
//				return s.cacheDelUnsafeRowByKey(entity, key, rowPtr)
//			}
//			return nil
//		}
//		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && rowPtr != nil {}", []interface{}{entity, rowPtr, key}).PrintfError()
//	}
//	return nil // не используется globalCache
//}

//func (s *Service) cacheDelUnsafeRowByKey(entity *_meta.Entity, key *_meta.Key, rowPtr interface{}) (err error) {
//	if s != nil && s.globalCache != nil {
//		if entity != nil && rowPtr != nil && key != nil {
//
//			// Функция восстановления после паники
//			defer func() {
//				r := recover()
//				if r != nil {
//					err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "cacheSetRowUnsafePtrByKey", entity.Name)
//				}
//			}()
//
//			keyArgs, err := entity.KeyFieldsValuePtr(key, reflect.ValueOf(rowPtr))
//			if err != nil {
//				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s' - error get fields value", entity.Name, key.Name))
//			}
//
//			if len(keyArgs) > 0 {
//				s.globalCache.DelPtr(entity, key, keyArgs...)
//			}
//
//			return nil
//		}
//		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && rowPtr != nil && key != nil {}", []interface{}{entity, rowPtr, key}).PrintfError()
//	}
//	return nil // не используется globalCache
//}

//func (s *Service) cacheSetSlice(entity *_meta.Entity, key *_meta.Key, slicePtr interface{}) (err error) {
//	if s != nil && s.globalCache != nil {
//		if entity != nil && slicePtr != nil {
//			// Функция восстановления после паники
//			defer func() {
//				r := recover()
//				if r != nil {
//					err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "cacheSetSlice", entity.Name)
//				}
//			}()
//
//			// Добавим считанные данные в globalCache и переформируем выходную структуру
//			sliceValue := reflect.Indirect(reflect.ValueOf(slicePtr)) // Собственно sliceValue с данными
//			for i := 0; i < sliceValue.Len(); i++ {
//				rowPtr := sliceValue.Index(i).Addr().Interface() // текущий объект
//				if err = s.cacheSetRowUnsafePtr(entity, key, rowPtr); err != nil {
//					return err
//				}
//			}
//			return nil
//		}
//		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && slicePtr != nil {}", []interface{}{entity, slicePtr}).PrintfError()
//	}
//	return nil // не используется globalCache
//}

func (s *Service) PopulateEntityCache(ctx context.Context, entity *_meta.Entity) (err error) {
	if s != nil && s.globalCache != nil && entity != nil {
		var tic = time.Now()

		//_, _, err = s.selectCache(ctx, 0, entity, nil, nil, nil, nil)
		//if err != nil {
		//	return err
		//}

		_log.Info("SUCCESS: entityName, duration", entity.Name, time.Now().Sub(tic))
	}
	return nil
}

func (s *Service) PopulateAllEntityCache(ctx context.Context) (err error) {
	if s != nil && s.globalCache != nil {
		var tic = time.Now()

		for _, entity := range s.meta.Entities {
			if !entity.SkipCache {
				err = s.PopulateEntityCache(ctx, entity)
				if err != nil {
					return err
				}
			}
		}
		_log.Info("SUCCESS: duration", time.Now().Sub(tic))
	}
	return nil
}
