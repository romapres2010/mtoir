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

// processMarshal - в ходе разбора сообщения marshaling
func (s *Service) processMarshal(ctx context.Context, requestID uint64, row *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool, useExistRow bool) (err error, errors _err.Errors) {
	if s != nil && row != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		_log.Debug("START: requestID, row.EntityName", requestID, row.Entity.Name)

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Обрабатываем только структуры
		if row.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if row.IsSlice {}", []interface{}{s, row}).PrintfError(), errors
		} else {

			// Обработать FK M:1 - при отрицательном cascadeUp - обрабатываем все уровни без ограничений
			if row.Entity.HasAssociations() {
				if cascadeUp != 0 {
					//  уровень уменьшаем на 1, 0 - означает только себя
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					//if err, localErrors = s.processAssociations(ctx, requestID, nil, row, action, exprAction, cascadeUp-1, cascadeDown, validate, calculate, s.marshalAssociation); err != nil {
					// У Associations не должно быть своих Compositions
					if err, localErrors = s.processAssociations(ctx, requestID, nil, row, action, exprAction, cascadeUp-1, 0, validate, calculate, s.marshalAssociation); err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)
				} else {
					// Если разбираем не на всю глубину, то нужно очистить лишние уровни
					if err = s.clearAssociations(ctx, requestID, row); err != nil {
						errors.Append(requestID, err)
					}
				}

			}

			// Вычисляемые поля
			if calculate {
				if !row.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_POST_MARSHAL); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			// Обработать FK 1:M - при отрицательном cascadeDown - обрабатываем все уровни без ограничений
			if row.Entity.HasCompositions() {
				if cascadeDown != 0 {
					//  уровень уменьшаем на 1, 0 - означает только себя
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if err, localErrors = s.processCompositions(ctx, requestID, nil, row, action, exprAction, cascadeUp, cascadeDown-1, validate, calculate, s.marshalComposition); err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)
				} else {
					// Если разбираем не на всю глубину, то нужно очистить лишние уровни
					if err = s.clearCompositions(ctx, requestID, row); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			// Вычисляемые поля
			if calculate {
				if !row.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_POST_MARSHAL); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			// Валидация данных
			if validate {
				if row.Options.Global.Validate && s.validator != nil {
					if err = s.validator.ValidateObject(requestID, row); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, row.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, row, errors, row.Options.Global.EmbedError, "Marshal")
			return err, errors
		}
	}

	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError(), errors
}

