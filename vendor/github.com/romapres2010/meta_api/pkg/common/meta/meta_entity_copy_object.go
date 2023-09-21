package meta

import (
	"fmt"
	"reflect"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// CopyObjectStruct Копирование struct через его указатель по совпадающим полям в структуре
func (entity *Entity) CopyObjectStruct(fromStruct *Object, toStruct *Object, outFields FieldsMap) (err error) {
	if fromStruct != nil && toStruct != nil {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CopyObjectStruct", entity.Name)
			}
		}()

		if entity.Modify.CopyRestrict {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - forbiden to copy", entity.Name)).PrintfError()
		}

		fromPtrValue := fromStruct.RV
		toPtrValue := toStruct.RV

		//_log.Debug("fromPtrValue.Kind(), toPtrValue.Kind()", fromPtrValue.Kind(), toPtrValue.Kind())
		// На вход получаем только указатели
		if fromPtrValue.Kind() != reflect.Ptr && toPtrValue.Kind() != reflect.Ptr {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromPtrValue.Kind() == reflect.Val && toPtrValue.Kind() == reflect.Val {}", []interface{}{fromPtrValue.Kind().String(), toPtrValue.Kind().String()}).PrintfError()
		}

		// Собственно структуры
		fromRow := reflect.Indirect(fromPtrValue)
		toRow := reflect.Indirect(toPtrValue)
		//_log.Debug("fromRow.Kind(), toRow.Kind()", fromRow.Kind(), toRow.Kind())
		// На вход получаем только указатели на struct
		if fromRow.Kind() != reflect.Struct || toRow.Kind() != reflect.Struct {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromRow.Kind() != reflect.Struct || toRow.Kind() != reflect.Struct {}", []interface{}{fromRow.Kind().String(), toRow.Kind().String()}).PrintfError()
		}

		// проверим, совпадают ли типы структур
		//fromRowType := reflect.TypeOf(fromRow.Interface())
		//toRowType := reflect.TypeOf(toRow.Interface())
		fromRowType := fromStruct.StructType
		toRowType := toStruct.StructType
		//_log.Debug("toRowType.Kind(), fromRowType.Kind()", toRowType.Kind(), fromRowType.Kind())

		//if toRowType == fromRowType {
		if fromRowType.AssignableTo(toRowType) {
			//_log.Debug("Structs has the same type. Do not need to copy")
			toRow.Set(fromRow)
		} else {
			//_log.Debug("Structs was different. Need to copy")
			if len(outFields) == 0 {
				// структуры не совпадают делаем копирование по совпадению полей и их типов
				toRowNumField := toRow.NumField()
				for i := 0; i < toRowNumField; i++ {
					toFieldRV := toRow.Field(i)
					if toFieldRV.IsValid() {
						toFieldName := toRowType.Field(i).Name
						fromFieldRV := fromRow.FieldByName(toFieldName)
						if fromFieldRV.IsValid() {

							// Пустые интерфейсы не копируем
							if fromFieldRV.Kind() == reflect.Interface && fromFieldRV.IsNil() {
								continue
							}

							// Если типы можно присваивать
							//if reflect.TypeOf(toFieldRV.Interface()) == reflect.TypeOf(fromFieldRV.Interface()) { // Если типы точно совпадают
							if reflect.TypeOf(fromFieldRV).AssignableTo(reflect.TypeOf(toFieldRV)) {
								toFieldRV.Set(fromFieldRV)
							} else {
								return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - deep struct copy error: Incompatible Struct fields type, inFieldType='%s', outFieldType='%s'", entity.Name, toFieldName, fromFieldRV.String(), toFieldRV.String())).PrintfError()
								//return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - deep struct copy error: Incompatible Struct fields type, inFieldType='%s', outFieldType='%s'", entity.Name, toFieldName, reflect.TypeOf(fromFieldRV.Interface()).String(), reflect.TypeOf(toFieldRV.Interface()).String())).PrintfError()
							}
						}
					}
				}
			} else {
				// Копируем по конкретным именам полей и их индексам в структурах
				for _, field := range outFields {

					// Признак, запрета на копирования поля
					if field.Modify.CopyRestrict {
						continue
					}

					// поля, которые запрещено извлекать не копируем
					if field.Modify.CopyRestrict {
						_log.Debug("field forbidden to copy", field.Name)
						continue
					}

					toFieldIndex := field.indexMap[toRowType]
					if toFieldIndex == nil {
						return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - deep struct copy error: destination field index does not exists", entity.Name, field.Name)).PrintfError()
					}
					toField, errInner := toRow.FieldByIndexErr(toFieldIndex)
					if errInner != nil {
						return errInner
					}

					fromFieldIndex := field.indexMap[fromRowType]
					if fromFieldIndex == nil {
						return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - deep struct copy error: source field index does not exists", entity.Name, field.Name)).PrintfError()
					}
					fromField, errInner := fromRow.FieldByIndexErr(fromFieldIndex)
					if errInner != nil {
						return errInner
					}

					if fromField.IsValid() {
						// Если типы можно присваивать
						//if reflect.TypeOf(fromField.Interface()).AssignableTo(reflect.TypeOf(toField.Interface())) {
						//if reflect.TypeOf(toField.Interface()) == reflect.TypeOf(fromField.Interface()) { // Если типы точно совпадают
						toField.Set(fromField)
						//} else {
						//	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - deep struct copy error: Incompatible Struct fields type, fromFieldType='%s', toFieldType='%s'", entity.Name, field.Name, reflect.TypeOf(fromField.Interface()).String(), reflect.TypeOf(toField.Interface()).String())).PrintfError()
						//}
					}
				}
			}
		}

		toStruct.OrgObject = fromStruct // Сохраним ссылку на указатель оригинальной структуры

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromStruct != nil && toStruct != nil {}", []interface{}{fromStruct, toStruct}).PrintfError()
}

