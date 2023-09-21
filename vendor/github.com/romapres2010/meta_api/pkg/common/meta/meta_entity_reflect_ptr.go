package meta

import (
    "fmt"
    "reflect"

    "encoding/xml"

    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// FieldValuePtr - получить в указанной структуре поле по его индексу
func (entity *Entity) FieldValuePtr(field *Field, structPtrValue reflect.Value) (value reflect.Value, err error) {
    if entity != nil && field != nil {

        if !structPtrValue.IsValid() {
            return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        // Функция восстановления после паники в reflect
        defer func() {
            if r := recover(); r != nil {
                err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "FieldValuePtr", entity.Name)
            }
        }()

        structValue := reflect.Indirect(structPtrValue)
        structType := reflect.TypeOf(structValue.Interface())
        //_ = structType

        // На вход получаем только указатель на struct
        if structPtrValue.Kind() != reflect.Ptr || structValue.Kind() != reflect.Struct {
            return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' must be pointer to struct", []interface{}{structValue.Kind()}).PrintfError()
        }

        fieldIndex := field.indexMap[structType] // индекс поля в полученной структуре

        if fieldIndex != nil && len(fieldIndex) != 0 {

            //fieldValue := structValue.FieldByName(field.Name) // значение поля в структуре
            fieldValue, errInner := structValue.FieldByIndexErr(fieldIndex) // значение поля в структуре
            if errInner != nil {
                return reflect.Value{}, errInner
            }

            if fieldValue.IsValid() {
                return fieldValue, nil
            } else {
                //return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - does not exists or not valid index '%v'", entity.Name, field.Name))
                return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - does not exists or not valid index '%v'", entity.Name, field.Name, fieldIndex))
            }
        } else {
            return reflect.Value{}, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - index of field '%s' not found", entity.Name, field.Name))
        }
    }
    return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && field != nil {}", []interface{}{entity, field}).PrintfError()
}

// FieldsValuePtr - получить в указанной структуре поля по списку
func (entity *Entity) FieldsValuePtr(fields Fields, structPtrValue reflect.Value) (values []interface{}, err error) {
    if entity != nil && fields != nil && len(fields) > 0 {

        if !structPtrValue.IsValid() {
            return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        // Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
        for i, field := range fields {
            if field != nil {
                keyArgFieldValue, errInner := entity.FieldValuePtr(field, structPtrValue)
                if errInner != nil {
                    return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, errInner, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by in index", entity.Name, field.Name))
                } else {
                    values = append(values, keyArgFieldValue.Interface())
                }
            } else {
                return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty field pointer '%v'", entity.Name, i)).PrintfError()
            }
        }

        return values, nil
    }
    return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && fields != nil && len(fields) > 0 {}", []interface{}{entity, fields}).PrintfError()
}

// KeyFieldsValuePtr - получить в указанной структуре поля ключа
func (entity *Entity) KeyFieldsValuePtr(key *Key, structPtrValue reflect.Value) (values []interface{}, err error) {
    if entity != nil && key != nil {

        // Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
        values, err = entity.FieldsValuePtr(key.fields, structPtrValue)
        if err != nil {
            return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Key '%s' - error get fields value", entity.Name, key.Name))
        }

        return values, nil
    }
    return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && key != nil {}", []interface{}{entity, key}).PrintfError()
}

// ReferenceFieldsValuePtr - получить в указанной структуре поля reference
func (entity *Entity) ReferenceFieldsValuePtr(reference *Reference, structPtrValue reflect.Value) (values []interface{}, err error) {
    if entity != nil && reference != nil {

        // Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
        values, err = entity.FieldsValuePtr(reference.fields, structPtrValue)
        if err != nil {
            return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Reference '%s' - error get fields value", entity.Name, reference.Name))
        }

        return values, nil
    }
    return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && reference != nil {}", []interface{}{entity, reference}).PrintfError()
}

// SetFieldValuePtr - установить в указанной структуре поле по его индексу
func (entity *Entity) SetFieldValuePtr(field *Field, structPtrValue reflect.Value, val reflect.Value) (err error) {
    if entity != nil && field != nil {
        if !structPtrValue.IsValid() {
            return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        if !val.IsValid() {
            return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - 'val' is not valid", entity.Name, field.Name))
        }

        // Функция восстановления после паники в reflect
        defer func() {
            if r := recover(); r != nil {
                err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetFieldValuePtr", entity.Name)
            }
        }()

        if value, err := entity.FieldValuePtr(field, structPtrValue); err != nil {
            return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by index", entity.Name, field.Name)).PrintfError()
        } else {
            if reflect.TypeOf(val.Interface()).AssignableTo(reflect.TypeOf(value.Interface())) {
                value.Set(val)
            } else {
                return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', field '%s' - Incompatible type, fromFieldType='%s', toFieldType='%s'", entity.Name, field.Name, reflect.TypeOf(val.Interface()).String(), reflect.TypeOf(value.Interface()).String())).PrintfError()
            }
        }

        return nil
    }
    return nil
}

func (entity *Entity) SetXmlNameValuePtr(structPtrValue reflect.Value, xmlName *xml.Name) (err error) {
    if entity != nil {
        if !structPtrValue.IsValid() {
            return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        if entity.xmlNameField != nil {
            if value, err := entity.FieldValuePtr(entity.xmlNameField, structPtrValue); err != nil {
                return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' by index", entity.Name, entity.xmlNameField.Name)).PrintfError()
            } else {
                value.Set(reflect.ValueOf(*xmlName))
            }
        } else {
            return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty xmlNameField", entity.Name)).PrintfError()
        }
        return nil
    }
    return nil
}

func (entity *Entity) SetXmlNameValuePtrFromTag(structPtrValue reflect.Value) (err error) {
    if entity != nil {
        if entity.xmlNameField != nil {
            xmlName := entity.GetXmlNameFromTag(true)
            if err = entity.SetXmlNameValuePtr(structPtrValue, xmlName); err != nil {
                return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR set XMLName", entity.Name)).PrintfError()
            }
        }
    }
    return nil
}

func (entity *Entity) SetXmlNameValuePtrFromTagSlice(slicePtrValue reflect.Value) (err error) {
    if entity != nil {
        if entity.xmlNameField != nil {

            sliceValue := reflect.Indirect(slicePtrValue) // Собственно sliceValue с данными

            // На вход получаем только указатели на slice
            if sliceValue.Kind() != reflect.Slice {
                return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if sliceValue.Kind() != reflect.Slice {}", []interface{}{sliceValue.Kind().String()}).PrintfError()
            }

            xmlName := entity.GetXmlNameFromTag(true)

            sliceValueLen := sliceValue.Len()
            for i := 0; i < sliceValueLen; i++ {
                //structPtrValue := sliceValue.Index(i).Addr()
                structPtrValue := sliceValue.Index(i)
                if err = entity.SetXmlNameValuePtr(structPtrValue, xmlName); err != nil {
                    return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR set XMLName", entity.Name)).PrintfError()
                }
            }
        }
    }
    return nil
}

func (entity *Entity) SetErrorValuePtr(structPtrValue reflect.Value, errors _err.Errors) (err error) {
    if entity != nil && errors != nil && len(errors) > 0 {
        if !structPtrValue.IsValid() {
            return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        if entity.errorsField != nil { // Поле в которое поместить ошибку

            // найдем значение поля, в которое поместить структуру
            if value, err := entity.FieldValuePtr(entity.errorsField, structPtrValue); err != nil {
                return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.errorsField.Name)).PrintfError()
            } else {
                value.Set(reflect.ValueOf(errors))
            }
        } else {
            return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty errorsField", entity.Name)).PrintfError()
        }

        return nil
    }
    return nil
}

func (entity *Entity) CacheInvalidValuePtr(structPtrValue reflect.Value) (invalid bool, err error) {
    if entity != nil {
        if !structPtrValue.IsValid() {
            return true, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        // Функция восстановления после паники в reflect
        defer func() {
            if r := recover(); r != nil {
                err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CacheInvalidValuePtr", entity.Name)
            }
        }()

        if entity.cacheInvalidField != nil {
            // TODO - вынести в общий блок
            if entity.mxField != nil {
                // Определим поле с MX
                if valueMx, err := entity.FieldValuePtr(entity.mxField, structPtrValue); err != nil {
                    return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.mxField.Name)).PrintfError()
                } else {
                    valueMx.Addr().MethodByName("RLock").Call(nil)
                    defer valueMx.Addr().MethodByName("RUnlock").Call(nil)
                }
            } else {
                return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty mxField", entity.Name)).PrintfError()
            }
            // TODO - вынести в общий блок

            if value, err := entity.FieldValuePtr(entity.cacheInvalidField, structPtrValue); err != nil {
                return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.cacheInvalidField.Name)).PrintfError()
            } else {
                return value.Bool(), nil
            }
        } else {
            return true, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty cacheInvalidField", entity.Name)).PrintfError()
        }
    }
    return true, nil
}

