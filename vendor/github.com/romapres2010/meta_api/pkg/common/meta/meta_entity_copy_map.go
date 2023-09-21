package meta

import (
	"encoding"
	"fmt"
	"math"
	"reflect"

	"gopkg.in/guregu/null.v4"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// CopyMapToStruct - Копирование map в struct
func (entity *Entity) CopyMapToStruct(fromMapInterface interface{}, toRow *Object, outFields FieldsMap) (empty bool, err error) {

	if fromMapInterface != nil && toRow != nil {
		//https://www.golinuxcloud.com/go-map-to-struct/

		var fromMap map[string]interface{}
		var ok bool
		var restrict = outFields != nil && len(outFields) > 0
		var errors _err.Errors

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CopyMapToStruct", entity.Name)
			}
		}()

		// На вход получаем map[string]interface{} или *map[string]interface{}
		if reflect.TypeOf(fromMapInterface).Kind() != reflect.Ptr {
			//fromMapInterface = reflect.Indirect(reflect.ValueOf(fromMapInterface)).Interface()
			fromMap, ok = fromMapInterface.(map[string]interface{})
			if !ok {
				return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "fromMapInterface must be map[string]interface{}", []interface{}{reflect.TypeOf(fromMapInterface).Kind().String()}).PrintfError()
			}

		} else {
			fromMapPtr, ok := fromMapInterface.(*map[string]interface{})
			if !ok {
				return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "fromMapInterface must be map[string]interface{}", []interface{}{reflect.TypeOf(fromMapInterface).Kind().String()}).PrintfError()
			}
			fromMap = *fromMapPtr
		}

		// Нет полей для копирования
		if len(fromMap) == 0 {
			return true, nil
		}

		// На вход получаем только указатели на struct
		if toRow.RV.Kind() != reflect.Ptr || reflect.Indirect(toRow.RV).Kind() != reflect.Struct {
			return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "toRow.Value must be pointer to struct", []interface{}{toRow.RV.Kind().String(), reflect.Indirect(toRow.RV).Kind()}).PrintfError()
		}

		// Собственно структура
		toRowRV := reflect.Indirect(toRow.RV)

		// Копируем по конкретным именам полей и их индексам
		// TODO - добавить проверку по лишним полям в fromMap
		for _, field := range toRow.Entity.StructFields() {

			// Ограничить поля определенным списком
			if restrict {
				if _, ok = outFields[field.Name]; !ok {
					continue
				}
			}

			toFieldIndex := field.indexMap[toRow.StructType]
			if toFieldIndex == nil {
				err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - deep struct copy error: destination field index does not exists", entity.Name, field.Name)).PrintfError()
				errors.Append(_err.ERR_UNDEFINED_ID, err)
				continue // собираем все ошибки, что сможем конвертируем
			}

			toFieldRV, err := toRowRV.FieldByIndexErr(toFieldIndex)
			if err != nil {
				errors.Append(_err.ERR_UNDEFINED_ID, err)
				continue // собираем все ошибки, что сможем конвертируем
			}

			if fromMapValue, ok := fromMap[field.GetTagName("json", false)]; ok {

				fromMapRV := reflect.ValueOf(fromMapValue)

				if fromMapRV.IsValid() && !fromMapRV.IsZero() {

					// Reference оставляем без изменений - их обработает рекурсия
					if field.reference != nil {

						if field.InternalType == FIELD_TYPE_ASSOCIATION {
							// Обработать Association

							// TODO - если целевое поле *struct{}, что делать с map[string]interface{}
							// Association оставляем без изменений - их обработает рекурсия
							if _, ok = fromMapValue.(map[string]interface{}); ok {
								toFieldRV.Set(reflect.ValueOf(fromMapValue)) // присвоить map[string]interface{} к interface{}
							} else {
								err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - Association '%s' must be 'map[string]interface{}' but type fromFieldType='%s'", entity.Name, field.Name, field.reference.Name, reflect.TypeOf(fromMapValue).String()))
								errors.Append(_err.ERR_UNDEFINED_ID, err)
								continue // собираем все ошибки, что сможем конвертируем
							}

						} else if field.InternalType == FIELD_TYPE_COMPOSITION {
							// Обработать Composition

							if !field.reference.Embed {
								// Composition НЕ встраиваемые - slice оставляем без изменений - их обработает рекурсия
								if _, ok = fromMapValue.([]interface{}); ok {
									toFieldRV.Set(reflect.ValueOf(fromMapValue)) // присвоить []interface{} к interface{}
								} else {
									err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - Composition '%s' must be '[]interface{}' but type fromFieldType='%s'", entity.Name, field.Name, field.reference.Name, reflect.TypeOf(fromMapValue).String()))
									errors.Append(_err.ERR_UNDEFINED_ID, err)
									continue // собираем все ошибки, что сможем конвертируем
								}
							} else {
								// Composition встраиваемые - []map[string]interface{} оставляем без изменений - их обработает рекурсия
								if _, ok = fromMapValue.(map[string]interface{}); ok {
									toFieldRV.Set(reflect.ValueOf(fromMapValue)) // присвоить []map[string]interface{} к interface{}
								} else {
									err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - Composition '%s' embeded must be 'map[string]interface{}' but type fromFieldType='%s'", entity.Name, field.Name, field.reference.Name, reflect.TypeOf(fromMapValue).String()))
									errors.Append(_err.ERR_UNDEFINED_ID, err)
									continue // собираем все ошибки, что сможем конвертируем
								}
							}
						} else {
							err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported Reference '%s' InternalType='%s'", entity.Name, field.Name, field.reference.Name, field.InternalType))
							errors.Append(_err.ERR_UNDEFINED_ID, err)
							continue // собираем все ошибки, что сможем конвертируем
						}
					} else {
						// Остальные поля пробуем копировать, кроме системных
						if field.System {
							continue
						}

						if fromMapRV.Type().AssignableTo(toFieldRV.Type()) {

							// Если типы можно напрямую присваивать
							toFieldRV.Set(fromMapRV)

						} else if fromMapRV.Kind() == reflect.Float64 {

							// Из JSON все числа парсятся в float64, попробуем привести к типу
							if fromValueFloat, ok := fromMapValue.(float64); ok {
								if err = floatToReflectValue(entity.Name, field.Name, fromValueFloat, toFieldRV); err != nil {
									errors.Append(_err.ERR_UNDEFINED_ID, err)
									continue // собираем все ошибки, что сможем конвертируем
								}
							} else {
								return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromValueFloat, ok := fromMapValue.(float64); ok {}").PrintfError()
							}

						} else if fromMapRV.Kind() == reflect.String {

							// Если на входе строка, которую можно распарсить
							if fromValueString, ok := fromMapValue.(string); ok {
								if err = stringToReflectValue(entity.Name, field.Name, fromValueString, toFieldRV); err != nil {
									errors.Append(_err.ERR_UNDEFINED_ID, err)
									continue // собираем все ошибки, что сможем конвертируем
								}
							} else {
								return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromValueFloat, ok := fromMapValue.(string); ok {}").PrintfError()
							}

						} else {

							// Не поддерживаемый тип
							err = _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - deep struct copy - incompatible Struct fields type, fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entity.Name, field.Name, reflect.TypeOf(fromMapRV.Interface()).String(), reflect.TypeOf(toFieldRV.Interface()).String(), fromMapValue)).PrintfError()
							errors.Append(_err.ERR_UNDEFINED_ID, err)
							continue // собираем все ошибки, что сможем конвертируем
						}
					}
				} else {
					// TODO - Если поле обязательно - давать ошибку?
					_log.Debug("Entity field empty value in map: entityName, fieldName", entity.Name, field.Name)
				}
			} else {
				// TODO - Если поле обязательно - давать ошибку?
				_log.Debug("Entity field have not value in map: entityName, fieldName", entity.Name, field.Name)
			}
		}

		// Накопили ошибки
		if errors.HasError() {
			return false, errors.Error(_err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - error fill from map[string]interface{}", entity.Name))
		}
		return false, nil
	}
	return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromMapInterface != nil && toRow != nil {}", []interface{}{fromMapInterface, toRow}).PrintfError()
}

