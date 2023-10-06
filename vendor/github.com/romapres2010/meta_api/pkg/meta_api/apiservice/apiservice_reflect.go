package apiservice

import (
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

func (s *Service) newSliceAny(requestID uint64, entity *_meta.Entity, options *_meta.Options, fieldsMap _meta.FieldsMap, typeCacheKey string, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {

		cacheHit := false

		// Создать новый объект - возврат указателя и признак был ли тип в globalCache
		if cacheHit, slice, err = entity.NewSliceAny(fieldsMap, typeCacheKey, len, cap); err != nil {
			return nil, err
		}
		slice.Options = options

		// регистрируем правило на созданный тип структуры если его еще не было в globalCache
		if !cacheHit {
			rules := entity.ValidationRules()
			s.validator.RegisterValidationRules(rules, slice.Value)
			_log.Debug("RegisterValidationRules: externalId, entity.Name, entity.Type, rules", requestID, entity.Name, slice.StructType.String(), rules)
		}

		return slice, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newSlice(requestID uint64, entity *_meta.Entity, options *_meta.Options, fieldsMap _meta.FieldsMap, typeCacheKey string, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {

		cacheHit := false

		// Создать новый объект - возврат указателя и признак был ли тип в globalCache
		if cacheHit, slice, err = entity.NewSlice(fieldsMap, typeCacheKey, len, cap); err != nil {
			return nil, err
		}
		slice.Options = options

		// регистрируем правило на созданный тип структуры если его еще не было в globalCache
		if !cacheHit {
			rules := entity.ValidationRules()
			s.validator.RegisterValidationRules(rules, slice.Value)
			_log.Debug("RegisterValidationRules: externalId, entity.Name, entity.Type, rules", requestID, entity.Name, slice.StructType.String(), rules)
		}

		return slice, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newSliceAnyAll(requestID uint64, entity *_meta.Entity, options *_meta.Options, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {
		return s.newSliceAny(requestID, entity, options, nil, entity.DefTypeCacheKey(), len, cap)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) NewSliceAll(requestID uint64, entity *_meta.Entity, options *_meta.Options, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {
		return s.newSlice(requestID, entity, options, nil, entity.DefTypeCacheKey(), len, cap)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newSliceAnyRestrict(requestID uint64, entity *_meta.Entity, options *_meta.Options, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {

		if options.Fields != nil {
			entity.AddPkFieldsToMap(&options.Fields)     // Ключевые поля PK всегда выводим
			entity.AddSystemFieldsToMap(&options.Fields) // Добавим системные поля
		}

		typeCacheKey := entity.TypeCacheKey(_meta.TYPE_CACHE_KEY_PREFIX_ALL, options.Fields)

		slice, err = s.newSliceAny(requestID, entity, options, options.Fields, typeCacheKey, len, cap)
		if err != nil {
			return nil, err
		}

		return slice, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newSliceRestrict(requestID uint64, entity *_meta.Entity, options *_meta.Options, len, cap int) (slice *_meta.Object, err error) {
	if s != nil && entity != nil {

		if options.Fields != nil {
			entity.AddPkFieldsToMap(&options.Fields)     // Ключевые поля PK всегда выводим
			entity.AddSystemFieldsToMap(&options.Fields) // Добавим системные поля
		}

		typeCacheKey := entity.TypeCacheKey(_meta.TYPE_CACHE_KEY_PREFIX_ALL, options.Fields)

		slice, err = s.newSlice(requestID, entity, options, options.Fields, typeCacheKey, len, cap)
		if err != nil {
			return nil, err
		}

		return slice, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newRow(requestID uint64, entity *_meta.Entity, options *_meta.Options, fieldsMap _meta.FieldsMap, typeCacheKey string) (row *_meta.Object, err error) {
	if s != nil && entity != nil {

		cacheHit := false

		// Создать новый объект - возврат указателя и признак был ли тип в globalCache
		cacheHit, row, err = entity.NewStruct(fieldsMap, typeCacheKey)
		if err != nil {
			return nil, err
		}
		row.Options = options

		// регистрируем правило на созданный тип структуры если его еще не было в globalCache
		if !cacheHit {
			rules := entity.ValidationRules()
			s.validator.RegisterValidationRules(rules, row.Value)
			_log.Debug("RegisterValidationRules: externalId, entity.Name, entity.Type, rules", requestID, entity.Name, row.StructType.String(), rules)
		}

		return row, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) NewRowAll(requestID uint64, entity *_meta.Entity, options *_meta.Options) (row *_meta.Object, err error) {
	if s != nil && entity != nil {
		return s.newRow(requestID, entity, options, nil, entity.DefTypeCacheKey())
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newRowAllEmptyRef(requestID uint64, entity *_meta.Entity, options *_meta.Options) (row *_meta.Object, err error) {
	if s != nil && entity != nil {
		return s.newRow(requestID, entity, options, nil, entity.DefTypeCacheKeyEmptyRef())
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newRowAllEmptyTag(requestID uint64, entity *_meta.Entity, options *_meta.Options) (row *_meta.Object, err error) {
	if s != nil && entity != nil {
		return s.newRow(requestID, entity, options, nil, entity.DefTypeCacheKeyEmptyTag())
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) newRowRestrict(requestID uint64, entity *_meta.Entity, options *_meta.Options) (row *_meta.Object, err error) {
	if s != nil && entity != nil {

		if options.Fields != nil {
			entity.AddPkFieldsToMap(&options.Fields)     // Ключевые поля PK всегда выводим
			entity.AddSystemFieldsToMap(&options.Fields) // Добавим системные поля
		}

		typeCacheKey := entity.TypeCacheKey(_meta.TYPE_CACHE_KEY_PREFIX_ALL, options.Fields)

		row, err = s.newRow(requestID, entity, options, options.Fields, typeCacheKey)
		if err != nil {
			return nil, err
		}

		return row, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}
