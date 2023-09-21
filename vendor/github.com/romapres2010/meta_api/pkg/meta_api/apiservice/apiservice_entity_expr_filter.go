package apiservice

import (
	"fmt"
	"reflect"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// processCompositionFiltering - отфильтровать composition и вернуть новую composition
func (s *Service) processCompositionFiltering(requestID uint64, expr *_meta.Expr, row *_meta.Object, compositionRowsIn *_meta.Object, composition *_meta.Reference) (rowsOut *_meta.Object, err error) {
	if s != nil && row != nil && row.Entity != nil && compositionRowsIn != nil && compositionRowsIn.IsSlice && composition != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, row.Entity.Name)

		if expr.Type == _meta.EXPR_FILTER {

			// Функция восстановления после паники в reflect
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, requestID, "processCompositionFiltering", row.Entity.Name)
				}
			}()

			// должно быть поле - composition, которое фильтруем
			compositionField := expr.Field()
			if compositionField == nil {
				return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Expression '%s' Code ['%s'] - empty Composition field to filtering to", row.Entity.Name, expr.Name, expr.Code)).PrintfError()
			}

			// Найдем значение поля, в которое поместить новый отфильтрованный slice
			compositionFieldRV, err := row.FieldRV(compositionField)
			if err != nil {
				return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Composition field='%s' value by in index", row.Entity.Name, compositionField.Name))
			}

			// Выполним фильтрацию Composition
			if rowsFiltered, err := expr.FilterSliceFromStructField(requestID, row, nil); err != nil {
				// TODO - если ошибка фильтрации, то это равноценно пустой фильтрации? Очищать Composition
				compositionFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_COMPOSITION_RT)) // Встроим пустой *interface{}
				return nil, err
			} else {
				_log.Debug("Success filter: entityName, exprName", row.Entity.Name, expr.Name)

				if rowsFiltered != nil {

					// Отфильтрованные строки, из которого будем считывать данные
					rowsFilteredRV := reflect.ValueOf(rowsFiltered)

					// Количество отфильтрованных строк больше 0
					if rowsFilteredRVLen := rowsFilteredRV.Len(); rowsFilteredRVLen > 0 {

						// Создать новый Slice и перенести в него данные
						if rowsOut, err = s.constructFilteredSlice(requestID, rowsFiltered, compositionRowsIn); err != nil {
							return nil, err
						}

						// Если типы можно присваивать
						if reflect.TypeOf(rowsOut.RV).AssignableTo(compositionField.ReflectType()) {
							compositionFieldRV.Set(rowsOut.RV) // Встроим указатель на rowsOut в структуру
						} else {
							return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' filtering CompositionField='%s' - incompatible Struct type  CompositionType='%s', compositionRowsType='%s'", row.Entity.Name, compositionField.Name, compositionField.ReflectType().String(), reflect.TypeOf(rowsOut.Value).String())).PrintfError()
						}

						_log.Debug("SUCCESS FILTERING: requestID, entityName, rowsFiltered, duration", requestID, row.Entity.Name, rowsFilteredRVLen, time.Now().Sub(tic))
						return rowsOut, nil

					} else {
						// Пустой набор после фильтрации
						compositionFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_COMPOSITION_RT)) // Встроим пустой *interface{}

						_log.Debug("EMPTY FILTERING: requestID, entityName, duration", requestID, row.Entity.Name, time.Now().Sub(tic))
						return nil, nil
					}
				}
			}
		}

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, row.Entity.Name, time.Now().Sub(tic))

		return rowsOut, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil  && row != nil && row.Entity != nil && compositionRowsIn != nil && compositionRowsIn.IsSlice && composition != nil {}", []interface{}{s, row, compositionRowsIn, composition}).PrintfError()
}