func floatToReflectValue(entityName string, fieldName string, floatValue float64, value reflect.Value) error {
	if value.IsValid() {

		// Типы, к которым можно приводить
		switch value.Kind() {

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fromValueFloatround := math.Round(floatValue)

			// Если число целое, то можно преобразовать
			if floatValue == fromValueFloatround {
				switch value.Kind() {
				case reflect.Int:
					value.Set(reflect.ValueOf(int(fromValueFloatround)))
				case reflect.Int8:
					value.Set(reflect.ValueOf(int8(fromValueFloatround)))
				case reflect.Int16:
					value.Set(reflect.ValueOf(int16(fromValueFloatround)))
				case reflect.Int32:
					value.Set(reflect.ValueOf(int32(fromValueFloatround)))
				case reflect.Int64:
					value.Set(reflect.ValueOf(int64(fromValueFloatround)))
				}
			} else {
				// Дробное число
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(value.Interface()).String(), floatValue)).PrintfError()
			}
			break

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fromValueFloatround := math.Round(floatValue)

			// Если число целое, то можно преобразовать
			if floatValue > 0 && floatValue == fromValueFloatround {
				switch value.Kind() {
				case reflect.Uint:
					value.Set(reflect.ValueOf(uint(fromValueFloatround)))
				case reflect.Uint8:
					value.Set(reflect.ValueOf(uint8(fromValueFloatround)))
				case reflect.Uint16:
					value.Set(reflect.ValueOf(uint16(fromValueFloatround)))
				case reflect.Uint32:
					value.Set(reflect.ValueOf(uint32(fromValueFloatround)))
				case reflect.Uint64:
					value.Set(reflect.ValueOf(uint64(fromValueFloatround)))
				}
			} else {
				// Дробное число
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(value.Interface()).String(), floatValue)).PrintfError()
			}
			break

		case reflect.Float32:
			value.Set(reflect.ValueOf(float32(floatValue)))
			break

		case reflect.Struct:

			switch (value.Interface()).(type) {

			case null.Int:
				fromValueFloatround := math.Round(floatValue)
				// Если число целое, то можно преобразовать
				if floatValue > 0 && floatValue == fromValueFloatround {
					value.Set(reflect.ValueOf(null.IntFrom(int64(fromValueFloatround))))
				} else {
					// Дробное число
					return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(value.Interface()).String(), floatValue)).PrintfError()
				}

			case null.Float:
				value.Set(reflect.ValueOf(null.FloatFrom(floatValue)))

			default:
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(value.Interface()).String(), floatValue)).PrintfError()
			}

		default:
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(value.Interface()).String(), floatValue)).PrintfError()
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if value.IsValid() {}", []interface{}{value}).PrintfError()
}

