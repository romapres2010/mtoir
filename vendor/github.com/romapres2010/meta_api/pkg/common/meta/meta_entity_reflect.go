package meta

import (
	"fmt"
	"reflect"

	"encoding/xml"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// FieldRV - получить в указанной структуре поле по его индексу
func (entity *Entity) FieldRV(field *Field, row *Object) (value reflect.Value, err error) {
	if entity != nil && field != nil && row != nil && entity == row.Entity {

		rowValue := row.RV
		structValue := reflect.Indirect(rowValue)

		if !rowValue.IsValid() {
			return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "FieldRV", entity.Name)
			}
		}()

		// На вход получаем только указатель на struct
		if rowValue.Kind() != reflect.Ptr || structValue.Kind() != reflect.Struct {
			return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'row.Value' must be pointer to struct", []interface{}{structValue.Kind()}).PrintfError()
		}

		fieldIndex := field.indexMap[row.StructType] // индекс поля в полученной структуре

		if fieldIndex != nil && len(fieldIndex) != 0 {

			fieldValue, errInner := structValue.FieldByIndexErr(fieldIndex) // значение поля в структуре
			if errInner != nil {
				return reflect.Value{}, errInner
			}

			if fieldValue.IsValid() {
				return fieldValue, nil
			} else {
				return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - does not exists or not valid, index '%v'", entity.Name, field.Name, fieldIndex)).PrintfError()
			}
		} else {
			return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - index of field '%s' not found", entity.Name, field.Name)).PrintfError()
		}
	}
	return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && field != nil && row != nil && entity == row.Entity {}", []interface{}{entity, field, row}).PrintfError()
}

