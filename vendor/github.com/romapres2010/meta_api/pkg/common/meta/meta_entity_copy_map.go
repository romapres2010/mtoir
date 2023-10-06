package meta

import (
	"encoding"
	"fmt"
	"math"
	"reflect"

	"gopkg.in/guregu/null.v4"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// CopyMapToStruct - Копирование map в struct
func (entity *Entity) CopyMapToStruct(fromMap map[string]interface{}, toRow *Object, toFields FieldsMap) (empty bool, err error) {

	if fromMap != nil && toRow != nil {
		//https://www.golinuxcloud.com/go-map-to-struct/

		var ok bool
		var restrict = toFields != nil && len(toFields) > 0
		var errors _err.Errors
		var inFormat = toRow.Options.Global.InFormat

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "CopyMapToStruct", entity.Name)
			}
		}()

		// Нет полей для копирования
		if len(fromMap) == 0 {
			return true, nil
		}

		// Обрабатываем только структуры
		if toRow.IsSlice {
			return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "toRow must be pointer to struct", []interface{}{toRow.RV.Kind().String(), reflect.Indirect(toRow.RV).Kind()}).PrintfError()
		}

		// Собственно структура
		toRowRV := reflect.Indirect(toRow.RV)

		// Копируем по конкретным именам полей и их индексам
		for _, field := range toRow.Entity.StructFields() {

			// Ограничить поля определенным списком
			if restrict {
				if _, ok = toFields[field.Name]; !ok {
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

			//if fromMapValue, ok := fromMap[field.GetTagName("json", false)]; ok {
			if fromMapValue, ok := fromMap[field.GetTagName(inFormat, false)]; ok {

				fromMapRV := reflect.ValueOf(fromMapValue)

				if fromMapRV.IsValid() && !fromMapRV.IsZero() {

					// Reference оставляем без изменений - их обработает рекурсия
					if field.reference != nil {

						if field.InternalType == FIELD_TYPE_ASSOCIATION {
							// Обработать Association

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
								if err = floatToRV(entity.Name, field.Name, fromValueFloat, toFieldRV); err != nil {
									errors.Append(_err.ERR_UNDEFINED_ID, err)
									continue // собираем все ошибки, что сможем конвертируем
								}
							} else {
								return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromValueFloat, ok := fromMapValue.(float64); ok {}").PrintfError()
							}

						} else if fromMapRV.Kind() == reflect.String {

							// Если на входе строка, которую можно распарсить
							if fromValueString, ok := fromMapValue.(string); ok {
								if err = stringToRV(entity.Name, field.Name, fromValueString, toFieldRV); err != nil {
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
					//_log.Debug("Entity field empty value in map: entityName, fieldName", entity.Name, field.Name)
				}
			} else {
				//_log.Debug("Entity field have not value in map: entityName, fieldName", entity.Name, field.Name)
			}
		}

		// Накопили ошибки
		if errors.HasError() {
			return false, errors.Error(_err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - error fill from map[string]interface{}", entity.Name))
		}
		return false, nil
	}
	return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if fromMap != nil && toRow != nil {}", []interface{}{fromMap, toRow}).PrintfError()
}

func floatToRV(entityName string, fieldName string, floatValue float64, rv reflect.Value) error {
	if rv.IsValid() {

		// Типы, к которым можно приводить
		switch rv.Kind() {

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fromValueFloatRound := math.Round(floatValue)

			// Если число целое, то можно преобразовать
			if floatValue == fromValueFloatRound {
				switch rv.Kind() {
				case reflect.Int:
					rv.Set(reflect.ValueOf(int(fromValueFloatRound)))
				case reflect.Int8:
					rv.Set(reflect.ValueOf(int8(fromValueFloatRound)))
				case reflect.Int16:
					rv.Set(reflect.ValueOf(int16(fromValueFloatRound)))
				case reflect.Int32:
					rv.Set(reflect.ValueOf(int32(fromValueFloatRound)))
				case reflect.Int64:
					rv.Set(reflect.ValueOf(int64(fromValueFloatRound)))
				}
			} else {
				// Дробное число
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(rv.Interface()).String(), floatValue)).PrintfError()
			}
			break

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fromValueFloatRound := math.Round(floatValue)

			// Если число целое, то можно преобразовать
			if floatValue > 0 && floatValue == fromValueFloatRound {
				switch rv.Kind() {
				case reflect.Uint:
					rv.Set(reflect.ValueOf(uint(fromValueFloatRound)))
				case reflect.Uint8:
					rv.Set(reflect.ValueOf(uint8(fromValueFloatRound)))
				case reflect.Uint16:
					rv.Set(reflect.ValueOf(uint16(fromValueFloatRound)))
				case reflect.Uint32:
					rv.Set(reflect.ValueOf(uint32(fromValueFloatRound)))
				case reflect.Uint64:
					rv.Set(reflect.ValueOf(uint64(fromValueFloatRound)))
				}
			} else {
				// Дробное число
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(rv.Interface()).String(), floatValue)).PrintfError()
			}
			break

		case reflect.Float32:
			rv.Set(reflect.ValueOf(float32(floatValue)))
			break

		case reflect.Struct:

			switch (rv.Interface()).(type) {

			case null.Int:
				fromValueFloatRound := math.Round(floatValue)
				// Если число целое, то можно преобразовать
				if floatValue > 0 && floatValue == fromValueFloatRound {
					rv.Set(reflect.ValueOf(null.IntFrom(int64(fromValueFloatRound))))
				} else {
					// Дробное число
					return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(rv.Interface()).String(), floatValue)).PrintfError()
				}

			case null.Float:
				rv.Set(reflect.ValueOf(null.FloatFrom(floatValue)))

			default:
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(rv.Interface()).String(), floatValue)).PrintfError()
			}

		default:
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(floatValue).String(), reflect.TypeOf(rv.Interface()).String(), floatValue)).PrintfError()
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if rv.IsValid() {}", []interface{}{rv}).PrintfError()
}

