package meta

import (
	"fmt"
	"reflect"

	"github.com/antonmedv/expr/vm"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

func (ex *Expr) CopyField(externalId uint64, rowIn *Object, rowOut *Object, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil && rowIn != nil && rowIn.RV.IsValid() && rowOut != nil && rowOut.RV.IsValid() {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, externalId, "CopyField", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", ex.entity.Name, ex.Name, ex.Code))
			}
		}()

		// На вход получаем только указатели на struct
		if rowIn.RV.Kind() != reflect.Ptr || reflect.Indirect(rowIn.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "rowIn.Value must be pointer to struct", []interface{}{rowIn.RV.Kind().String(), reflect.Indirect(rowIn.RV).Kind().String()}).PrintfError()
		}

		// На вход получаем только указатели на struct
		if rowOut.RV.Kind() != reflect.Ptr || reflect.Indirect(rowOut.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "rowOut.Value must be pointer to struct", []interface{}{rowOut.RV.Kind().String(), reflect.Indirect(rowOut.RV).Kind().String()}).PrintfError()
		}

		if ex.field == nil {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - empty field pointer", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		}

		if ex.Type != EXPR_COPY {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.entity.Name, ex.field.Name, ex.Code, ex.Name, ex.Type, EXPR_COPY))
		}

		args := make(map[string]interface{})

		// Сформируем набор значений полей для вычисления из исходной структуры
		for _, argsField := range ex.argsFields {
			if exprFieldRV, err := rowIn.FieldRV(argsField); err != nil {
				return nil, err
			} else {
				// В выражении может быть указано не имя поля, а его таг 'expr'
				args[argsField.GetTag("expr", true)] = exprFieldRV.Interface()
			}
		}

		// вызовем расчет
		if output, err = ex.Run(externalId, args, exprVm); err != nil {
			return nil, err
		}

		// Установим поле по результатам расчета
		if err = rowOut.SetFieldRV(ex.field, reflect.Indirect(reflect.ValueOf(output))); err != nil {
			return nil, err
		}

		return output, nil

	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && rowIn != nil && rowIn.Value.IsValid() && && rowOut != nil && rowOut.Value.IsValid() {}", []interface{}{ex, rowIn, rowOut}).PrintfError()
}

func (ex *Expr) Convert(externalId uint64, rowIn *Object, rowOut *Object, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil && rowIn != nil && rowIn.RV.IsValid() && rowOut != nil && rowOut.RV.IsValid() {

		// На вход получаем только указатели на struct
		if rowIn.RV.Kind() != reflect.Ptr || reflect.Indirect(rowIn.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "rowIn.Value must be pointer to struct", []interface{}{rowIn.RV.Kind().String(), reflect.Indirect(rowIn.RV).Kind().String()}).PrintfError()
		}

		// На вход получаем только указатели на struct
		if rowOut.RV.Kind() != reflect.Ptr || reflect.Indirect(rowOut.RV).Kind() != reflect.Struct {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "rowOut.Value must be pointer to struct", []interface{}{rowOut.RV.Kind().String(), reflect.Indirect(rowOut.RV).Kind().String()}).PrintfError()
		}

		if ex.field == nil {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - empty field pointer", ex.entity.Name, ex.Name, ex.Code))
		}

		if ex.Type != EXPR_CONVERT {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be 'Convert'", ex.entity.Name, ex.field.Name, ex.Code, ex.Name, ex.Type))
		}

		env := make(map[string]interface{})

		// Сформируем набор значений полей для вычисления из исходной структуры
		for _, exprField := range ex.argsFields {
			if exprFieldRV, err := rowIn.FieldRV(exprField); err != nil {
				return nil, err
			} else {
				// В выражении может быть указано не имя поля, а его таг 'expr'
				env[exprField.GetTag("expr", true)] = exprFieldRV.Interface()
			}
		}

		// вызовем расчет
		if output, err = ex.Run(externalId, env, exprVm); err != nil {
			return nil, err
		}

		// Установим поле по результатам расчета
		if err = rowOut.SetFieldRV(ex.field, reflect.Indirect(reflect.ValueOf(output))); err != nil {
			return nil, err
		}

		return output, nil

	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && rowIn != nil && rowIn.Value.IsValid() && && rowOut != nil && rowOut.Value.IsValid() {}", []interface{}{ex, rowIn, rowOut}).PrintfError()
}

//func (ex *Expr) CopyPtr(externalId uint64, structPtrInValue reflect.Value, structPtrOutValue reflect.Value) (output interface{}, err error) {
//	if ex != nil && ex.entity != nil && ex.program != nil && structPtrInValue.IsValid() && structPtrOutValue.IsValid() {
//
//		// На вход получаем только указатели на struct
//		if structPtrOutValue.Kind() != reflect.Ptr || reflect.Indirect(structPtrOutValue).Kind() != reflect.Struct {
//			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "structPtrOutValue must be pointer to struct", []interface{}{structPtrOutValue.Kind().String(), reflect.Indirect(structPtrOutValue).Kind().String()}).PrintfError()
//		}
//
//		// На вход получаем только указатели на struct
//		if structPtrInValue.Kind() != reflect.Ptr || reflect.Indirect(structPtrOutValue).Kind() != reflect.Struct {
//			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "structPtrOutValue must be pointer to struct", []interface{}{structPtrOutValue.Kind().String(), reflect.Indirect(structPtrOutValue).Kind().String()}).PrintfError()
//		}
//
//		if ex.field == nil {
//			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' - empty field pointer", ex.entity.Name, ex.Name))
//		}
//
//		if ex.Type != EXPR_COPY {
//			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' - incorrect 'type' '%s'. Must be 'CopyField'", ex.entity.Name, ex.field.Name, ex.Name, ex.Type))
//		}
//
//		env := make(map[string]interface{})
//
//		// Сформируем набор значений полей для вычисления
//		for _, exprField := range ex.fields {
//			if exprFieldValue, err := ex.entity.FieldValuePtr(exprField, structPtrInValue); err != nil {
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
//	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && structPtrInValue.IsValid() && structPtrOutValue.IsValid() {}", []interface{}{ex}).PrintfError()
//}
