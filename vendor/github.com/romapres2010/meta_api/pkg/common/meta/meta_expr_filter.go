package meta

import (
	"fmt"
	"github.com/antonmedv/expr/vm"
	"reflect"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

func (ex *Expr) FilterSliceFromStructField(externalId uint64, row *Object, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil && row != nil && row.RV.IsValid() {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, externalId, "Filter", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", ex.entity.Name, ex.Name, ex.Code))
			}
		}()

		if !ex.isInit {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - is not init", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		}

		// На вход получаем только указатели на struct
		if row.RV.Kind() != reflect.Ptr || reflect.Indirect(row.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "row.Value must be pointer to struct", []interface{}{row.RV.Kind().String(), reflect.Indirect(row.RV).Kind().String()}).PrintfError()
		}

		if ex.Type != EXPR_FILTER {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.entity.Name, ex.Name, ex.Code, ex.Type, EXPR_FILTER))
		}

		env := make(map[string]interface{})

		// Сформируем набор значений полей для вычисления
		for _, exprField := range ex.argsFields {
			if exprFieldRV, err := row.FieldRV(exprField); err != nil {
				return nil, err
			} else {

				if exprFieldRV.IsValid() && !exprFieldRV.IsZero() {

					exprFieldValueIndirect := reflect.Indirect(exprFieldRV) // interface{}

					if exprFieldValueIndirect.IsValid() && !exprFieldValueIndirect.IsZero() {
						exprSliceRV := reflect.Indirect(reflect.ValueOf(exprFieldValueIndirect.Interface())) // []struct

						if exprSliceRV.Kind() != reflect.Slice {
							return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - incorrect field '%s' type '%s'. Must be 'Slice'", ex.entity.Name, ex.Name, ex.Code, exprSliceRV.Kind().String(), exprField.Name)).PrintfError()
						}

						if exprSliceRV.IsValid() && !exprSliceRV.IsZero() {
							env[exprField.GetTag("expr", true)] = exprSliceRV.Interface() // []struct
						}
					}
				}
			}
		}

		if len(env) > 0 { // на пустом массиве запускать нельзя

			// вызовем расчет
			if output, err = ex.Run(externalId, env, exprVm); err != nil {
				return nil, err
			}

			// Пустой output - допустимая ситуация для фильтрации
			return output, nil
		}

		return nil, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && row != nil && row.Value.IsValid() {}", []interface{}{ex, row}).PrintfError()
}

func (ex *Expr) FilterSlice(externalId uint64, rows *Object, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil && rows != nil && rows.RV.IsValid() {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, externalId, "Filter", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", ex.entity.Name, ex.Name, ex.Code))
			}
		}()

		if !ex.isInit {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - is not init", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		}

		if !rows.IsSlice {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - can be aplicabe only to slice object", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		}

		// На вход получаем только указатели на Slice
		if rows.RV.Kind() != reflect.Ptr || reflect.Indirect(rows.RV).Kind() != reflect.Slice {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "rows.Value must be pointer to slice", []interface{}{rows.RV.Kind().String(), reflect.Indirect(rows.RV).Kind().String()}).PrintfError()
		}

		if ex.Type != EXPR_FILTER {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s 'Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.entity.Name, ex.Name, ex.Code, ex.Type, EXPR_FILTER)).PrintfError()
		}

		env := make(map[string]interface{})

		// rows.Value содержит *[]interface{}
		if rows.Value != nil {
			rowsRV := reflect.Indirect(reflect.ValueOf(rows.Value))

			if rowsRV.IsValid() && !rowsRV.IsZero() {
				// Имя элемента фиксированное = Entity.Name
				env[rows.Entity.Name] = rowsRV.Interface() // []struct
			}
		}

		if len(env) > 0 { // на пустом массиве запускать нельзя

			// вызовем расчет
			if output, err = ex.Run(externalId, env, exprVm); err != nil {
				return nil, err
			}

			// Пустой output - допустимая ситуация для фильтрации
			return output, nil
		}

		return nil, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && rows != nil && rows.Value.IsValid() {}", []interface{}{ex, rows}).PrintfError()
}