// FieldValue - получить в указанной структуре значение поля
func (entity *Entity) FieldValue(field *Field, row *Object) (value interface{}, err error) {
	if entity != nil && field != nil && row != nil && entity == row.Entity {

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		keyArgFieldRV, errInner := entity.FieldRV(field, row)
		if errInner != nil {
			return nil, err
		} else {
			switch keyArgFieldRV.Kind() {
			case reflect.Float32, reflect.Float64:
				value = keyArgFieldRV.Float()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value = keyArgFieldRV.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value = keyArgFieldRV.Uint()
			case reflect.Bool:
				value = keyArgFieldRV.Bool()
			case reflect.String:
				value = keyArgFieldRV.String()
			default:
				value = keyArgFieldRV.Interface() // Дорогая операция
			}
		}

		return value, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && field != nil && row != nil && entity == row.Entity {}", []interface{}{entity, field, row}).PrintfError()
}

// FieldsValue - получить в указанной структуре поля по списку
func (entity *Entity) FieldsValue(fields Fields, row *Object) (values []interface{}, err error) {
	if entity != nil && fields != nil && len(fields) > 0 && row != nil && entity == row.Entity {

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		values = make([]interface{}, 0, len(fields))

		for i, field := range fields {
			if field != nil {
				fieldValue, errInner := entity.FieldValue(field, row)
				if errInner != nil {
					return nil, err
				} else {
					values = append(values, fieldValue)
				}
			} else {
				return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty field pointer '%v'", entity.Name, i)).PrintfError()
			}
		}

		return values, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && fields != nil && row != nil && len(fields) > 0 && entity == row.Entity {}", []interface{}{entity, fields, row}).PrintfError()
}

// FieldsRV - получить в указанной структуре поля по списку
func (entity *Entity) FieldsRV(fields Fields, row *Object) (values []reflect.Value, err error) {
	if entity != nil && fields != nil && len(fields) > 0 && row != nil && entity == row.Entity {

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		values = make([]reflect.Value, 0, len(fields))

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		for i, field := range fields {
			if field != nil {
				fieldRV, errInner := entity.FieldRV(field, row)
				if errInner != nil {
					return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, errInner, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by in index", entity.Name, field.Name)).PrintfError()
				} else {
					values = append(values, fieldRV) // Дорогая операция
				}
			} else {
				return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty field pointer '%v'", entity.Name, i)).PrintfError()
			}
		}

		return values, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && fields != nil && row != nil && len(fields) > 0 && entity == row.Entity {}", []interface{}{entity, fields, row}).PrintfError()
}

// KeyFieldsValue - получить в указанной структуре поля ключа
func (entity *Entity) KeyFieldsValue(key *Key, row *Object) (values []interface{}, err error) {
	if entity != nil && row != nil && entity == row.Entity {

		if key == nil {
			return nil, err
		}

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		values, err = entity.FieldsValue(key.fields, row)
		if err != nil {
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s'['%s'] - error get fields value", entity.Name, key.Name, key.FieldsString()))
		}

		return values, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

// KeyFieldsRV - получить в указанной структуре поля ключа
func (entity *Entity) KeyFieldsRV(key *Key, row *Object) (values []reflect.Value, err error) {
	if entity != nil && row != nil && entity == row.Entity {

		if key == nil {
			return nil, err
		}

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		values, err = entity.FieldsRV(key.fields, row)
		if err != nil {
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s'['%s'] - error get fields value", entity.Name, key.Name, key.FieldsString()))
		}

		return values, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

// ReferenceFieldsValue - получить в указанной структуре поля reference
func (entity *Entity) ReferenceFieldsValue(reference *Reference, row *Object) (values []interface{}, err error) {
	if entity != nil && reference != nil && row != nil && entity == row.Entity {

		if !row.RV.IsValid() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		values, err = entity.FieldsValue(reference.fields, row)
		if err != nil {
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Reference '%s' - error get fields value", entity.Name, reference.Name)).PrintfError()
		}

		return values, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && reference != nil && row != nil && entity == row.Entity {}", []interface{}{entity, reference, row}).PrintfError()
}

// SetFieldRV - установить в указанной структуре поле по его индексу
func (entity *Entity) SetFieldRV(field *Field, row *Object, inRV reflect.Value) (err error) {
	if entity != nil && field != nil && row != nil && entity == row.Entity {

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetFieldRV", entity.Name)
			}
		}()

		if !row.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - 'row' is not valid", entity.Name, field.Name)).PrintfError()
		}

		if !inRV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - 'inRV' is not valid", entity.Name, field.Name)).PrintfError()
		}

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetFieldRV", entity.Name)
			}
		}()

		if fieldRV, err := entity.FieldRV(field, row); err != nil {
			return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by index", entity.Name, field.Name)).PrintfError()
		} else {
			if inRV.Type().AssignableTo(fieldRV.Type()) {
				fieldRV.Set(inRV)
			} else {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - Incompatible type, fromFieldType='%s', toFieldType='%s'", entity.Name, field.Name, reflect.TypeOf(inRV.Interface()).String(), reflect.TypeOf(fieldRV.Interface()).String())).PrintfError()
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && field != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

// ZeroFieldRV - обнулить в указанной структуре поле по его индексу
func (entity *Entity) ZeroFieldRV(field *Field, row *Object) (err error) {
	if entity != nil && field != nil && row != nil && entity == row.Entity {

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetFieldRV", entity.Name)
			}
		}()

		if !row.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - 'row' is not valid", entity.Name, field.Name)).PrintfError()
		}

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetFieldRV", entity.Name)
			}
		}()

		if fieldRV, err := entity.FieldRV(field, row); err != nil {
			return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by index", entity.Name, field.Name)).PrintfError()
		} else {
			fieldRV.Set(reflect.Zero(fieldRV.Type())) // Очищаем поле
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && field != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

func (entity *Entity) SetXmlNameValue(row *Object, xmlName *xml.Name) (err error) {
	if entity != nil && row != nil && entity == row.Entity {

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetXmlNameValue", entity.Name)
			}
		}()

		if !row.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		if entity.xmlNameField != nil {
			if value, err := entity.FieldRV(entity.xmlNameField, row); err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by index", entity.Name, entity.xmlNameField.Name)).PrintfError()
			} else {
				value.Set(reflect.ValueOf(xmlName).Elem())
			}
		} else {
			return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty xmlNameField", entity.Name)).PrintfError()
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

func (entity *Entity) SetXmlNameValueFromTag(row *Object) (err error) {
	if entity != nil && row != nil && entity == row.Entity {

		if !row.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		if entity.xmlNameField != nil {
			xmlName := entity.GetXmlNameFromTag(true)
			if err = entity.SetXmlNameValue(row, xmlName); err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR set XMLName", entity.Name)).PrintfError()
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil && entity == row.Entity {}", []interface{}{entity, row}).PrintfError()
}

func (entity *Entity) SetXmlNameValueSlice(slice *Object, xmlName *xml.Name) (err error) {
	if entity != nil && slice != nil && entity == slice.Entity {

		if !slice.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'slice' is not valid", entity.Name)).PrintfError()
		}

		if entity.xmlNameField != nil {

			sliceValue := reflect.Indirect(slice.RV) // Собственно sliceValue с данными

			// На вход получаем только указатели на slice
			if sliceValue.Kind() != reflect.Slice {
				return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if sliceValue.Kind() != reflect.Slice {}", []interface{}{sliceValue.Kind().String()}).PrintfError()
			}

			for _, row := range slice.Objects {
				if err = entity.SetXmlNameValue(row, xmlName); err != nil {
					return err
				}
			}

			//for i := 0; i < sliceValue.Len(); i++ {
			//
			//	rowPtrValue := sliceValue.Index(i).Addr() // указатель на текущую структуру
			//	row := slice.NewFromRV(rowPtrValue)
			//
			//	if err = entity.SetXmlNameValue(row, xmlName); err != nil {
			//		return err
			//	}
			//}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && slice != nil && entity == slice.Entity {}", []interface{}{entity, slice}).PrintfError()
}

func (entity *Entity) SetXmlNameValueFromTagSlice(slice *Object) (err error) {
	if entity != nil && slice != nil && entity == slice.Entity {

		if !slice.RV.IsValid() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
		}

		if entity.xmlNameField != nil {

			sliceValue := reflect.Indirect(slice.RV) // Собственно sliceValue с данными

			// На вход получаем только указатели на slice
			if sliceValue.Kind() != reflect.Slice {
				return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if sliceValue.Kind() != reflect.Slice {}", []interface{}{sliceValue.Kind().String()}).PrintfError()
			}

			xmlName := entity.GetXmlNameFromTag(true)
			if err = entity.SetXmlNameValueSlice(slice, xmlName); err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR set XMLName", entity.Name)).PrintfError()
			}

			//for i := 0; i < sliceValue.Len(); i++ {
			//	row := &Object{
			//		StructType: slice.StructType,
			//		fields:     slice.fields,
			//		//Val:        sliceValue.Index(i).Addr(),
			//		Value: sliceValue.Index(i).Addr(),
			//	}
			//	if err = entity.SetXmlNameValue(row, xmlName); err != nil {
			//		return err
			//	}
			//}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && slice != nil && entity == slice.Entity {}", []interface{}{entity, slice}).PrintfError()
}

func (entity *Entity) SetErrorValue(row *Object, errors _err.Errors) (err error) {
	if entity != nil && row != nil && entity == row.Entity {

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetErrorValue", entity.Name)
			}
		}()

		if errors != nil && len(errors) > 0 {

			if !row.RV.IsValid() {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
			}

			if entity.errorsField != nil { // Поле в которое поместить ошибку

				// найдем значение поля, в которое поместить структуру
				if value, err := entity.FieldRV(entity.errorsField, row); err != nil {
					return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.errorsField.Name)).PrintfError()
				} else {
					if value.IsValid() {
						if value.IsZero() {
							// Предыдущих ошибок нет
							value.Set(reflect.ValueOf(errors))
						} else {

							valueIndirect := reflect.Indirect(value) // Собственно slice

							// Дописываем ошибки
							for _, errValue := range errors {
								// добавляем в slice структуру из указателя
								valueIndirect.Set(reflect.Append(valueIndirect, reflect.ValueOf(errValue)))
							}
						}
					}
				}
			} else {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty errorsField", entity.Name)).PrintfError()
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil && entity == row.Entity  {}", []interface{}{entity, row}).PrintfError()
}

//
//func (entity *Entity) CacheInvalidValue(row *Object) (invalid bool, err error) {
//	if entity != nil && row != nil {
//
//		if !row.Value.IsValid() {
//			return true, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
//		}
//
//		// Функция восстановления после паники в reflect
//		defer func() {
//			if r := recover(); r != nil {
//				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CacheInvalidValue", entity.Name)
//			}
//		}()
//
//		if entity.cacheInvalidField != nil {
//			// TODO - вынести в общий блок
//			if entity.mxField != nil {
//				// Определим поле с MX
//				if valueMx, err := entity.FieldRV(entity.mxField, row); err != nil {
//					return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.mxField.Name)).PrintfError()
//				} else {
//					valueMx.Addr().MethodByName("RLock").Call(nil)
//					defer valueMx.Addr().MethodByName("RUnlock").Call(nil)
//				}
//			} else {
//				return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty mxField", entity.Name)).PrintfError()
//			}
//			// TODO - вынести в общий блок
//
//			if value, err := entity.FieldRV(entity.cacheInvalidField, row); err != nil {
//				return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.cacheInvalidField.Name)).PrintfError()
//			} else {
//				return value.Bool(), nil
//			}
//		} else {
//			return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty cacheInvalidField", entity.Name)).PrintfError()
//		}
//	}
//	return true, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil {}", []interface{}{entity, row}).PrintfError()
//}
//
//func (entity *Entity) SetCacheInvalidValue(row *Object, invalid bool) (err error) {
//	if entity != nil && row != nil {
//
//		if !row.Value.IsValid() {
//			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'row' is not valid", entity.Name)).PrintfError()
//		}
//
//		// Функция восстановления после паники в reflect
//		defer func() {
//			if r := recover(); r != nil {
//				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetCacheInvalidValue", entity.Name)
//			}
//		}()
//
//		if entity.cacheInvalidField != nil {
//			// TODO - вынести в общий блок
//			// Заблокировать значение на время изменения
//			if entity.mxField != nil {
//				// Определим поле с MX
//				if valueMx, err := entity.FieldRV(entity.mxField, row); err != nil {
//					return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.mxField.Name)).PrintfError()
//				} else {
//					valueMx.Addr().MethodByName("Lock").Call(nil)
//					defer valueMx.Addr().MethodByName("Unlock").Call(nil)
//				}
//			} else {
//				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty mxField", entity.Name)).PrintfError()
//			} // Заблокировать значение на время изменения
//			// TODO - вынести в общий блок
//
//			if value, err := entity.FieldRV(entity.cacheInvalidField, row); err != nil {
//				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.cacheInvalidField.Name)).PrintfError()
//			} else {
//				value.SetBool(invalid)
//			}
//
//		} else {
//			return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty cacheInvalidField", entity.Name)).PrintfError()
//		}
//
//		return nil
//	}
//	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && row != nil {}", []interface{}{entity, row}).PrintfError()
//}