func stringToReflectValue(entityName string, fieldName string, stringValue string, value reflect.Value) error {
	if value.IsValid() {

		var valuePtr reflect.Value
		var valuePtrInterface interface{}

		if value.Kind() == reflect.Ptr {

			// Если указатель, то создать объект типа, в который будем парсить
			valuePtr = reflect.New(value.Type().Elem())
			value.Set(valuePtr) // TODO - перенести присвоение после успешного парсинга
			valuePtrInterface = valuePtr.Interface()

		} else if value.Kind() == reflect.Struct {

			// Если тип структура, то получим указатель для парсинга
			valuePtr = value.Addr()
			valuePtrInterface = valuePtr.Interface()

		} else {

			// Не поддерживаемый тип
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(stringValue).String(), reflect.TypeOf(value.Interface()).String(), stringValue)).PrintfError()
		}

		if valuePtrInterfaceTM, ok := valuePtrInterface.(encoding.TextUnmarshaler); ok {

			if valuePtrInterfaceTM != nil {

				if reflect.ValueOf(valuePtrInterfaceTM).Kind() == reflect.Ptr && reflect.ValueOf(valuePtrInterfaceTM).IsNil() {
					// пустой указатель в интерфейсе
					return _err.NewTypedTraceEmpty(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if reflect.ValueOf(valuePtrInterfaceTM).Kind() == reflect.Ptr && reflect.ValueOf(valuePtrInterfaceTM).IsNil() {}").PrintfError()
				} else {
					// Парсим данные
					err := valuePtrInterfaceTM.UnmarshalText([]byte(stringValue))
					if err != nil {
						return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Field '%s' - deep struct copy - error UnmarshalText, fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(stringValue).String(), reflect.TypeOf(value.Interface()).String(), stringValue)).PrintfError()
					}
				}
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if value.IsValid() {}", []interface{}{value}).PrintfError()
}
