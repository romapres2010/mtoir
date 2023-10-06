package apiservice

import (
	"context"
	"fmt"
	"time"

	_ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
)

// GetMarshal извлечь данные и преобразовать в JSON / XML / YAML / XLS
func (s *Service) GetMarshal(ctx context.Context, entityName string, inFormat string, queryOptions _meta.QueryOptions, keyArgs ...interface{}) (exists bool, outBuf []byte, outFormat string, err error, errors _err.Errors) {
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

		if entity = s.GetEntityUnsafe(entityName); entity == nil {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)), errors
		}

		if options, err = s.ParseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
			return false, nil, inFormat, err, errors
		}
		options.Global.InFormat = inFormat

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)), errors
		}

		if entity.PKKey() == nil {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Key with 'key.type'='PK' was not defined", entity.Name)), errors
		}

		if len(entity.PKKey().Fields()) != len(keyArgs) {
			return false, nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Key '%s' with 'key.type'='PK' - incorrect number of arguments, expected '%v' recieved '%v'", entity.Name, entity.PKKey().Name, len(entity.PKKey().Fields()), len(keyArgs))), errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if exists, outObject, err, localErrors = s.GetSingle(localCtx, requestID, entity, options, options.CascadeUp, options.CascadeDown, entity.PKKey(), keyArgs...); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Найти в globalCache или считать из внешнего сервиса

		// сформируем ответ, возможно с уже встроенной ошибкой
		if outObject != nil {
			var errInner error
			outBuf, errInner = s.MarshalEntity(requestID, outObject.Value, "Get", entityName, options.Global.OutFormat)
			if errInner != nil {
				_log.Debug("ERROR - MarshalEntity: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
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

// GetSingle извлечь данные в struct - только запрошенные поля
func (s *Service) GetSingle(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, cascadeUp int, cascadeDown int, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && entity != nil && options != nil && key != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, entityName, keyName, keyArgs", requestID, entity.Name, key.Name, keyArgs)

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

		// Проверить если все аргументы пустые - то сразу отказ
		if _meta.ArgsAllEmpty(keyArgs) {
			return false, nil, nil, errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			exists, rowOut, err, localErrors = s.GetSingleUnsafe(ctx, requestID, entity, options, cascadeUp, cascadeDown, key, keyArgs...)
			if err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Найти в globalCache или считать из внешнего сервиса

		// Постобработка - каскадная обработка, вычисления, валидация
		if exists && !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processPost(ctx, requestID, rowOut, PERSIST_ACTION_GET, _meta.EXPR_ACTION_POST_GET, true, true, true, false, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		_metrics.IncMetaCountVec("Get", entity.Name)
		_metrics.AddMetaDurationVec("Get", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			if key != nil {
				return false, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' Key '%s'['%s']=['%s'] - error 'Get'", entity.Name, key.Name, key.FieldsString(), _meta.ArgsToString("','", keyArgs...))), errors
			} else {
				return false, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - error 'Get'", entity.Name)), errors
			}
		} else {
			return exists, rowOut, nil, errors
		}
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && options != nil  && key != nil {}", []interface{}{s, entity, options, key}).PrintfError(), errors
}

// GetSingleUnsafe извлечь данные в struct - только запрошенные поля
func (s *Service) GetSingleUnsafe(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, cascadeUp int, cascadeDown int, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && entity != nil && options != nil && key != nil {

		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, entityName, keyName, keyArgs", requestID, entity.Name, key.Name, keyArgs)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)).PrintfError(), errors
		}

		// Проверить если все аргументы пустые - то сразу отказ
		if _meta.ArgsAllEmpty(keyArgs) {
			return false, nil, nil, errors
		}

		{ // Извлечь из globalCache или внешнего источника
			var rowResult *_meta.Object

			if options.Global.UseCache {
				// Найти в globalCache или считать из внешнего сервиса все поля
				// Найденная запись будет заблокирована на чтение до завершения текущей операции
				exists, rowResult, err = s.getCache(ctx, requestID, entity, options, key, keyArgs...)
				if err != nil {
					errors.Append(requestID, err)
				} else {
					if exists && rowResult != nil {
						//if options.ArgsFields != nil {
						// TODO - копировать при извлечении нужно всегда, чтобы не менялись данные в globalCache

						// Сформировать выходную структуру
						if rowOut, err = s.newRowRestrict(requestID, entity, options); err != nil {
							errors.Append(requestID, err)
						} else {
							// Скопируем только те поля, которые нужные в ответе
							_log.Debug("Deep CopyField Struct: entityName", entity.Name)
							if err = entity.CopyObjectStruct(rowResult, rowOut, options.Fields); err != nil {
								errors.Append(requestID, err)
							}
						}
						//} else {
						//	rowOut = rowResult // копировать структуры не нужно
						//}
					}
				}
			} else {
				exists, rowResult, err = s.getExternal(ctx, requestID, entity, options, key, keyArgs...)
				if err != nil {
					errors.Append(requestID, err)
				} else {
					if exists && rowResult != nil {
						rowOut = rowResult // копировать структуры не нужно, так как возвращаются только нужные поля
					}
				}
			}
		} // Извлечь из globalCache или внешнего источника

		// После извлечения - вычисления, валидация, без каскадной обработки
		if exists && !errors.HasError() && !innerErrors.HasError() {
			// Пост обработка только для полей, которые нужно возвращать
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processGet(ctx, requestID, nil, rowOut, PERSIST_ACTION_GET, _meta.EXPR_ACTION_POST_FETCH, 0, 0, &processOptions{validate: false, calculate: true, addCrossRef: false}); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Постобработка - каскадная обработка, вычисления, валидация
		if exists && !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processGet(ctx, requestID, nil, rowOut, PERSIST_ACTION_GET, _meta.EXPR_ACTION_INSIDE_GET, cascadeUp, cascadeDown, &processOptions{validate: rowOut.Options.Global.Validate, calculate: true, addCrossRef: false}); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		if errors.HasError() {
			err = s.processErrors(requestID, rowOut, errors, options.Global.EmbedError, "Get")
			return false, nil, err, errors
		} else {
			return exists, rowOut, nil, errors
		}

	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && options != nil  && key != nil {}", []interface{}{s, entity, options, key}).PrintfError(), errors
}

// GetSingleByKeyUnsafe считать поля ключа
func (s *Service) GetSingleByKeyUnsafe(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, rowIn *_meta.Object, key *_meta.Key) (exists bool, rowOut *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && entity != nil && rowIn != nil && key != nil && options != nil {

		innerErrors := _err.Errors{} // Ошибки вложенных методов

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Разрешено ли запрашивать
		if entity.Modify.RetrieveRestrict {
			return false, nil, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to retrieve", entity.Name)).PrintfError(), errors
		}

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		keyArgs, err = rowIn.KeyFieldsValue(key)
		if err != nil {
			errors.Append(requestID, err)
		} else {
			{ // Найти в globalCache или считать из внешнего сервиса, возврат только полей ключа
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				exists, rowOut, err, localErrors = s.GetSingleUnsafe(ctx, requestID, entity, options, 0, 0, key, keyArgs...)
				if err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // Найти в globalCache или считать из внешнего сервиса, возврат только полей ключа
		}

		if errors.HasError() {
			return false, nil, keyArgs, err, errors
		} else {
			return exists, rowOut, keyArgs, nil, errors
		}

	}
	return false, nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && rowIn != nil  && key != nil && options != nil {}", []interface{}{s, entity, rowIn, key}).PrintfError(), errors
}

// getCache извлечь данные в struct - все поля
func (s *Service) getCache(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && options != nil {

		tic := time.Now()
		cacheHit := false

		_log.Debug("START: requestID, entityName, keyName, keyArgs", requestID, entity.Name, key.Name, keyArgs)

		// Ищем в globalCache - там содержится полный набор полей
		if cacheHit, rowOut, err = s.cacheGetRowUnsafeByKey(nil, entity, key, options.Global.KeepLock, keyArgs...); cacheHit {
			//_log.Info("START - found in globalCache: requestID, entityName, keyName, keyArgs, object.Value", requestID, entity.Name, key.Name, keyArgs, &rowOut.Value)
			return true, rowOut, nil
			//} else {
			//	_log.Info("START - NOT FOUND in globalCache: requestID, entityName, keyName, keyArgs", requestID, entity.Name, key.Name, keyArgs)
		}

		// Запрашиваем все поля для наполнения globalCache
		_log.Debug("USE globalCache - get ALL fields from external service: entityName", entity.Name)
		if exists, rowOut, err = s.getExternal(ctx, requestID, entity, options, key, keyArgs...); err != nil {
			return false, nil, err
		}

		// Добавим считанные данные в globalCache
		if exists && rowOut != nil {

			// TODO - блокировки перенести на уровень выше
			//// Заблокируем объект - теперь его можно использовать на чтение
			//rowOut.RLock()

			// TODO - ответ из функции происходит быстрее, чем реально помещается в globalCache
			if err = s.cacheSetRowUnsafe(nil, nil, rowOut); err != nil { // кэшируем все кличи
				//rowOut.RUnlock()
				return false, nil, err
				//} else {
				//	_log.Info("START - ADD to globalCache: requestID, entityName, keyName, keyArgs, object.PtrValue", requestID, entity.Name, key.Name, keyArgs, &rowOut.Value)
			}

			//// Не держать блокировку текущего объекта
			//if !keepRLock {
			//	rowOut.RUnlock()
			//}
		}

		_log.Debug("SUCCESS: requestID, entityName, key, keyArgs, exists, duration", requestID, entity.Name, key.Name, keyArgs, exists, time.Now().Sub(tic))

		return exists, rowOut, nil
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError()
}

// getExternal извлечь данные в struct - только запрошенные поля
func (s *Service) getExternal(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, key *_meta.Key, keyArgs ...interface{}) (exists bool, rowOut *_meta.Object, err error) {
	if s != nil && entity != nil && options != nil {

		tic := time.Now()
		txId := options.Global.TxExternal // по умолчанию работаем в глобальной транзакции

		_log.Debug("START: requestID, entityName, keyName, keyArgs", requestID, entity.Name, key.Name, keyArgs)

		storage, err := s.getStorageByEntity(entity)
		if err != nil {
			return false, nil, err
		}

		// Если на указана глобальная транзакция, то ищем локальную из контекста
		if txId == 0 {
			txId = fromContextTxId(ctx)
		}

		if rowOut, err = s.newRowRestrict(requestID, entity, options); err != nil {
			return false, nil, err
		}

		// Запрос во внешний сервис
		exists, err = storage.Get(ctx, requestID, txId, rowOut, key, keyArgs...)
		if err != nil {
			return false, nil, err
		}

		// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
		if err = rowOut.SetXmlNameValueFromTag(); err != nil {
			return false, nil, err
		}

		_log.Debug("SUCCESS: requestID, entityName, key, keyArgs, exists, duration", requestID, entity.Name, key.Name, keyArgs, exists, time.Now().Sub(tic))

		return exists, rowOut, nil
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && options != nil {}", []interface{}{s, entity, options}).PrintfError()
}
