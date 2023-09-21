package meta

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antonmedv/expr/vm"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

func (ex *Expr) CalculateField(externalId uint64, row *Object, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil && row != nil && row.RV.IsValid() {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, externalId, "CalculateField", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", ex.entity.Name, ex.Name, ex.Code))
			}
		}()

		if !ex.isInit {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] is not init", ex.entity.Name, ex.field.Name, ex.Name, ex.Code)).PrintfError()
		}

		// На вход получаем только указатели на struct
		if row.RV.Kind() != reflect.Ptr || reflect.Indirect(row.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "row.Value must be pointer to struct", []interface{}{row.RV.Kind().String(), reflect.Indirect(row.RV).Kind().String()}).PrintfError()
		}

		if ex.field == nil {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - empty field pointer", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		}

		if ex.Type != EXPR_CALCULATE {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.entity.Name, ex.field.Name, ex.Name, ex.Code, ex.Type, EXPR_CALCULATE))
		}

		args := make(map[string]interface{}, len(ex.argsFields))

		// Сформируем набор значений полей для вычисления
		for _, argsField := range ex.argsFields {
			if fieldValue, err := row.FieldValue(argsField); err != nil {
				return nil, err
			} else {
				// В выражении может быть указано не имя поля, а его таг 'expr'
				args[argsField.GetTag("expr", true)] = fieldValue
			}
		}

		// вызовем расчет
		if output, err = ex.Run(externalId, args, exprVm); err != nil {
			//if output, err = ex.Run(externalId, row.Value, exprVm); err != nil { // снижает производительность в 5 раз
			return nil, err
		}

		// TODO пустое вычисление может быть нормальной ситуацией
		if output == nil {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - empty calculation result, fields ['%s'], values ['%s']", ex.entity.Name, ex.field.Name, ex.Name, ex.Code, strings.Join(ex.FieldsArgsName, "','"), ArgsMapToStrings("','", args)))
		}

		// Установим поле по результатам расчета
		if err = row.SetFieldRV(ex.field, reflect.ValueOf(output)); err != nil {
			return nil, err
		}

		return output, nil

	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && row != nil && row.Value.IsValid() {}", []interface{}{ex, row}).PrintfError()
}

//func (ex *Expr) CalculatePtr(externalId uint64, structPtrOutValue reflect.Value) (output interface{}, err error) {
//	if ex != nil && ex.entity != nil && ex.program != nil && structPtrOutValue.IsValid() {
//
//		// На вход получаем только указатели на struct
//		if structPtrOutValue.Kind() != reflect.Ptr || reflect.Indirect(structPtrOutValue).Kind() != reflect.Struct {
//			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "structPtrOutValue must be pointer to struct", []interface{}{structPtrOutValue.Kind().String(), reflect.Indirect(structPtrOutValue).Kind().String()}).PrintfError()
//		}
//
//		if ex.field == nil {
//			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' - empty field pointer", ex.entity.Name, ex.Name))
//		}
//
//		if ex.Type != EXPR_CALCULATE {
//			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' - incorrect 'type' '%s'. Must be 'CalculateField'", ex.entity.Name, ex.field.Name, ex.Name, ex.Type))
//		}
//
//		env := make(map[string]interface{})
//
//		// Сформируем набор значений полей для вычисления
//		for _, exprField := range ex.fields {
//			if exprFieldValue, err := ex.entity.FieldValuePtr(exprField, structPtrOutValue); err != nil {
//				return nil, err
//			} else {
//				// В выражении может быть указано не имя поля, а его таг 'expr'
//				env[exprField.GetTag("expr", true)] = exprFieldValue.Interface()
//			}
//		}
//
//		// вызовем расчет
//		if output, err = ex.Run(externalId, env); err != nil {
//			return nil, err
//		}
//
//		// Установим поле по результатам расчета
//		if err = ex.entity.SetFieldValuePtr(ex.field, structPtrOutValue, reflect.Indirect(reflect.ValueOf(output))); err != nil {
//			return nil, err
//		}
//
//		return output, nil
//
//	}
//	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && structPtrOutValue.IsValid() {}", []interface{}{ex}).PrintfError()
//}
//
