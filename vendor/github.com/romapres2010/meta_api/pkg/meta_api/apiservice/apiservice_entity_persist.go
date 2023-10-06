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

// PersistMarshal создать строку во внешнем сервисе
func (s *Service) PersistMarshal(ctx context.Context, action Action, entityName string, inBuf []byte, inFormat string, queryOptions _meta.QueryOptions) (outBuf []byte, outFormat string, err error, errors _err.Errors) {
	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context

	if s != nil && entityName != "" {

		var localCtx = contextWithOptionsCache(ctx) // Создадим новый контекст и встроим в него OptionsCache
		var entity *_meta.Entity
		var options *_meta.Options
		var outObject *_meta.Object

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, entityName", requestID, entityName)

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

		if entity = s.GetEntityUnsafe(entityName); entity == nil {
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)), errors
		}

		if options, err = s.ParseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
			return nil, inFormat, err, errors
		}
		options.Global.InFormat = inFormat

		switch action {
		case PERSIST_ACTION_CREATE:
			if entity.Modify.CreateRestrict {
				return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
			}
		case PERSIST_ACTION_UPDATE:
			if entity.Modify.UpdateRestrict {
				return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to update", entity.Name)), errors
			}
		default:
			return nil, inFormat, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - 'PersistMarshal' incorrect action '%s'", entity.Name, action)), errors
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
			var rowsInI []interface{}

			// Распарсить многострочный запрос - список полей для JSON вернется только те, что в запросе
			if rowsInI, err = s.UnmarshalMultiInterface(requestID, entity, inBuf, inFormat); err != nil {
				return nil, inFormat, err, errors
			}

			{ // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if outObject, err, localErrors = s.PersistMultiUnsafe(localCtx, requestID, action, entity, options, rowsInI, options.CascadeUp, options.CascadeDown, options.Global.UseCache); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
		} else {
			var rowInI interface{}

			// Распарсить запрос - список полей для JSON вернется только те, что в запросе
			if rowInI, err = s.UnmarshalSingleInterface(requestID, entity, inBuf, inFormat); err != nil {
				return nil, inFormat, err, errors
			}

			{ // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if outObject, err, localErrors = s.PersistSingleUnsafe(localCtx, requestID, action, entity, options, rowInI, options.CascadeUp, options.CascadeDown, options.Global.UseCache); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // создать строки во внешнем сервисе - поддерживается встраивание ошибок в исходный запрос
		}

		// сформируем ответ, возможно с уже встроенной ошибкой
		if outObject != nil {
			var errInner error
			outBuf, errInner = s.MarshalEntity(requestID, outObject.Value, "Persist", entityName, options.Global.OutFormat)
			if errInner != nil {
				_log.Debug("ERROR - MarshalEntity: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				errors.Append(requestID, errInner)
				return nil, inFormat, errInner, errors
			} else {
				//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				return outBuf, options.Global.OutFormat, err, errors
			}
		} else {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
			return nil, inFormat, err, errors
		}
	}
	return nil, inFormat, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entityName != '' {}", []interface{}{s, entityName}).PrintfError(), errors
}

// PersistMultiUnsafe сохранить строки во внешнем сервисе
func (s *Service) PersistMultiUnsafe(ctx context.Context, requestID uint64, action Action, entity *_meta.Entity, options *_meta.Options, rowsInI []interface{}, cascadeUp int, cascadeDown int, useCache bool) (rowsOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowsInI != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

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
				err = _recover.GetRecoverError(r, requestID, "PersistMultiUnsafe", entity.Name)
			}
		}()

		switch action {
		case PERSIST_ACTION_CREATE:
			if entity.Modify.CreateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
			}
		case PERSIST_ACTION_UPDATE:
			if entity.Modify.UpdateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to update", entity.Name)), errors
			}
		default:
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - 'PersistMultiUnsafe' incorrect action '%s'", entity.Name, action)), errors
		}

		// Собственно slice, из которого будем считывать данные
		rowsInIValueLen := len(rowsInI)

		// Сформировать выходную структуру - slice []interface{}, так как на вход могли прийти разные структуры
		if rowsOut, err = s.newSliceAnyRestrict(requestID, entity, options, 0, rowsInIValueLen); err != nil {
			return nil, err, errors
		}

		// Собственно slice, в который будем вставлять
		rowsOutRV := reflect.Indirect(rowsOut.RV)

		{ // Обработаем все строки во входной структуре
			for i := 0; i < rowsInIValueLen; i++ {
				rowInI := rowsInI[i]

				var rowOut *_meta.Object
				{ // обработать строку во внешнем сервисе
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if rowOut, err, localErrors = s.PersistSingleUnsafe(ctx, requestID, action, entity, options, rowInI, cascadeUp, cascadeDown, false); err != nil {
						// Если ошибки встраиваем, то добавить структуру ответа в общий список
						if options.Global.EmbedError && rowOut != nil {
							rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV)) // добавляем в slice структуру из указателя - ошибки встроены
						} else {
							errors.Append(requestID, err) // Ошибки не встраиваем - накапливаем
						}
					} else {
						rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV)) // Успешная обработка - добавляем в slice структуру из указателя
					}
					innerErrors.AppendErrors(localErrors)
				} // обработать строку во внешнем сервисе
			}
		} // Обработаем все строки во входной структуре

		_metrics.IncMetaCountVec("PersistMultiUnsafe", entity.Name)
		_metrics.AddMetaDurationVec("PersistMultiUnsafe", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			return nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - error 'Create multi'", entity.Name)), errors
		} else {
			return rowsOut, nil, errors
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowsInI != nil {}", []interface{}{s, rowsInI}).PrintfError(), errors
}

