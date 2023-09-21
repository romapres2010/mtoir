package apiservice

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// persistSingleUnsafe сохранить строку во внешнем сервисе
func (s *Service) persistSingleUnsafe(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object, cascadeUp int, cascadeDown int, useCache bool) (rowOut *_meta.Object, err error, errors _err.Errors) {
	if s != nil && rowIn != nil {

		entity := rowIn.Entity
		innerErrors := _err.Errors{} // Ошибки вложенных методов

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

		{ // Поместить данные в globalCache или внешнего источника
			var rowResult *_meta.Object

			if useCache {
				// При использовании кэша будут созданы и возвращены все поля
				if rowResult, err = s.persistCacheSingle(ctx, requestID, action, rowIn); err != nil {
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
				if rowResult, err = s.persistExternalSingle(ctx, requestID, action, rowIn); err != nil {
					errors.Append(requestID, err)
				} else {
					rowOut = rowResult // копировать структуры не нужно, так как возвращаются только нужные поля
				}
			}
		} // Поместить данные в globalCache или внешнего источника

		// Каскадная обработка для создания подчиненных объектов - родительские объекты не обрабатывать
		if !errors.HasError() && !innerErrors.HasError() {
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processPersist(ctx, requestID, action, _meta.EXPR_ACTION_INSIDE_PUT, rowIn, rowOut, cascadeUp, cascadeDown); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Пост обработка только для полей, которые нужно возвращать - каскадно вниз не делать иначе потеряем накопленные ошибки
		if !errors.HasError() && !innerErrors.HasError() {
			// Каскадно вниз не делать, иначе потеряем накопленные ошибки в несозданных записях
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processGet(ctx, requestID, rowIn, rowOut, PERSIST_ACTION_CREATE, _meta.EXPR_ACTION_POST_GET, 1, 0, rowOut.Options.Global.Validate, true); err != nil {
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
func (s *Service) persistCacheSingle(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object) (rowOut *_meta.Object, err error) {
	if s != nil && rowIn != nil {

		tic := time.Now()
		entity := rowIn.Entity

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Создадим во внешнем сервисе и вернем ВСЕ поля
		if rowOut, err = s.persistExternalSingle(ctx, requestID, action, rowIn); err != nil {
			return nil, err
		}

		// Добавим считанные данные в globalCache
		if rowOut != nil {
			// В кэш сохраняем все поля из внешнего сервиса со всеми ключами
			if err = s.cacheSetRowUnsafe(nil, nil, rowOut); err != nil {
				return nil, err
			}
		}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError()
}

// persistExternalSingle создать строку во внешнем сервисе
func (s *Service) persistExternalSingle(ctx context.Context, requestID uint64, action Action, rowIn *_meta.Object) (rowOut *_meta.Object, err error) {
	if s != nil && rowIn != nil {

		tic := time.Now()
		entity := rowIn.Entity
		txId := rowIn.Options.Global.TxExternal // по умолчанию работаем в глобальной транзакции

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		storage, err := s.getStorageByEntity(entity)
		if err != nil {
			return nil, err
		}

		// Если на указана глобальная транзакция, то ищем локальную из контекста
		if txId == 0 {
			txId = fromContextTxId(ctx)
		}

		if rowOut, err = s.newRowRestrict(requestID, entity, rowIn.Options); err != nil {
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

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError()
}
