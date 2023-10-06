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

// processPrePersist - в ходе разбора сообщения marshaling
func (s *Service) processMarshal(ctx context.Context, requestID uint64, row *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (err error, errors _err.Errors) {
	if s != nil && row != nil {

		//tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, row.EntityName", requestID, row.Entity.Name)

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Вычисляемые поля
		if opt.calculate {
			if !row.Options.Global.SkipCalculation {
				if err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_PRE_MARSHAL); err != nil {
					errors.Append(requestID, err)
				}
			}
		}

		// Обработать FK M:1 - при отрицательном cascadeUp - обрабатываем все уровни без ограничений
		if row.Entity.HasAssociations() {
			if cascadeUp != 0 {
				//  уровень уменьшаем на 1, 0 - означает только себя
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				// У Association не должно быть своих Composition
				if err, localErrors = s.processAssociations(ctx, requestID, nil, row, action, exprAction, cascadeUp-1, 0, opt, s.marshalAssociation); err != nil {
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

		// Обработать FK 1:M - при отрицательном cascadeDown - обрабатываем все уровни без ограничений
		if row.Entity.HasCompositions() {
			if cascadeDown != 0 {
				//  уровень уменьшаем на 1, 0 - означает только себя
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processCompositions(ctx, requestID, nil, row, action, exprAction, cascadeUp, cascadeDown-1, opt, s.marshalComposition); err != nil {
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

		//_log.Debug("END: requestID, EntityName, duration, errors", requestID, row.Entity.Name, time.Now().Sub(tic), len(errors))
		err = s.processErrors(requestID, row, errors, row.Options.Global.EmbedError, "Marshal")
		return err, errors
	}

	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError(), errors
}

func (s *Service) marshalAssociation(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, associationField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (associationRow *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && associationField != nil {

		//tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := associationField.Reference()
		toEntity := reference.ToEntity()
		//toKey := reference.ToKey()

		//_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

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
				err = _recover.GetRecoverError(r, requestID, "prePersistAssociation", rowOut.Entity.Name)
			}
		}()

		// найдем значение поля, в котором хранятся входные данные, они имеют формат произвольного интерфейса
		associationFieldRV, err := rowOut.FieldRV(associationField)
		if err != nil {
			return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Association field='%s' value by in index", rowOut.Entity.Name, associationField.Name)), errors
		}

		// Проверим, что есть данные для копирования
		if associationFieldRV.IsValid() && !associationFieldRV.IsNil() && !associationFieldRV.IsZero() {

			optionsRef, err := s.ParseQueryOptions(ctx, requestID, rowOut.Options.Key+".association."+toEntity.Name, toEntity, associationField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
			if err != nil {
				return nil, nil, err, errors
			}

			// Сформировать struct из map[string]interface{} и скопировать только нужные поля
			if associationRow, err = s.MarshalMap(requestID, toEntity, optionsRef, associationFieldRV.Interface(), opt.persistRestrictFields, opt.addRefFields, opt.addUKFields); err != nil {
				errors.Append(requestID, err) // накапливаем ошибки
			}

			{ // Рекурсивная обработка каскадно
				if associationRow != nil {
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if err, localErrors = s.processMarshal(ctx, requestID, associationRow, action, exprAction, cascadeUp, cascadeDown, opt); err != nil {
						errors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
					}
					innerErrors.AppendErrors(localErrors)
				}
			} // Рекурсивная обработка каскадно

			//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, associationRow, errors, rowOut.Options.Global.EmbedError, "Association MarshalEntity")
			return associationRow, keyArgs, err, errors

		} else {
			// Данных нет, обрабатывать не чего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil  && associationField != nil {}", []interface{}{s, rowOut, associationField}).PrintfError(), errors
}

func (s *Service) marshalComposition(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (compositionRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && compositionField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := compositionField.Reference()
		toEntity := reference.ToEntity()
		//toKey := reference.ToKey()

		//_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

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
				err = _recover.GetRecoverError(r, requestID, "prePersistComposition", rowOut.Entity.Name)
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

			optionsRef, err := s.ParseQueryOptions(ctx, requestID, rowOut.Options.Key+".composition."+toEntity.Name, toEntity, compositionField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
			if err != nil {
				return nil, nil, err, errors
			}

			compositionFieldValueRVLen := 0

			if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
				// На вход получаем только указатели на slice
				if compositionFieldValueRV.Kind() != reflect.Slice {
					return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - MarshalEntity error - must be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
				} else {
					compositionFieldValueRVLen = compositionFieldValueRV.Len() // Ожидаем на вход много строк

					// Создадим slice []interface{}, так как на вход могли прийти разные структуры
					if compositionRows, err = s.newSliceAnyRestrict(requestID, toEntity, optionsRef, 0, compositionFieldValueRVLen); err != nil {
						return nil, nil, err, errors
					}
				}
			} else {
				// На вход НЕ получаем указатели на slice
				if compositionFieldValueRV.Kind() == reflect.Slice {
					return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - MarshalEntity error - must NOT be an array for cardinality='%s'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.Cardinality)), errors
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

				// Сформировать struct из map[string]interface{} и скопировать только нужные поля
				compositionDestRow, err := s.MarshalMap(requestID, toEntity, optionsRef, compositionSourceRowRV.Interface(), opt.persistRestrictFields, opt.addRefFields, opt.addUKFields)
				if err != nil {
					errors.Append(requestID, err) // накапливаем ошибки
				}

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

				{ // Рекурсивная обработка каскадно
					if compositionDestRow != nil {
						localErrors := _err.Errors{} // локальные ошибки вложенного метода
						if err, localErrors = s.processMarshal(ctx, requestID, compositionDestRow, action, exprAction, cascadeUp, cascadeDown, opt); err != nil {
							rowErrors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
						}
						//innerErrors.AppendErrors(localErrors)
						rowErrors.AppendErrors(localErrors)
					}
				} // Рекурсивная обработка каскадно

				//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
				err = s.processErrors(requestID, compositionDestRow, rowErrors, rowOut.Options.Global.EmbedError, "Composition MarshalEntity")
				errors.Append(requestID, err)
				innerErrors.AppendErrors(rowErrors)
			}

			if errors.HasError() {
				_log.Debug("ERROR: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - MarshalEntity error", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name)), errors
			} else {
				//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return compositionRows, keyArgs, nil, errors
			}

		} else {
			// Данных нет, обрабатывать нечего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && compositionField != nil {}", []interface{}{s, rowOut, compositionField}).PrintfError(), errors
}

// MarshalMap - сформировать Object из map[string]interface{}
func (s *Service) MarshalMap(requestID uint64, entity *_meta.Entity, options *_meta.Options, rowInI interface{}, restrictOutFields bool, addRefFields bool, addUKFields bool) (rowOut *_meta.Object, err error) {
	if s != nil && entity != nil && options != nil && rowInI != nil {

		var mapIn map[string]interface{}
		var ok bool
		var inFieldsName []string
		var outFields _meta.FieldsMap

		// На вход получаем map[string]interface{} или *map[string]interface{}
		if reflect.TypeOf(rowInI).Kind() != reflect.Ptr {
			mapIn, ok = rowInI.(map[string]interface{})
			if !ok {
				return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - MarshalEntity error - must be *map[string]interface{} got '%s'", entity.Name, reflect.TypeOf(rowInI).Kind().String()))
			}
		} else {
			fromMapPtr, ok := rowInI.(*map[string]interface{})
			if !ok {
				return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - MarshalEntity error - must be map[string]interface{} got '%s'", entity.Name, reflect.TypeOf(rowInI).Kind().String()))
			}
			mapIn = *fromMapPtr
		}

		if restrictOutFields {
			// Список полей для копирования
			for fieldName, _ := range mapIn {
				inFieldsName = append(inFieldsName, fieldName)
			}

			if len(inFieldsName) > 0 {
				// Подготовим map с полями, которые нужны
				outFields, err = s.constructFieldsMap(requestID, entity, inFieldsName, options.Global.InFormat, options.Global.IgnoreExtraField, addRefFields, addUKFields)
				if err != nil {
					return nil, err
				}
			} else {
				// на входе нет данных - пустой объект не создаем
				return nil, nil
			}

			// Ограничить поля выходной структуры, тем, что в map - используется для update
			localOptions := options.Clone()
			localOptions.Fields = outFields

			// Создадим структуру под пришедший на вход список полей
			rowOut, err = s.newRowRestrict(requestID, entity, localOptions)
			if err != nil {
				return nil, err
			}
		} else {
			// Создадим структуру под полный список полей
			rowOut, err = s.NewRowAll(requestID, entity, options)
			if err != nil {
				return nil, err
			}
		}

		// Скопировать только нужные поля
		if _, err = entity.CopyMapToStruct(mapIn, rowOut, outFields); err != nil {
			return nil, err
		} else {
			// Очистим map - больше она не нужна
			clear(mapIn)
		}

		return rowOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && options != nil && rowInI != nil {}", []interface{}{s, entity, options, rowInI}).PrintfError()
}
