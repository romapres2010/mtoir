package apiservice

import (
	"context"
	"fmt"
	"reflect"
	"time"

	_ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// SelectMarshal извлечь данные и преобразовать в JSON / XML / YAML / XLS
func (s *Service) SelectMarshal(ctx context.Context, entityName string, inFormat string, queryOptions _meta.QueryOptions) (exists bool, outBuf []byte, outFormat string, err error, errors _err.Errors) {
	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context

	if s != nil && entityName != "" {

		var localCtx = contextWithOptionsCache(ctx) // Создадим новый контекст и встроим в него OptionsCache
		var entity *_meta.Entity
		var options *_meta.Options
		var outObject *_meta.Object

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, entityName", requestID, entityName)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Meta может меняться в online - повесим запрет на чтение
		s.metaRLock()
		defer s.metaRUnLock()

		if entity = s.getEntityUnsafe(entityName); entity == nil {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)), errors
		}

		if options, err = s.parseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
			return false, nil, inFormat, err, errors
		}

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)), errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if exists, outObject, err, localErrors = s.Select(localCtx, requestID, entity, options, options.CascadeUp, options.CascadeDown, nil); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Найти в globalCache или считать из внешнего сервиса

		// сформируем ответ, возможно с уже встроенной ошибкой
		if outObject != nil {
			var errInner error
			outBuf, errInner = s.marshal(requestID, outObject.Value, "Select", entityName, options.Global.OutFormat)
			if errInner != nil {
				_log.Debug("ERROR - marshal: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				errors.Append(requestID, errInner)
				return false, nil, inFormat, errInner, errors
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				return exists, outBuf, options.Global.OutFormat, nil, errors
			}
		} else {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
			return false, nil, options.Global.OutFormat, err, errors
		}
	}
	return false, nil, inFormat, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entityName != '' {}", []interface{}{s, entityName}).PrintfError(), errors
}

// Select извлечь данные в slice - только запрошенные поля
func (s *Service) Select(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, cascadeUp int, cascadeDown int, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowsOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && s.storageMap != nil && entity != nil && options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				rowsOut = nil
				err = _recover.GetRecoverError(r, requestID, "Select", entity.Name)
			}
		}()

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)), errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			exists, rowsOut, err, localErrors = s.selectUnsafe(ctx, requestID, entity, options, cascadeUp, cascadeDown, nil)
			if err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Найти в globalCache или считать из внешнего сервиса

		// Постобработка - каскадная обработка, вычисления, валидация
		if exists && !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processPost(ctx, requestID, rowsOut, PERSIST_ACTION_GET, _meta.EXPR_ACTION_POST_GET, true, true, true, false, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		{ // Если в опциях есть динамическая POST фильтрация, то отработаем ее
			if exists && !errors.HasError() && !innerErrors.HasError() {
				if expr := rowsOut.Options.FilterPostExpr; expr != nil {
					// Только выражения, которые применимы к типу запроса
					if expr.CheckAction(_meta.EXPR_ACTION_POST_GET) {
						if rowsOutFiltering, err := s.processRowsFiltering(requestID, expr, rowsOut); err != nil {
							errors.Append(requestID, err)
						} else {
							// Переопределим после фильтрации
							if rowsOutFiltering != nil { // Фильтрация может удалить все строки
								rowsOut = rowsOutFiltering
							} else {
								exists = false
								rowsOut = nil
							}
						}
					}
				}
			}
		} // Если в опциях есть динамическая POST фильтрация, то отработаем ее

		_metrics.IncMetaCountVec("Select", entity.Name)
		_metrics.AddMetaDurationVec("Select", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			if key != nil {
				return false, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' Key '%s'['%s']=['%s'] - error 'Select'", entity.Name, key.Name, key.FieldsString(), _meta.ArgsToString("','", keyArgs...))), errors
			} else {
				return false, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - error 'Select'", entity.Name)), errors
			}
		} else {
			return exists, rowsOut, nil, errors
		}
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError(), errors
}

