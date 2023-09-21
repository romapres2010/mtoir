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
)

// ConvertMarshal преобразовать между сущностями
func (s *Service) ConvertMarshal(ctx context.Context, entityName string, inBuf []byte, inFormat string, queryOptions _meta.QueryOptions) (outBuf []byte, outFormat string, err error) {
	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context

	if s != nil && s.storageMap != nil && entityName != "" {

		var localCtx = contextWithOptionsCache(ctx) // Создадим новый контекст и встроим в него OptionsCache
		var entity *_meta.Entity
		var options *_meta.Options
		var fromObject *_meta.Object
		var outObject *_meta.Object

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, entityName)

		// Meta может меняться в online - повесим запрет на чтение
		s.metaRLock()
		defer s.metaRUnLock()

		if entity = s.getEntityUnsafe(entityName); entity == nil {
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)).PrintfError()
		}

		if options, err = s.parseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
			return nil, inFormat, err
		}

		if options.FromEntity == nil {
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - query parameter '%s' does not defined - can not convert", entity.Name, s.cfg.QueryOption.FromEntity)).PrintfError()
		}

		//// Разрешено ли вставлять
		//if entity.Modify.CreateRestrict {
		//	return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name)).PrintfError()
		//}

		if options.Global.MultiRow {
			// Распарсить многострочный запрос в формате FromEntity
			if fromObject, err = s.unmarshalMulti(requestID, options.FromEntity, options, inBuf, inFormat, options.Global.IgnoreExtraField); err != nil {
				return nil, inFormat, err
			}
		} else {
			// Распарсить однострочный запрос в формате FromEntity
			if fromObject, err = s.unmarshalSingle(requestID, options.FromEntity, options, inBuf, inFormat, options.Global.IgnoreExtraField); err != nil {
				return nil, inFormat, err
			}
		}

		{ // для ссылочных полей добавим ограничение - считывать только ключи
			var refKeyFields []string
			refFieldsName := options.QueryOptionsDown[s.cfg.QueryOption.FieldsFull]
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
				options.QueryOptionsDown[s.cfg.QueryOption.FieldsFull] = refFieldsName + strings.Join(refKeyFields, ",")
			}
		} // для ссылочных полей добавим ограничение - считывать только ключи

		if options.Global.MultiRow {
			// конвертировать массив строк
			if outObject, err = s.convertMultiUnsafe(localCtx, requestID, entity, options, fromObject); err != nil {
				if outObject == nil { // структуры с встроенной ошибкой нет - возвращаем просто ошибки
					return nil, inFormat, err
				}
			}
		} else {
			// конвертировать строку
			if outObject, err = s.convertSingleUnsafe(localCtx, requestID, entity, options, fromObject); err != nil {
				if outObject == nil { // структуры с встроенной ошибкой нет - возвращаем просто ошибки
					return nil, inFormat, err
				}
			}
		}

		// сформируем ответ, возможно с уже встроенной ошибкой
		if outObject != nil {
			var errInner error
			outBuf, errInner = s.marshal(requestID, outObject.Value, "Convert", entityName, options.Global.OutFormat)
			if errInner != nil {
				_log.Debug("ERROR - marshal: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				return nil, inFormat, errInner
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
				return outBuf, options.Global.OutFormat, err
			}
		} else {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
			return nil, inFormat, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty outObject", entityName))
		}
	}
	return nil, inFormat, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entityName != '' {}", []interface{}{s, entityName}).PrintfError()
}