// processRowsFiltering - отфильтровать и вернуть новый набор строк
func (s *Service) processRowsFiltering(requestID uint64, expr *_meta.Expr, rowsIn *_meta.Object) (rowsOut *_meta.Object, err error) {
	if s != nil && rowsIn != nil && rowsIn.Entity != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName", requestID, rowsIn.Entity.Name)

		if expr.Type == _meta.EXPR_FILTER {

			// Функция восстановления после паники в reflect
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, requestID, "processRowsFiltering", rowsIn.Entity.Name)
				}
			}()

			// Выполним фильтрацию
			if rowsFiltered, err := expr.FilterSlice(requestID, rowsIn, nil); err != nil {
				return nil, err
				//return nil, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' Expression '%s' Code ['%s'] - filtering error", rowsIn.Entity.Name, expr.Name, expr.Code))
			} else {
				_log.Debug("Success filter: entityName, exprName, action", rowsIn.Entity.Name, expr.Name)

				if rowsFiltered != nil {
					// Создать новый Slice и перенести в него данные
					if rowsOut, err = s.constructFilteredSlice(requestID, rowsFiltered, rowsIn); err != nil {
						return nil, err
					}
					_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowsIn.Entity.Name, time.Now().Sub(tic))
					return rowsOut, nil
				}
			}
		}

		return nil, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowsIn != nil && rowsIn.Entity != nil {}", []interface{}{s, rowsIn}).PrintfError()
}

// filterSlice - вернуть новый набор строк по отфильтрованным
func (s *Service) constructFilteredSlice(requestID uint64, rowsFiltered interface{}, rowsIn *_meta.Object) (rowsOut *_meta.Object, err error) {
	if s != nil && rowsIn != nil && rowsIn.Entity != nil {

		tic := time.Now()

		_log.Debug("START: entityName", requestID, rowsIn.Entity.Name)

		if rowsFiltered != nil {

			// Функция восстановления после паники в reflect
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, requestID, "processRowsFiltering", rowsIn.Entity.Name)
				}
			}()

			// Отфильтрованные строки, из которого будем считывать данные []interface{}
			rowsFilteredRV := reflect.ValueOf(rowsFiltered)

			// Количество отфильтрованных строк больше 0
			if rowsFilteredRVLen := rowsFilteredRV.Len(); rowsFilteredRVLen > 0 {

				// Создадим новый rowsOut под тот же тип данных и набор полей
				rowsOut, err = s.newSliceRestrict(requestID, rowsIn.Entity, rowsIn.Options, 0, rowsFilteredRVLen)
				if err != nil {
					return nil, err
				}
				rowsOut.CopyAssociationFrom(rowsIn) // Скопировать все Association
				rowsOut.CopyCompositionFrom(rowsIn) // Скопировать все Composition

				// Собственно rowsOut, в который будем вставлять *struct
				rowsOutRV := reflect.Indirect(rowsOut.RV)

				// Обработаем все строки в отфильтрованном []interface{}
				for i := 0; i < rowsFilteredRVLen; i++ {
					filteredRow := rowsFilteredRV.Index(i).Interface() // Получаем одну структуру в виде interface{}
					filteredRowRV := reflect.ValueOf(filteredRow)      // Текущий элемент массива в виде []struct

					// Если фильтрация вернула ту же структуру, то можем взять ее напрямую
					if filteredRowRV.CanConvert(rowsOut.StructPtrType) {
						rowsOutRV.Set(reflect.Append(rowsOutRV, filteredRowRV)) // добавляем []struct в slice rowsOut
					} else {
						return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' filtering - incompatible Struct type ExpectedType='%s', FilteringType='%s'", rowsIn.Entity.Name, rowsOut.StructType.String(), reflect.TypeOf(filteredRow).String())).PrintfError()
					}
				}

				// Второй проход по наполненному массиву для формирования объектов с указателями на сохраненные в Slice struct
				// TODO - второй проход не нужен
				rowsOutRVLen := rowsOutRV.Len()
				for i := 0; i < rowsOutRVLen; i++ {
					//rowOutPtrRV := rowsOutRV.Index(i).Addr() // указатель на текущую struct
					rowOutPtrRV := rowsOutRV.Index(i) // указатель на текущую struct
					rowOut := rowsOut.NewFromRV(rowOutPtrRV, false)
					rowsOut.AppendObject(rowOut)
				}

				_log.Debug("SUCCESS FILTERING: requestID, entityName, inRows, rowsFiltered, duration", requestID, rowsIn.Entity.Name, len(rowsIn.Objects), rowsFilteredRVLen, time.Now().Sub(tic))
				return rowsOut, nil
			}
		} else {
			_log.Debug("EMPTY FILTERING: requestID, entityName, inRows, rowsFiltered, duration", requestID, rowsIn.Entity.Name, len(rowsIn.Objects), 0, time.Now().Sub(tic))
			return nil, nil
		}

		return nil, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowsIn != nil && rowsIn.Entity != nil {}", []interface{}{s, rowsIn}).PrintfError()
}
