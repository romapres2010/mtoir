package apiservice

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"

	_storage "github.com/romapres2010/meta_api/pkg/meta_api/storageservice"
)

// CreateMarshal создать строку во внешнем сервисе
func (s *Service) CreateMarshal(ctx context.Context, entityName string, inBuf []byte, inFormat string, queryOptions _meta.QueryOptions) (outBuf []byte, outFormat string, err error, errors _err.Errors) {
	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context

	if s != nil && entityName != "" {

		var localCtx = contextWithOptionsCache(ctx) // Создадим новый контекст и встроим в него OptionsCache
		var entity *_meta.Entity
		var options *_meta.Options
		var inObject *_meta.Object
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

		// TODO - добавить локальный cache на время выполнения запроса -целостность чтения на уровне транзакции
		//localCache, err := s.newLocalCache(ctx, nil)
		//if err != nil {
		//	return nil, inFormat, err
		//
		//} else {
		//	defer localCache.CloseAll()
		//}

		// Meta может меняться в online - повесим запрет на чтение
		s.metaRLock()
		defer s.metaRUnLock()

		if entity = s.getEntityUnsafe(entityName); entity == nil {
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)), errors
		}

		if options, err = s.parseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
			return nil, inFormat, err, errors
		}

		// Разрешено ли вставлять
		if entity.Modify.CreateRestrict {
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
		}

		{ // TODO - переделать - для ссылочных полей добавим ограничение - считывать только ключи
			if len(options.Fields) > 0 {
				var refKeyFields []string
				// Добавить только поля ключей
				for _, field := range options.Fields {
					if ref := field.Reference(); ref != nil {
						if ref.ToKey() != nil {
							// Для каждого поля-ссылки добавить ссылочную сущность и поля этой сущности
							for _, toKeyField := range ref.ToKey().Fields() {
								refKeyFields = append(refKeyFields, ref.GetTagName(options.Global.NameFormat, true)+"."+toKeyField.GetTagName(options.Global.NameFormat, true))
							}
						}
					}
				}
				if len(refKeyFields) > 0 {
					refFieldsName := options.QueryOptionsDown[s.cfg.QueryOption.FieldsFull]
					options.QueryOptionsDown[s.cfg.QueryOption.FieldsFull] = refFieldsName + strings.Join(refKeyFields, ",")
				}
			}
		} // TODO - переделать - для ссылочных полей добавим ограничение - считывать только ключи

		if options.Global.MultiRow {
			// Распарсить многострочный запрос - список полей для JSON вернется только те, что в запросе
			if inObject, err = s.unmarshalMulti(requestID, entity, options, inBuf, inFormat, options.Global.IgnoreExtraField); err != nil {
				return nil, inFormat, err, errors
			}

			{ // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if outObject, err, localErrors = s.createMultiUnsafe(localCtx, requestID, inObject, options.CascadeUp, options.CascadeDown, options.Global.UseCache); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
		} else {
			// Распарсить однострочный запрос - список полей полный для сущности
			if inObject, err = s.unmarshalSingle(requestID, entity, options, inBuf, inFormat, options.Global.IgnoreExtraField); err != nil {
				return nil, inFormat, err, errors
			}

			{ // создать строку во внешнем сервисе
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if outObject, err, localErrors = s.createSingleUnsafe(localCtx, requestID, inObject, options.CascadeUp, options.CascadeDown, options.Global.UseCache); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // создать строку во внешнем сервисе
		}

		// сформируем ответ, возможно с уже встроенной ошибкой
		if outObject != nil {
			var errInner error
			outBuf, errInner = s.marshal(requestID, outObject.Value, "Create", entityName, options.Global.OutFormat)
			if errInner != nil {
				_log.Debug("ERROR - marshal: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				errors.Append(requestID, errInner)
				return nil, inFormat, errInner, errors
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				return outBuf, options.Global.OutFormat, err, errors
			}
		} else {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
			return nil, inFormat, err, errors
		}
	}
	return nil, inFormat, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entityName != '' {}", []interface{}{s, entityName}).PrintfError(), errors
}

// createMultiUnsafe создать строки во внешнем сервисе
func (s *Service) createMultiUnsafe(ctx context.Context, requestID uint64, rowsIn *_meta.Object, cascadeUp int, cascadeDown int, useCache bool) (rowsOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowsIn != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		entity := rowsIn.Entity

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
				err = _recover.GetRecoverError(r, requestID, "createMultiUnsafe", entity.Name)
			}
		}()

		// Обрабатываем только slice
		if !rowsIn.IsSlice {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if !rowIn.IsSlice {}", []interface{}{s, rowsIn}).PrintfError(), errors
		}

		// Разрешено ли запрашивать
		if entity.Modify.CreateRestrict {
			return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
		}

		// Собственно slice, из которого будем считывать данные
		rowsInValue := reflect.Indirect(rowsIn.RV)
		rowsInValueLen := rowsInValue.Len()

		// Временный объект, в который будем копировать элементы slice для дальнейшей обработки
		var rowInTmp *_meta.Object
		rowInTmp, err = s.newRowRestrict(requestID, entity, rowsIn.Options)
		if err != nil {
			return nil, err, errors
		}

		// Сформировать выходную структуру - указатель на slice
		if rowsOut, err = s.newSliceRestrict(requestID, entity, rowsIn.Options, 0, rowsInValueLen); err != nil {
			return nil, err, errors
		}

		// Собственно slice, в который будем вставлять
		rowsOutRV := reflect.Indirect(rowsOut.RV)

		{ // Обработаем все строки во входной структуре
			for i := 0; i < rowsInValueLen; i++ {
				rowInPtrValue := rowsInValue.Index(i) // указатель на текущую структуру
				rowInTmp.SetFromRV(rowInPtrValue)     // Инициируем временный объект

				var rowOut *_meta.Object
				{ // создать строку во внешнем сервисе
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if rowOut, err, localErrors = s.createSingleUnsafe(ctx, requestID, rowInTmp, cascadeUp, cascadeDown, false); err != nil {
						// Если ошибки встраиваем, то добавить структуру ответа в общий список
						if rowsIn.Options.Global.EmbedError && rowOut != nil {
							rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV)) // добавляем в slice структуру из указателя - ошибки встроены
						} else {
							errors.Append(requestID, err) // Ошибки не встраиваем - накапливаем
						}
					} else {
						rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV)) // Успешная обработка - добавляем в slice структуру из указателя
					}
					innerErrors.AppendErrors(localErrors)
				} // создать строку во внешнем сервисе
			}
		} // Обработаем все строки во входной структуре

		_metrics.IncMetaCountVec("CreateMulti", entity.Name)
		_metrics.AddMetaDurationVec("CreateMulti", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			return nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - error 'Create multi'", entity.Name)), errors
		} else {
			return rowsOut, nil, errors
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowsIn != nil {}", []interface{}{s, rowsIn}).PrintfError(), errors
}