// PersistSingleUnsafe создать строку во внешнем сервисе
func (s *Service) PersistSingleUnsafe(ctx context.Context, requestID uint64, action Action, entity *_meta.Entity, options *_meta.Options, rowInI interface{}, cascadeUp int, cascadeDown int, useCache bool) (rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowInI != nil {

		var rowIn *_meta.Object
		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		persistRestrictFields := false  // Признак - ограничить поля теми, что пришли на вход в MarshalEntity
		persistUseUK := false           // Признак - устанавливать PK по заполненному UK
		persistUpdateAllFields := false // Признак - обновлять все поля объекта
		addCrossRef := false            // Признак - добавлять кросс ссылки composition-association
		addRefFields := false           // Признак - добавлять ссылочные поля для соответствующих Reference
		addUKFields := false            // Признак - добавлять поля всех UK
		checkExists := false
		checkNotExists := false

		switch action {
		case PERSIST_ACTION_CREATE:
			if entity.Modify.CreateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
			}
			persistRestrictFields = false  // Для Create НЕ ограничивать список полей, тем, что пришло на вход - вставляем всю структуру
			persistUpdateAllFields = false // Для Create НЕ применим
			persistUseUK = false           // Для Create НЕ нужно искать и подменять PK по UK
			addCrossRef = true             // Для Create нужно включать кросс ссылки на родителя
			addRefFields = false           // Для Create НЕ нужно добавлять ссылочные поля, чтобы разбирались Association
			addUKFields = false            // Для Create НЕ нужно добавлять поля UK
			checkExists = true
			checkNotExists = false

		case PERSIST_ACTION_UPDATE:
			if entity.Modify.UpdateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to update", entity.Name)), errors
			}
			//persistRestrictFields = options.Global.PersistRestrictFields   // Для Update ограничивать список полей, тем, что пришло на вход
			persistRestrictFields = true                                   // Для Update ограничивать список полей, тем, что пришло на вход
			persistUpdateAllFields = options.Global.PersistUpdateAllFields // Для Update обновлять все поля объекта
			persistUseUK = options.Global.PersistUseUK                     // Для Update нужно искать и подменять PK по UK
			addCrossRef = false                                            // Для Update НЕ нужно включать кросс ссылки на родителя
			addRefFields = true                                            // Для Update нужно добавлять ссылочные поля, чтобы разбирались Association
			addUKFields = options.Global.PersistUseUK                      // Для Update нужно добавлять поля UK
			checkExists = false
			checkNotExists = true

		default:
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - 'PersistSingleUnsafe' incorrect action '%s'", entity.Name, action)), errors
		}

		{ // Парсим сообщение и готовим к persist

			// Сформировать struct из map[string]interface{} и скопировать только нужные поля
			if rowIn, err = s.MarshalMap(requestID, entity, options, rowInI, persistRestrictFields, addRefFields, addUKFields); err != nil {
				return nil, err, errors
			}

			{ // Разбираем сообщение
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processMarshal(ctx, requestID, rowIn, action, _meta.EXPR_ACTION_NONE, cascadeUp, cascadeDown, &processOptions{validate: false, calculate: true, addCrossRef: false, isComposition: true, persistRestrictFields: persistRestrictFields, addRefFields: addRefFields, addUKFields: addUKFields}); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // Разбираем сообщение

			{ // Второй проход, проверяем ассоциации, выстраиваем ключи по иерархии
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPrePersist(ctx, requestID, rowIn, action, _meta.EXPR_ACTION_NONE, cascadeUp, cascadeDown, &processOptions{validate: false, calculate: true, addCrossRef: addCrossRef, isComposition: true, persistUseUK: persistUseUK, persistUpdateAllFields: persistUpdateAllFields}); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // Второй проход, проверяем ассоциации, выстраиваем ключи по иерархии

			{ // Делаем, валидацию, проверяем наличие записей
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				// TODO - перейти на общие опции
				if err, localErrors = s.processPost(ctx, requestID, rowIn, action, _meta.EXPR_ACTION_NONE, true, false, false, checkExists, checkNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			} // Делаем, валидацию, проверяем наличие записей

		} // Парсим сообщение и готовим к persist

		// Если требуется сохранение
		if options.Global.Persist {
			if !errors.HasError() && !innerErrors.HasError() {

				localCtx := ctx // Контекст, в котором будем создавать
				isGlobalTx := options.Global.TxExternal != 0
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
					if rowOut, err, localErrors = s.persistSingleUnsafe(localCtx, requestID, action, rowIn, cascadeUp, cascadeDown, useCache); err != nil {
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
					rowOut = rowIn // Вернем назад исходные данные для встраивания ошибки
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

		_metrics.IncMetaCountVec("PersistSingleUnsafe", entity.Name)
		_metrics.AddMetaDurationVec("PersistSingleUnsafe", entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if errors.HasError() {
			err = s.processErrors(requestID, rowIn, errors, rowIn.Options.Global.EmbedError, "Persist Single")
			return rowIn, err, errors
		} else {
			return rowOut, nil, errors
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowInI != nil {}", []interface{}{s, rowInI}).PrintfError(), errors
}

// persistSingleUnsafe сохранить строку во внешнем сервисе
func (s *Service) persistSingleUnsafe(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object, cascadeUp int, cascadeDown int, useCache bool) (rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowIn != nil {

		entity := rowIn.Entity
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

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

		switch action {
		case PERSIST_ACTION_CREATE:
			if entity.Modify.CreateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)), errors
			}
		case PERSIST_ACTION_UPDATE:
			if entity.Modify.UpdateRestrict {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to update", entity.Name)), errors
			}
		default:
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - 'persistSingleUnsafe' incorrect action '%s'", entity.Name, action)), errors
		}

		{ // Поместить данные в globalCache или внешнего источника
			var rowResult *_meta.Object

			// Возвращаем другой состав полей чем сохраняем
			optionsOut := rowIn.Options.Clone()
			if rowIn.Options.Global.PersistUpdateAllFields {
				optionsOut.Fields = nil  // Нужны все поля
				optionsOut.CascadeUp = 1 // Запросить со всеми ассоциациями
			}

			if useCache {
				// При использовании кэша будут созданы и возвращены все поля
				if rowResult, err = s.persistCacheSingle(ctx, requestID, action, rowIn, optionsOut); err != nil {
					errors.Append(requestID, err)
				} else {
					if rowResult != nil {
						if rowIn.Options.Fields != nil {
							// Сформировать выходную структуру с нужным составом полей
							if rowOut, err = s.newRowRestrict(requestID, entity, rowIn.Options); err != nil {
								errors.Append(requestID, err)
							} else {
								// Скопируем только те поля, которые нужные в ответе
								if err = entity.CopyObjectStruct(rowResult, rowOut, rowIn.Options.Fields); err != nil {
									errors.Append(requestID, err)
								}
							}
						} else {
							rowOut = rowResult // копировать структуры не нужно
						}
					}
				}
			} else {
				// При прямом создании возвращаются только запрошенные поля копировать не нужно
				if rowResult, err = s.persistExternalSingle(ctx, requestID, action, rowIn, optionsOut); err != nil {
					errors.Append(requestID, err)
				} else {
					rowOut = rowResult // копировать структуры не нужно, так как возвращаются только нужные поля
				}
			}
		} // Поместить данные в globalCache или внешнего источника

		// Каскадная обработка для создания подчиненных объектов - родительские объекты не обрабатывать
		if !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processPersist(ctx, requestID, action, _meta.EXPR_ACTION_INSIDE_PUT, rowIn, rowOut, cascadeUp, cascadeDown, &processOptions{}); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Пост обработка только для полей, которые нужно возвращать - каскадно вниз не делать иначе потеряем накопленные ошибки
		if !errors.HasError() && !innerErrors.HasError() {
			// Каскадно вниз не делать, иначе потеряем накопленные ошибки в несозданных записях
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			// TODO - не выводить для вложенных композиций родительские ассоциации
			if err, localErrors = s.processGet(ctx, requestID, rowIn, rowOut, PERSIST_ACTION_NONE, _meta.EXPR_ACTION_POST_GET, 1, 0, &processOptions{validate: rowOut.Options.Global.Validate, calculate: true, addCrossRef: false}); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		if errors.HasError() {
			err = s.processErrors(requestID, rowIn, errors, rowIn.Options.Global.EmbedError, "Persist Single")
			return rowIn, err, errors
		} else {
			return rowOut, nil, errors
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError(), errors
}

// persistCacheSingle создать строку во внешнем сервисе
func (s *Service) persistCacheSingle(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object, optionsOut *_meta.Options) (rowOut *_meta.Object, err error) {
	if s != nil && rowIn != nil {

		//tic := time.Now()
		//entity := rowIn.Entity
		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Создадим во внешнем сервисе и вернем ВСЕ поля
		if rowOut, err = s.persistExternalSingle(ctx, requestID, action, rowIn, optionsOut); err != nil {
			return nil, err
		}

		// Добавим считанные данные в globalCache
		if rowOut != nil {
			// В кэш сохраняем все поля из внешнего сервиса со всеми ключами
			if err = s.cacheSetRowUnsafe(nil, nil, rowOut); err != nil {
				return nil, err
			}
		}

		//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError()
}

// persistExternalSingle создать строку во внешнем сервисе
func (s *Service) persistExternalSingle(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object, optionsOut *_meta.Options) (rowOut *_meta.Object, err error) {
	if s != nil && rowIn != nil {

		//tic := time.Now()
		entity := rowIn.Entity
		txId := rowIn.Options.Global.TxExternal // по умолчанию работаем в глобальной транзакции

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

		storage, err := s.getStorageByEntity(entity)
		if err != nil {
			return nil, err
		}

		// Если на указана глобальная транзакция, то ищем локальную из контекста
		if txId == 0 {
			txId = fromContextTxId(ctx)
		}

		if rowOut, err = s.newRowRestrict(requestID, entity, optionsOut); err != nil {
			return nil, err
		}

		localAction := PERSIST_ACTION_NONE
		switch action {
		case PERSIST_ACTION_CREATE:
			if entity.Embed {
				localAction = PERSIST_ACTION_UPDATE // для встроенных объектов всегда выполняется обновление
			} else {
				localAction = PERSIST_ACTION_CREATE
			}
		default:
			localAction = action
		}

		switch localAction {
		case PERSIST_ACTION_CREATE:
			// создать строку и получить назад указанные поля из БД
			if err = storage.Create(ctx, requestID, txId, rowIn, rowOut); err != nil {
				return nil, err
			}
		case PERSIST_ACTION_UPDATE:
			// обновить строку и получить назад указанные поля из БД
			if err = storage.Update(ctx, requestID, txId, rowIn, rowOut); err != nil {
				return nil, err
			}
		default:
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s] - incorrect action '%s'", rowOut.Entity.Name, rowOut.KeysValueString(), action))
		}

		// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
		if err = rowOut.SetXmlNameValueFromTag(); err != nil {
			return nil, err
		}

		//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError()
}