func stringToRV(entityName string, fieldName string, stringValue string, rv reflect.Value) error {
	if rv.IsValid() {

		var rvPtrInterface interface{}

		if rv.Kind() == reflect.Ptr {

			// Если указатель, то создать объект типа, в который будем парсить
			rvPtr := reflect.New(rv.Type().Elem())
			rv.Set(rvPtr) // TODO - перенести присвоение после успешного парсинга
			rvPtrInterface = rvPtr.Interface()

		} else if rv.Kind() == reflect.Struct {

			// Если тип структура, то получим указатель для парсинга
			rvPtr := rv.Addr()
			rvPtrInterface = rvPtr.Interface()

		} else {

			// Не поддерживаемый тип
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - unsupported type for deep struct copy fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(stringValue).String(), reflect.TypeOf(rv.Interface()).String(), stringValue)).PrintfError()
		}

		if rvPtrInterfaceTM, ok := rvPtrInterface.(encoding.TextUnmarshaler); ok {

			if rvPtrInterfaceTM != nil {

				if reflect.ValueOf(rvPtrInterfaceTM).Kind() == reflect.Ptr && reflect.ValueOf(rvPtrInterfaceTM).IsNil() {
					// пустой указатель в интерфейсе
					return _err.NewTypedTraceEmpty(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if reflect.ValueOf(rvPtrInterfaceTM).Kind() == reflect.Ptr && reflect.ValueOf(rvPtrInterfaceTM).IsNil() {}").PrintfError()
				} else {
					// Парсим данные
					err := rvPtrInterfaceTM.UnmarshalText([]byte(stringValue))
					if err != nil {
						return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Field '%s' - deep struct copy - error UnmarshalText, fromFieldType='%s', toFieldType='%s', fromFieldValue='%v'", entityName, fieldName, reflect.TypeOf(stringValue).String(), reflect.TypeOf(rv.Interface()).String(), stringValue)).PrintfError()
					}
				}
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if rv.IsValid() {}", []interface{}{rv}).PrintfError()
}
