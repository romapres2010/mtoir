package apiservice

import (
	"context"
	"fmt"
	"reflect"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// processPersist - создание одной строки и всех подчиненных объектов этой строки
func (s *Service) processPersist(ctx context.Context, requestID uint64, action Action, exprAction _meta.ExprAction, rowIn *_meta.Object, rowOut *_meta.Object, cascadeUp int, cascadeDown int) (err error, errors _err.Errors) {
	if s != nil && rowIn != nil && rowOut != nil && rowIn.Options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, EntityName", requestID, rowOut.Entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Обрабатываем только структуры
		if rowOut.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, rowOut}).PrintfError(), errors
		}

		//// Обработать FK M:1 - при отрицательном cascadeUp - обрабатываем все уровни без ограничений
		//if rowOut.Entity.HasAssociations() && cascadeUp != 0 {
		//    //  уровень уменьшаем на 1, 0 - означает только себя
		//    localErrors := _err.Errors{} // локальные ошибки вложенного метода
		//    if err, localErrors = s.processAssociations(ctx, requestID, rowIn, rowOut, _meta.EXPR_ACTION_NULL, cascadeUp-1, 0, false, false, s.createAssociation); err != nil {
		//        errors.Append(requestID, err)
		//    }
		//    innerErrors.AppendErrors(localErrors)
		//}

		// Обработать FK 1:M - при отрицательном cascadeDown - обрабатываем все уровни без ограничений
		if rowOut.Entity.HasCompositions() && cascadeDown != 0 {
			//  уровень уменьшаем на 1, 0 - означает только себя
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processCompositions(ctx, requestID, rowIn, rowOut, action, exprAction, cascadeUp, cascadeDown-1, false, false, s.pesistComposition); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
		err = s.processErrors(requestID, rowOut, errors, rowOut.Options.Global.EmbedError, "Create - process Composition")
		return err, errors
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil && rowOut != nil && rowIn.Options != nil {}", []interface{}{s, rowIn, rowOut}).PrintfError(), errors
}

func (s *Service) pesistComposition(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool) (compositionOutRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowIn != nil && compositionField != nil && rowIn.Options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := compositionField.Reference()
		toEntity := reference.ToEntity()
		toKey := reference.ToKey()

		_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowIn.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString()) // Консолидируем все ошибки

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				compositionOutRows = nil
				keyArgs = nil
				err = _recover.GetRecoverError(r, requestID, "pesistComposition", rowIn.Entity.Name)
			}
		}()

		// Найдем входной slice composition в исходной строке rowIn
		if compositionInRows, ok := rowIn.CompositionMap[reference]; ok {

			// Обрабатываем только непустые slice
			if compositionInRowsLen := len(compositionInRows.Objects); compositionInRowsLen > 0 {

				optionsRef, err := s.parseQueryOptions(ctx, requestID, rowIn.Options.Key+".composition."+toEntity.Name, toEntity, compositionField, rowIn.Options.QueryOptionsDown, rowIn.Options.Global)
				if err != nil {
					return nil, nil, err, errors
				}

				if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
					// На вход получаем только указатели на slice
					if !compositionInRows.IsSlice {
						return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - must be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
					} else {
						// Создадим выходной slice composition под список полей
						if compositionOutRows, err = s.newSliceRestrict(requestID, toEntity, optionsRef, 0, compositionInRowsLen); err != nil {
							return nil, nil, err, errors
						}
					}
				} else {
					// На вход НЕ получаем только указатели на slice
					if compositionInRows.IsSlice {
						return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - must NOT be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
					}
				}

				// Обработаем все строки во входной структуре
				for _, compositionInRow := range compositionInRows.Objects {

					localAction := PERSIST_ACTION_NONE

					switch action {
					case PERSIST_ACTION_CREATE:
						if reference.Embed {
							localAction = PERSIST_ACTION_UPDATE // для встроенных объектов выполняется обновление
						} else {
							localAction = PERSIST_ACTION_CREATE
						}
					case PERSIST_ACTION_UPDATE:
						localAction = PERSIST_ACTION_UPDATE
					default:
						return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - incorrect action '%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, action)), errors
					}

					// Рекурсивно создадим объект и получим указатель на структуру - при сохранении композиции не кешировать
					compositionOutRow, err, localErrors := s.persistSingleUnsafe(ctx, requestID, localAction, compositionInRow, cascadeUp, cascadeDown, false)
					if err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)

					// Добавляем в массив объектов - даже если были ошибки
					if compositionOutRow != nil {
						if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
							compositionOutRowsRV := reflect.Indirect(compositionOutRows.RV)                      // Собственно выходной slice, в который будем вставлять *[]struct{} -> []struct{}
							compositionOutRowsRV.Set(reflect.Append(compositionOutRowsRV, compositionOutRow.RV)) // Успешная обработка - вставляем в slice *struct{}
						} else {
							compositionOutRows = compositionOutRow // Обрабатываем только одну строку
						}
						compositionOutRows.AppendObject(compositionOutRow) // Добавляем в массив объектов
					}
				}

				if errors.HasError() {
					_log.Debug("ERROR: requestID, entityName, duration", requestID, rowIn.Entity.Name, time.Now().Sub(tic))
					return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - error 'Create'", rowIn.Entity.Name, rowIn.KeysValueString(), compositionField.Name)), errors
				} else {
					_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowIn.Entity.Name, time.Now().Sub(tic))
					return compositionOutRows, keyArgs, nil, errors
				}
			}
		}
		// Данных нет, обрабатывать нечего
		return nil, nil, nil, errors
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil && compositionField != nil && rowIn.Options != nil {}", []interface{}{s, rowIn, compositionField}).PrintfError(), errors
}