func (s *Service) convertMultiUnsafe(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, rowsFrom *_meta.Object) (rowsOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && rowsFrom != nil && options != nil {

		tic := time.Now()
		errors := _err.Errors{}
		action := _meta.EXPR_ACTION_INSIDE_GET

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				rowsOut = nil
				err = _recover.GetRecoverError(r, requestID, "convertMultiUnsafe", entity.Name)
			}
		}()

		//// Разрешено ли
		//if entity.Modify.CreateRestrict {
		//	return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name))
		//}

		// Собственно slice, из которого будем считывать данные
		rowsFromRV := reflect.Indirect(rowsFrom.RV)
		rowsFromRVLen := rowsFromRV.Len()

		// Временный объект, в который буем копировать элементы slice для дальнейшей обработки
		rowFromTmp, err := s.newRowRestrict(requestID, entity, options)
		if err != nil {
			return nil, err
		}

		// Сформировать выходную структуру - указатель на slice
		if rowsOut, err = s.newSliceRestrict(requestID, entity, options, 0, rowsFromRVLen); err != nil {
			return nil, err
		}

		// Собственно slice, в который будем вставлять
		rowsOutRV := reflect.Indirect(rowsOut.RV)

		{ // Конвертировать все строки из исходного в целевой набор
			for i := 0; i < rowsFromRVLen; i++ {
				rowFromPtrRV := rowsFromRV.Index(i) // указатель на текущий элемент массива

				rowFromTmp.SetFromRV(rowFromPtrRV)
				//rowFromTmp := rowsFrom.NewFromRV(rowFromPtrRV, false)
				//rowsFrom.AppendObject(rowFromTmp)

				// конвертировать строку
				if rowOut, err := s.convertSingleUnsafe(ctx, requestID, entity, options, rowFromTmp); err != nil {

					// Если ошибки встраиваем, то добавить структуру ответа в общий список
					if options.Global.EmbedError && rowOut != nil {
						// добавляем в slice структуру из указателя - ошибки встроены
						rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV))
					} else {
						// Ошибки не встраиваем
						errors.Append(requestID, err) // накапливаем ошибки
					}
				} else {
					// Успешная обработка - добавляем в slice структуру из указателя
					rowsOutRV.Set(reflect.Append(rowsOutRV, rowOut.RV))
				}
			}
		} // Конвертировать все строки из исходного в целевой набор

		{ // Если в опциях есть динамическая фильтрация, то отработаем ее
			if rowsFromRVLen > 0 {
				if rowsOut != nil {
					if expr := rowsOut.Options.FilterPreExpr; expr != nil {
						// Только выражения, которые применимы к типу запроса
						if expr.CheckAction(action) {
							if rowsOutFiltering, err := s.processRowsFiltering(requestID, expr, rowsOut); err != nil {
								return nil, err
							} else {
								// Переопределим после фильтрации
								if rowsOutFiltering != nil { // Фильтрация может удалить все строки
									rowsOut = rowsOutFiltering
								} else {
									rowsOut = nil
								}
							}
						}
					}
				} else {
					// ситуация крайне странная
					return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowsOut != nil {}", []interface{}{rowsOut}).PrintfError()
				}
			}
		} // Если в опциях есть динамическая фильтрация, то отработаем ее

		_metrics.IncMetaCountVec("ConvertMulti from", rowsFrom.Entity.Name)
		_metrics.AddMetaDurationVec("ConvertMulti from", rowsFrom.Entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if len(errors) > 0 {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return nil, errors.Error(requestID, fmt.Sprintf("Create Multi Entity '%s' - ERROR", entity.Name))
		} else {
			_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return rowsOut, nil
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && rowsFrom != nil && options != nil {}", []interface{}{s, entity, rowsFrom, options}).PrintfError()
}

func (s *Service) convertSingleUnsafe(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, rowFrom *_meta.Object) (rowOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		errors := _err.Errors{}      // Все ошибки накапливаем общем массиве
		//keepLock := true

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		////На вход получаем только указатели на struct
		//if rowFrom.RV.Kind() != reflect.Ptr || reflect.Indirect(rowFrom.RV).Kind() != reflect.Struct {
		//	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "rowFrom.Value must be pointer to struct", []interface{}{rowFrom.RV.Kind().String(), reflect.Indirect(rowFrom.RV).Kind()}).PrintfError()
		//}

		//// Разрешено ли создавать
		//if entity.Modify.CreateRestrict {
		//	return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - forbiden to create", entity.Name))
		//}

		{ // Предобработка - каскадная обработка, вычисления, валидация
			// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
			if err = rowFrom.SetXmlNameValueFromTag(); err != nil {
				errors.Append(requestID, err)
			}

			// В ходе предобработки нужно вычислить поля, перед их вставкой
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processMarshal(ctx, requestID, rowFrom, PERSIST_ACTION_CREATE, _meta.EXPR_ACTION_PRE_PUT, -1, -1, rowFrom.Options.Global.Validate, true, false); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Предобработка - каскадная обработка, вычисления, валидация

		if options.Fields != nil {
			// Сформировать выходную структуру с нужным составом полей
			if rowOut, err = s.newRowRestrict(requestID, entity, options); err != nil {
				errors.Append(requestID, err)
			} else {
				// Скопируем только те поля, которые нужные в ответе
				if err = entity.CopyObjectStruct(rowFrom, rowOut, options.Fields); err != nil {
					errors.Append(requestID, err)
				}
			}
		} else {
			rowOut = rowFrom
		}

		// TODO - проверка что ключи Association и Composition совпадают с у родителей и детей

		//// TODO - Добавить признак по каким ключа проверять
		//{ // Проверка ключей
		//	mess := _mes.Messages{}
		//
		//	for _, key := range entityIn.Keys {
		//		if key != nil {
		//			exists, _, keyArgs, err := s.getByKeyUnsafe(ctx, requestID, txId, entityIn, queryOptions, rowFrom, key.FieldsMap(), useCache, false, false, outTrace, keepRLock, key)
		//			if err != nil {
		//				errors.Append(requestID, err)
		//			} else {
		//				// Найден дубль по ключу
		//				if exists {
		//					mes := _mes.NewTypedMessage(requestID, _mes.MES_API_DUPLICATE_VALUE_ON_KEY_ERROR, entityIn.Name, key.Name, key.FieldsString(), _meta.ArgsToString("','", keyArgs...))
		//					mess.AddMessage(&mes)
		//				}
		//			}
		//		}
		//	}
		//
		//	// Все ошибки накапливаем общем массиве
		//	if mess.HasAnyError() {
		//		localErrors := mess.GetMessages().Errors(outTrace, 3) // Соберем сообщения в набор ошибок
		//		errors.AppendErrors(requestID, localErrors)
		//	}
		//} // Проверка ключей

		//{ // Конвертировать данные
		//    if len(errors) == 0 {
		//        var rowResult *_meta.Object
		//
		//        if useCache {
		//            // При использовании кэша будут созданы и возвращены все поля
		//            if rowResult, err = s.convertCacheSingle(ctx, requestID, txId, entityIn, entity, rowFrom, outFields, keepLock); err != nil {
		//                errors.Append(requestID, err)
		//            } else {
		//                if rowResult != nil {
		//                    // TODO - перенести после завершения транзакции, иначе в кэш невалидные записи
		//
		//                    // Разблокировать запись в кэш
		//                    //if keepLock {
		//                    //	rowResult.Unlock()
		//                    //}
		//
		//                    // TODO - вынести в общий блок rowInPtr и rowResultPtr
		//                    // Сформировать выходную структуру с нужным составом полей
		//                    if rowOut, err = s.newRowRestrict(requestID, entity, outFields); err != nil {
		//                        errors.Append(requestID, err)
		//                    } else {
		//                        // Скопируем только те поля, которые нужные в ответе
		//                        if err = entity.CopyObjectStruct(rowResult, rowOut, outFields); err != nil {
		//                            errors.Append(requestID, err)
		//                        }
		//                    }
		//                }
		//            }
		//        } else {
		//            // При прямом создании возвращаются только запрошенные поля копировать не нужно
		//            if rowResult, err = s.convertExternalSingle(ctx, requestID, txId, entityIn, entity, rowFrom, outFields); err != nil {
		//                errors.Append(requestID, err)
		//            } else {
		//                rowOut = rowResult // копировать структуры не нужно, так как возвращаются только нужные поля
		//            }
		//        }
		//    }
		//} // Конвертировать данные

		//{ // Постобработка - каскадная обработка, вычисления, валидация
		//	if len(errors) == 0 {
		//		if rowOut != nil {
		//			// Пост обработка только для полей, которые нужно возвращать
		//			if err = s.processGet(ctx, requestID, rowFrom, rowOut, queryOptions, _meta.EXPR_ACTION_GET, 1, 0); err != nil {
		//              errors.Append(requestID, err)
		//			}
		//		} else {
		//			// ситуация крайне странная
		//			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut != nil {}", []interface{}{rowOut}).PrintfError()
		//		}
		//	}
		//} // Постобработка - каскадная обработка, вычисления, валидация

		_metrics.IncMetaCountVec("ConvertSingle from", rowFrom.Entity.Name)
		_metrics.AddMetaDurationVec("ConvertSingle from", rowFrom.Entity.Name, time.Now().Sub(tic))
		_metrics.AddMetaDuration(time.Now().Sub(tic))

		if len(errors) > 0 {
			_log.Debug("ERROR: requestID, entityName, duration, errors", requestID, entity.Name, time.Now().Sub(tic), len(errors))
			resultErr := errors.Error(requestID, fmt.Sprintf("Create Single Entity '%s' - ERROR", entity.Name))

			if !options.Global.EmbedError {
				return nil, resultErr // возвращаем обобщенную ошибку
			} else {
				if rowOut != nil {
					// Встраиваем ошибки
					if err = rowOut.SetErrorValue(errors); err != nil {
						errors.Append(requestID, err)
						return nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - SetErrorValue error", entity.Name))
					}
					return rowOut, resultErr // возвращаем выходной объект со встроенной ошибкой и обобщенную ошибку
				} else {
					if err = rowFrom.SetErrorValue(errors); err != nil {
						errors.Append(requestID, err)
						return nil, errors.Error(requestID, fmt.Sprintf("Entity '%s' - SetErrorValue error", entity.Name))
					}

					// TODO - вынести в общий блок rowInPtr и rowResultPtr
					// Скопировать данные из rowInPtr - возвращаем только указанные поля
					if rowInRestrict, err := s.newRowRestrict(requestID, entity, options); err != nil {
						return nil, err
					} else {
						// Скопируем только те поля, которые нужные в ответе
						if err = entity.CopyObjectStruct(rowFrom, rowInRestrict, options.Fields); err != nil {
							return nil, err
						}
						return rowInRestrict, resultErr // возвращаем входной объект со встроенной ошибкой и обобщенную ошибку
					}
					// TODO - вынести в общий блок rowInPtr и rowResultPtr
				}
			}
		} else {
			// Успешная обработка
			if rowOut != nil {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
				return rowOut, nil
			} else {
				// ситуация крайне странная
				return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut != nil {}", []interface{}{rowOut}).PrintfError()
			}
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {}", []interface{}{s, entity, rowFrom, options}).PrintfError()
}

func (s *Service) convertCacheSingle(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, rowFrom *_meta.Object) (rowOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Создадим во внешнем сервисе и вернем ВСЕ поля
		if rowOut, err = s.convertExternalSingle(ctx, requestID, entity, options, rowFrom); err != nil {
			return nil, err
		}

		//// Добавим считанные данные в globalCache
		//if rowOut != nil {
		//	// TODO - блокировки перенести на уровень выше
		//	//// Заблокируем объект - теперь его можно использовать на чтение
		//	//rowOut.Lock()
		//
		//	// В кэш сохраняем все поля из внешнего сервиса со всеми ключами
		//	if err = s.cacheSetRowUnsafe(entity, nil, rowOut); err != nil {
		//		//rowOut.Unlock()
		//		return nil, err
		//	}
		//
		//	// Не держать блокировку текущего объекта
		//	//if !keepLock {
		//	//	rowOut.Unlock()
		//	//}
		//}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {}", []interface{}{s, entity, rowFrom, options}).PrintfError()
}

func (s *Service) convertExternalSingle(ctx context.Context, requestID uint64, entity *_meta.Entity, options *_meta.Options, rowFrom *_meta.Object) (rowOut *_meta.Object, err error) {
	if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		//if entity.StorageName == "" {
		//	return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty 'storage_name'", entity.Name))
		//}
		//
		//strg, err := s.getStorageByName(entity.StorageName)
		//if err != nil {
		//	return nil, err
		//}

		rowOut = rowFrom

		//if rowOut, err = s.newRowRestrict(requestID, entity, outFields); err != nil {
		//	return nil, err
		//}

		//if entityIn == entity {
		//	// Скопируем только те поля, которые нужные в ответе
		//	if err = entity.CopyObjectStruct(rowFrom, rowOut, outFields); err != nil {
		//		return nil, err
		//	}
		//}

		//// создать строку и получить назад указанные поля из БД
		//if err = strg.Create(ctx, requestID, txId, rowFrom, rowOut); err != nil {
		//	return nil, err
		//}

		// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
		if err = rowOut.SetXmlNameValueFromTag(); err != nil {
			return nil, err
		}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.storageMap != nil && entity != nil && rowFrom != nil && options != nil {}", []interface{}{s, entity, rowFrom, options}).PrintfError()
}