func (s *Service) marshalAssociation(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, associationField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool) (associationRow *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && associationField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := associationField.Reference()
		toEntity := reference.ToEntity()
		toKey := reference.ToKey()

		_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				associationRow = nil
				keyArgs = nil
				err = _recover.GetRecoverError(r, requestID, "marshalAssociation", rowOut.Entity.Name)
			}
		}()

		// найдем значение поля, в котором хранятся входные данные, они имеют формат произвольного интерфейса
		associationFieldRV, err := rowOut.FieldRV(associationField)
		if err != nil {
			return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Association field='%s' value by in index", rowOut.Entity.Name, associationField.Name)), errors
		}

		// Проверим, что есть данные для копирования
		if associationFieldRV.IsValid() && !associationFieldRV.IsNil() && !associationFieldRV.IsZero() {

			optionsRef, err := s.parseQueryOptions(ctx, requestID, rowOut.Options.Key+".association."+toEntity.Name, toEntity, associationField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
			if err != nil {
				return nil, nil, err, errors
			}

			// Создадим структуру под список полей
			associationRow, err = s.newRowRestrict(requestID, toEntity, optionsRef)
			if err != nil {
				return nil, nil, err, errors
			}

			{ // Два режима работы, либо получаем на вход *struct{}, либо *map[string]interface{}
				if associationFieldRV.Kind() == reflect.Ptr && reflect.Indirect(associationFieldRV).Kind() == reflect.Struct {
					// TODO - вложенные *struct{} нужно разобрать в Object и выстроить association и composition
					// Если получили *struct{}, то можем брать ее напрямую - конвертация не нужна
					associationRow.SetFromRV(associationFieldRV)

				} else {
					// Ожидаем получить на вход *map[string]interface{} - скопируем только те поля, которые нужные в ответе
					if empty, err := toEntity.CopyMapToStruct(associationFieldRV.Interface(), associationRow, optionsRef.Fields); err != nil {
						errors.Append(requestID, err) // накапливаем ошибки
					} else {
						if empty {
							// Пустая map - на выход
							return nil, nil, nil, errors
						}
					}
				}
			} // Два режима работы, либо получаем на вход *struct{}, либо *map[string]interface{}

			{ // Рекурсивная обработка каскадно
				if associationRow != nil {
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if err, localErrors = s.processMarshal(ctx, requestID, associationRow, action, exprAction, cascadeUp, cascadeDown, validate, calculate, true); err != nil {
						errors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
					}
					innerErrors.AppendErrors(localErrors)
				}
			} // Рекурсивная обработка каскадно

			{ // Если строка существует по PK UK - взять ее, иначе ошибка
				//if useExistRow {
				for _, key := range associationRow.Entity.Keys() {
					if key != nil && (key.Type == _meta.KEY_TYPE_PK || key.Type == _meta.KEY_TYPE_UK) {
						// TODO - добавить локальный cache на время выполнения запроса -целостность чтения на уровне транзакции
						exists, rowUK, ukKeyArgs, err, localErrors := s.getByKeyUnsafe(ctx, requestID, associationRow.Entity, associationRow.Options, associationRow, key)
						innerErrors.AppendErrors(localErrors)
						if err != nil {
							errors.Append(requestID, err)
						} else {
							if exists {
								associationRow.SetFromRV(rowUK.RV) // возьмем найденную строку
								break                              // Не проверяем, что по разным PK UK, получены разные строки
							} else {
								// Данных не найдено при заполненном ключе
								if !_meta.ArgsAllEmpty(ukKeyArgs) {
									errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_NO_DATA_FOUND, requestID, associationRow.Entity.Name, key.Name, key.FieldsString(), _meta.ArgsToString("','", ukKeyArgs...)))
								}
							}
						}
					}
				}
				//}
			} // Если строка существует по PK UK - взять ее, иначе ошибка

			// Вычисляемые поля после рекурсивной обработкой, чтобы они попали в родительские объекты
			if calculate {
				if !associationRow.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, associationRow, _meta.EXPR_ACTION_PRE_MARSHAL); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, associationRow, errors, rowOut.Options.Global.EmbedError, "Association marshal")
			return associationRow, keyArgs, err, errors
		} else {
			// Данных нет, обрабатывать не чего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil  && associationField != nil {}", []interface{}{s, rowOut, associationField}).PrintfError(), errors
}