func (entity *Entity) SetCacheInvalidValuePtr(structPtrValue reflect.Value, invalid bool) (err error) {
    if entity != nil {
        if !structPtrValue.IsValid() {
            return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "'structPtrValue' is not valid", []interface{}{structPtrValue}).PrintfError()
        }

        // Функция восстановления после паники в reflect
        defer func() {
            if r := recover(); r != nil {
                err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "SetCacheInvalidValuePtr", entity.Name)
            }
        }()

        if entity.cacheInvalidField != nil {
            // TODO - вынести в общий блок
            // Заблокировать значение на время изменения
            if entity.mxField != nil {
                // Определим поле с MX
                if valueMx, err := entity.FieldValuePtr(entity.mxField, structPtrValue); err != nil {
                    return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.mxField.Name)).PrintfError()
                } else {
                    valueMx.Addr().MethodByName("Lock").Call(nil)
                    defer valueMx.Addr().MethodByName("Unlock").Call(nil)
                }
            } else {
                return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty mxField", entity.Name)).PrintfError()
            } // Заблокировать значение на время изменения
            // TODO - вынести в общий блок

            if value, err := entity.FieldValuePtr(entity.cacheInvalidField, structPtrValue); err != nil {
                return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - ERROR get field '%s' value by in index", entity.Name, entity.cacheInvalidField.Name)).PrintfError()
            } else {
                value.SetBool(invalid)
            }

        } else {
            return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' - empty cacheInvalidField", entity.Name)).PrintfError()
        }

        return nil
    }
    return nil
}