// CopyObjectSlice Копирование slice через его указатель по совпадающим полям в структуре
func (entity *Entity) CopyObjectSlice(fromSlice *Object, toSlice *Object, outFields FieldsMap) (err error) {
	if fromSlice != nil && toSlice != nil {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CopySlicePtrByFieldName", entity.Name)
			}
		}()

		if entity.Modify.CopyRestrict {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - forbiden to copy", entity.Name)).PrintfError()
		}

		fromPtrValue := fromSlice.RV
		toPtrValue := toSlice.RV
		//_log.Debug("fromPtrValue.Kind(), toPtrValue.Kind()", fromPtrValue.Kind(), toPtrValue.Kind())

		// На вход получаем только указатели на slice
		if fromPtrValue.Kind() != reflect.Ptr && toPtrValue.Kind() != reflect.Ptr {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromPtrValue.Kind() == reflect.Val && toPtrValue.Kind() == reflect.Val {}", []interface{}{fromPtrValue.Kind().String(), toPtrValue.Kind().String()}).PrintfError()
		}

		fromRows := reflect.Indirect(fromPtrValue) // Собственно slice с данными входными
		toRows := reflect.Indirect(toPtrValue)     // Собственно slice с данными выходными
		//_log.Debug("fromRows.Kind(), toRows.Kind()", fromRows.Kind(), toRows.Kind())

		// На вход получаем только указатели на slice
		if fromRows.Kind() != reflect.Slice || toRows.Kind() != reflect.Slice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromRows.Kind() != reflect.Slice || toRows.Kind() != reflect.Slice {}", []interface{}{fromRows.Kind().String(), toRows.Kind().String()}).PrintfError()
		}

		// копировать в непустой slice нельзя
		if toRows.Len() > 0 {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if toRows.Len() > 0 {}", []interface{}{toRows.Len()}).PrintfError()
		}

		fromRowType := reflect.TypeOf(fromRows.Interface()).Elem() // тип структуры во входном массиве
		toRowType := reflect.TypeOf(toRows.Interface()).Elem()     // тип структуры в выходном массиве
		//_log.Debug("fromRowType.Kind(), toRowType.Kind()", fromRowType.Kind(), toRowType.Kind())

		// копируем только структуры
		if fromRowType.Kind() != reflect.Struct || toRowType.Kind() != reflect.Struct {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromRowType.Kind() != reflect.Struct || toRowType.Kind() != reflect.Struct {}", []interface{}{fromRowType.Kind().String(), toRowType.Kind().String()}).PrintfError()
		}

		//if toRowType == fromRowType {
		if fromRowType.AssignableTo(toRowType) {
			//_log.Debug("Structs has the same type. Do not need to copy")
			toRows.Set(fromRows)
		} else {
			//_log.Debug("Structs was different. Need to copy")
			fromRowsLen := fromRows.Len()
			for i := 0; i < fromRowsLen; i++ {
				//fromRowValuePtr := fromRows.Index(i).Addr() // указатель на текущую структуру
				fromRowValuePtr := fromRows.Index(i)    // указатель на текущую структуру
				toRowValuePtr := reflect.New(toRowType) // новая структура для копирования - указатель

				// Скопируем в новую структуру только те поля, которые нужные в ответе
				if err = entity.CopyStructPtrByFieldName(fromRowValuePtr.Interface(), toRowValuePtr.Interface(), outFields); err != nil {
					return err
				}

				// добавляем в slice структуру из указателя
				toRows.Set(reflect.Append(toRows, reflect.Indirect(toRowValuePtr)))
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromSlice != nil && toSlice != nil {}", []interface{}{fromSlice, toSlice}).PrintfError()
}