func (s *Service) marshalComposition(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool) (compositionRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && compositionField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := compositionField.Reference()
		toEntity := reference.ToEntity()
		toKey := reference.ToKey()

		_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				compositionRows = nil
				keyArgs = nil
				err = _recover.GetRecoverError(r, requestID, "marshalComposition", rowOut.Entity.Name)
			}
		}()

		// найдем значение поля, которое будем разбирать
		compositionFieldRV, err := rowOut.FieldRV(compositionField)
		if err != nil {
			return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Composition field='%s' value by in index", rowOut.Entity.Name, compositionField.Name)), errors
		}

		// Извлечем из interface{} -> []map[string]interface{}
		compositionFieldValueRV := reflect.ValueOf(compositionFieldRV.Interface())

		// Если исходных данных нет, то пробросим
		if compositionFieldValueRV.IsValid() && !compositionFieldValueRV.IsZero() {

			optionsRef, err := s.parseQueryOptions(ctx, requestID, rowOut.Options.Key+".composition."+toEntity.Name, toEntity, compositionField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
			if err != nil {
				return nil, nil, err, errors
			}

			compositionFieldValueRVLen := 0

			if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
				// На вход получаем только указатели на slice
				if compositionFieldValueRV.Kind() != reflect.Slice {
					return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - marshal error - must be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
				} else {
					compositionFieldValueRVLen = compositionFieldValueRV.Len() // Ожидаем на вход много строк

					// Создадим slice под список полей - массив строк
					if compositionRows, err = s.newSliceRestrict(requestID, toEntity, optionsRef, 0, compositionFieldValueRVLen); err != nil {
						return nil, nil, err, errors
					}
				}
			} else {
				// На вход НЕ получаем только указатели на slice
				if compositionFieldValueRV.Kind() == reflect.Slice {
					return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - marshal error - must NOT be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
				} else {
					compositionFieldValueRVLen = 1 // Ожидаем на вход одну строку
				}
			}

			// Обработаем все строки во входной структуре - или это одна строка
			for i := 0; i < compositionFieldValueRVLen; i++ {

				rowErrors := _err.Errors{}
				compositionSourceRowRV := reflect.Value{}

				if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
					compositionSourceRowRV = compositionFieldValueRV.Index(i) // текущий элемент slice - map[string]interface{} в виде interface {}
				} else {
					compositionSourceRowRV = compositionFieldValueRV // map[string]interface{} в виде interface {}
				}

				// Создадим структуру под список полей
				compositionDestRow, err := s.newRowRestrict(requestID, toEntity, optionsRef)
				if err != nil {
					return nil, nil, err, errors
				}

				{ // Два режима работы, либо получаем на вход *struct{}, либо *map[string]interface{}
					if compositionSourceRowRV.Kind() == reflect.Ptr && reflect.Indirect(compositionSourceRowRV).Kind() == reflect.Struct {
						// TODO - вложенные *struct{} нужно разобрать в Object и выстроить association и composition
						// Если получили *struct{}, то можем брать ее напрямую - конвертация не нужна
						compositionDestRow.SetFromRV(compositionSourceRowRV)

					} else {
						// Ожидаем получить на вход *map[string]interface{} - скопируем только те поля, которые нужные в ответе
						if _, err = toEntity.CopyMapToStruct(compositionSourceRowRV.Interface(), compositionDestRow, optionsRef.Fields); err != nil {
							rowErrors.Append(requestID, err)
						}
					}
				} // Два режима работы, либо получаем на вход *struct{}, либо *map[string]interface{}

				{ // Добавляем в массив объектов - даже если были ошибки
					if compositionDestRow != nil {
						if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
							compositionRowsRV := reflect.Indirect(compositionRows.RV)                       // Собственно целевой slice, в который будем вставлять *[]struct{} -> []struct{}
							compositionRowsRV.Set(reflect.Append(compositionRowsRV, compositionDestRow.RV)) // Успешная обработка - вставляем в slice *struct{}
						} else {
							compositionRows = compositionDestRow // Обрабатываем только одну строку
						}
						compositionRows.AppendObject(compositionDestRow) // Добавляем в массив объектов
					}
				} // Добавляем в массив объектов - даже если были ошибки

				// Вычисляемые поля перед рекурсивной обработкой, чтобы они попали в подчиненные объекты
				if calculate {
					if compositionDestRow != nil {
						if !compositionDestRow.Options.Global.SkipCalculation {
							if err = s.processExprs(ctx, requestID, nil, compositionDestRow, _meta.EXPR_ACTION_PRE_MARSHAL); err != nil {
								rowErrors.Append(requestID, err)
							}
						}
					}
				}

				{ // Рекурсивная обработка каскадно
					if compositionDestRow != nil {
						localErrors := _err.Errors{} // локальные ошибки вложенного метода
						if err, localErrors = s.processMarshal(ctx, requestID, compositionDestRow, action, exprAction, cascadeUp, cascadeDown, validate, calculate, false); err != nil {
							rowErrors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
						}
						//innerErrors.AppendErrors(localErrors)
						rowErrors.AppendErrors(localErrors)
					}
				} // Рекурсивная обработка каскадно

				_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
				err = s.processErrors(requestID, compositionDestRow, rowErrors, rowOut.Options.Global.EmbedError, "Composition marshal")
				errors.Append(requestID, err)
				innerErrors.AppendErrors(rowErrors)
			}

			if errors.HasError() {
				_log.Debug("ERROR: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - marshal error", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name)), errors
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return compositionRows, keyArgs, nil, errors
			}

		} else {
			// Данных нет, обрабатывать нечего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && compositionField != nil {}", []interface{}{s, rowOut, compositionField}).PrintfError(), errors
}