// selectUnsafe извлечь данные в slice - только запрошенные поля
func (s *Service) selectUnsafe(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, cascadeUp int, cascadeDown int, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowsOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && entity != nil && options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)), errors
		}

		{ // Извлечь из globalCache или внешнего источника
			var rowsResult *_meta.Object

			//if useCache {
			//	// TODO - отказаться от использования кэш для массовой выборки?
			//	if rowsResult, err = s.selectCache(ctx, requestID, txId, entity, outFields, queryOptions, keepRLock); err != nil {
			//		errors.Append(requestID, err)
			//	} else {
			//		if rowsResult != nil {
			//
			//			// Сформировать выходную структуру
			//			if rowsOut, err = s.newSliceRestrict(requestID, entity, outFields); err != nil {
			//				errors.Append(requestID, err)
			//			} else {
			//				// Скопируем только те поля, которые нужные в ответе
			//				_log.Debug("Deep CopyField Struct Slice: entityName", entity.Name)
			//				if err = entity.CopyObjectSlice(rowsResult, rowsOut, outFields); err != nil {
			//					errors.Append(requestID, err)
			//				}
			//			}
			//		}
			//	}
			//} else {
			if exists, rowsResult, err = s.selectExternal(ctx, requestID, entity, options, key, keyArgs...); err != nil {
				errors.Append(requestID, err)
			} else {
				rowsOut = rowsResult // копировать структуры не нужно, так как возвращаются только нужные поля
			}
			//}
		} // Извлечь из globalCache или внешнего источника

		// После извлечения - вычисления, валидация, без каскадной обработки
		if exists && !errors.HasError() && !innerErrors.HasError() {
			for _, row := range rowsOut.Objects {
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processGet(ctx, requestID, nil, row, PERSIST_ACTION_GET, _meta.EXPR_ACTION_POST_FETCH, 0, 0, false, true); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}
		}

		{ // Если в опциях есть динамическая PRE фильтрация, то отработаем ее
			if exists && !errors.HasError() && !innerErrors.HasError() {
				if expr := rowsOut.Options.FilterPreExpr; expr != nil {
					// Только выражения, которые применимы к типу запроса
					if expr.CheckAction(_meta.EXPR_ACTION_INSIDE_GET) {
						if rowsOutFiltering, err := s.processRowsFiltering(requestID, expr, rowsOut); err != nil {
							errors.Append(requestID, err)
						} else {
							// Переопределим после фильтрации
							if rowsOutFiltering != nil { // Фильтрация может удалить все строки
								rowsOut = rowsOutFiltering
							} else {
								exists = false // данных не найдено
								rowsOut = nil
							}
						}
					}
				}
			}
		} // Если в опциях есть динамическая PRE фильтрация, то отработаем ее

		// Постобработка - каскадная обработка, вычисления, валидация
		if exists && !errors.HasError() && !innerErrors.HasError() {
			for _, rowOut := range rowsOut.Objects {
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processGet(ctx, requestID, nil, rowOut, PERSIST_ACTION_GET, _meta.EXPR_ACTION_INSIDE_GET, cascadeUp, cascadeDown, rowOut.Options.Global.Validate, true); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}
		}

		if errors.HasError() {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return false, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - error 'Select'", entity.Name)), errors
		} else {
			_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return exists, rowsOut, nil, errors
		}
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError(), errors
}

// selectCache извлечь данные в slice - все поля
func (s *Service) selectCache(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowsOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && options != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Запрашиваем все поля для наполнения globalCache
		_log.Debug("USE globalCache - get ALL fields from external service: entityName", entity.Name)
		if exists, rowsOut, err = s.selectExternal(ctx, requestID, entity, options, key, keyArgs...); err != nil {
			return false, nil, err
		}

		//// Добавим считанные данные в globalCache
		//rowsOutValue := reflect.Indirect(rowsOut.Value)
		//for i := 0; i < rowsOutValue.Len(); i++ {
		//	rowOutPtrValue := rowsOutValue.Index(i).Addr() // указатель на текущую структуру
		//
		//	rowOut := &_meta.Object{
		//		//Key:        rowsIn.Key,
		//		StructType: rowsOut.StructType,
		//		fields:     rowsOut.fields,
		//		Val:        rowOutPtrValue.Interface(),
		//		Value:      rowOutPtrValue,
		//	}
		//
		//	// Заблокируем объект - теперь его можно использовать на чтение
		//	rowOut.RLock()
		//
		//	if err = s.cacheSetRowUnsafe(entity, nil, rowOut); err != nil { // кэшируем все кличи
		//		rowOut.RUnlock()
		//		return nil, err
		//	}
		//
		//	// Не держать блокировку текущего объекта
		//	if !keepRLock {
		//		rowOut.RUnlock()
		//	}
		//}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return exists, rowsOut, nil
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError()
}

// selectExternal извлечь данные в slice - только запрошенные поля
func (s *Service) selectExternal(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowsOut *_meta.Object, err error) {
	if s != nil && entity != nil && options != nil {

		tic := time.Now()
		txId := options.Global.TxExternal // по умолчанию работаем в глобальной транзакции

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		storage, err := s.getStorageByEntity(entity)
		if err != nil {
			return false, nil, err
		}

		// Если на указана глобальная транзакция, то ищем локальную из контекста
		if txId == 0 {
			txId = fromContextTxId(ctx)
		}

		if rowsOut, err = s.newSliceRestrict(requestID, entity, options, 0, 64); err != nil {
			return false, nil, err
		}

		// Запрос во внешний сервис - массив срок с пагинацией
		exists, err = storage.Select(ctx, requestID, txId, rowsOut, key, keyArgs...)
		if err != nil {
			return false, nil, err
		}

		if exists {
			if rowsOut != nil {
				rowsOutRV := reflect.Indirect(rowsOut.RV)
				rowsOutRVLen := rowsOutRV.Len()
				for i := 0; i < rowsOutRVLen; i++ {

					rowOutPtrRV := rowsOutRV.Index(i) // указатель на текущую структуру
					rowOut := rowsOut.NewFromRV(rowOutPtrRV, false)
					rowsOut.AppendObject(rowOut)

				}
			} else {
				// ситуация крайне странная
				return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowsOut != nil {}", []interface{}{rowsOut}).PrintfError()
			}
		}

		// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
		if err = rowsOut.SetXmlNameValueFromTagSlice(); err != nil {
			return false, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR set XMLName", entity.Name)).PrintfError()
		}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return exists, rowsOut, nil
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError()
}