// createSingleUnsafe создать строку во внешнем сервисе
func (s *Service) createSingleUnsafe(ctx context.Context, requestID uint64, rowIn *_meta.Object, cascadeUp int, cascadeDown int, useCache bool) (rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowIn != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		entity := rowIn.Entity

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Обрабатываем только структуры
		if rowIn.IsSlice {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowIn.IsSlice {}", []interface{}{s, rowIn}).PrintfError(), errors
		}

		// Разрешено ли создавать
		if entity.Modify.CreateRestrict {
			return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
		}

		// Предобработка Marshal - каскадная обработка, вычисления - валидацию не делать
		if !errors.HasError() && !innerErrors.HasError() {
			// Вычисляемые поля перед обработкой
			if !rowIn.Options.Global.SkipCalculation {
				if err = s.processExprs(ctx, requestID, nil, rowIn, _meta.EXPR_ACTION_PRE_MARSHAL); err != nil {
					errors.Append(requestID, err)
				}
			}

			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			// Разбираем сообщение, делаем вычисления PRE_MARSHAL, POST_MARSHAL
			if err, localErrors = s.processMarshal(ctx, requestID, rowIn, PERSIST_ACTION_CREATE, _meta.EXPR_ACTION_MARSHAL, cascadeUp, cascadeDown, false, true, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Постобработка Marshal - каскадная обработка, вычисления, валидация, фильтрация
		if !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода

			// Обрабатываем проверки на существование записей, фильтруем, делаем вычисления POST_MARSHAL
			if err, localErrors = s.processPost(ctx, requestID, rowIn, PERSIST_ACTION_CREATE, _meta.EXPR_ACTION_POST_MARSHAL, false, true, true, false, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)

			// Обрабатываем фильтруем делаем вычисления PRE_PUT, валидацию
			if err, localErrors = s.processPost(ctx, requestID, rowIn, PERSIST_ACTION_CREATE, _meta.EXPR_ACTION_PRE_PUT, true, true, false, true, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Если требуется сохранение
		if rowIn.Options.Global.Persist {
			if !errors.HasError() && !innerErrors.HasError() {

				localCtx := ctx // Контекст, в котором будем создавать
				isGlobalTx := rowIn.Options.Global.TxExternal != 0
				localTxId := uint64(0)
				var storage _storage.Service

				if !isGlobalTx {
					// Если не работаем в глобальной транзакции, то создадим новую локальную транзакцию
					if storage, err = s.getStorageByEntity(entity); err != nil {
						errors.Append(requestID, err)
					} else {
						// Проверим, что нет уже локальной транзакции в текущем контексте
						if txId := fromContextTxId(ctx); txId == 0 {
							// создаем новую транзакцию и помещаем в контекст
							if localTxId, err = storage.Begin(ctx, requestID); err != nil {
								errors.Append(requestID, err)
							} else {
								localCtx = contextWithTxId(ctx, localTxId)
							}
						}
					}
				}

				// Поместить данные в globalCache или внешнего источника
				if !errors.HasError() && !innerErrors.HasError() {
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if rowOut, err, localErrors = s.persistSingleUnsafe(localCtx, requestID, PERSIST_ACTION_CREATE, rowIn, cascadeUp, cascadeDown, useCache); err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)

					// Если работаем в локальной транзакции
					if localTxId != 0 && storage != nil {
						if localErrors.HasError() {
							if err = storage.Rollback(requestID, localTxId); err != nil {
								errors.Append(requestID, err)
							}
						} else {
							if err = storage.Commit(requestID, localTxId); err != nil {
								errors.Append(requestID, err)
							}
						}
					}
				} else {
					rowOut = rowIn
				}
			} else {
				rowOut = rowIn // Сохранение не требуется - транслируем назад подготовленные и валидированные данные
			}

		} else {
			// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
			if err = rowIn.SetXmlNameValueFromTag(); err != nil {
				errors.Append(requestID, err)
			}
			rowOut = rowIn // Сохранение не требуется - транслируем назад подготовленные и валидированные данные
		}

		_metrics.IncMetaCountVec("CreateSingle", entity.Name)
		_metrics.AddMetaDurationVec("CreateSingle", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			err = s.processErrors(requestID, rowIn, errors, rowIn.Options.Global.EmbedError, "Create Single")
			return rowIn, err, errors
		} else {
			return rowOut, nil, errors
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError(), errors
}
