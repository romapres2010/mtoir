package apiservice

import (
	"fmt"
	"reflect"
	"time"

	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"

	"github.com/bytedance/sonic"
	"github.com/tidwall/gjson"
	"github.com/ugorji/go/codec"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
)

// UnmarshalEntity разобрать произвольную структуру из 'json', 'yaml', 'xml'
func (s *Service) UnmarshalEntity(requestID uint64, entity *_meta.Entity, buf []byte, val any, operation, name string, format string, multiRow bool) (err error) {
	//_log.Debug("START: requestID, operation, name", requestID, operation, name)
	tic := time.Now()

	switch format {
	case "json":
		sonicConfig := sonic.Config{
			DisallowUnknownFields: true,
		}.Froze()
		if err = sonicConfig.Unmarshal(buf, val); err != nil {
			// Если JSON валидный, но не того формата -
			if sonic.ConfigDefault.Valid(buf) {
				outMes := ""
				// сформируем пример сообщения для вывода в ошибки - вернем ожидаемый формат структуры
				if entity != nil {
					rowExpected, errInner := s.NewRowAll(requestID, entity, nil)
					if errInner == nil {
						valJson, _ := json.MarshalIndent(rowExpected.Value, "", "    ")
						if multiRow {
							outMes = fmt.Sprintf("{\"expected_json\": \n[\n" + string(valJson) + "\n]\n, \n\"received_json\":" + string(buf) + "}")
						} else {
							outMes = fmt.Sprintf("{\"expected_json\": " + string(valJson) + ", \n\"received_json\":" + string(buf) + "}")
						}
					}
				} else {
					outMes = fmt.Sprintf("{\"received_json\":" + string(buf) + "}")
				}
				errMy := _err.WithCauseTyped(_err.ERR_JSON_UNMARSHAL_ERROR, requestID, err, "Incorrect JSON format")
				errMy.MessageJson = []byte(outMes)
				return errMy
			} else {
				// сформируем пример сообщения для вывода в ошибки - вернем ожидаемый формат структуры
				outMes := ""
				// сформируем пример сообщения для вывода в ошибки - вернем ожидаемый формат структуры
				if entity != nil {
					rowExpected, errInner := s.NewRowAll(requestID, entity, nil)
					if errInner == nil {
						valJson, _ := json.MarshalIndent(rowExpected.Value, "", "    ")
						if multiRow {
							outMes = fmt.Sprintf("{\"expected_json\": \n[\n" + string(valJson) + "\n]\n}")
						} else {
							outMes = fmt.Sprintf("{\"expected_json\": " + string(valJson) + "}")
						}
					}
				} else {
					outMes = "{}"
				}
				errMy := _err.WithCauseTyped(_err.ERR_JSON_UNMARSHAL_ERROR, requestID, err, "Can not parse JSON")
				errMy.MessageJson = []byte(outMes)
				return errMy
			}
		}
	case "xml":
		if err = xml.Unmarshal(buf, val); err != nil {
			return _err.WithCauseTyped(_err.ERR_XML_UNMARSHAL_ERROR, requestID, err).PrintfError()
		}
	case "yaml":
		if err = yaml.Unmarshal(buf, val); err != nil {
			return _err.WithCauseTyped(_err.ERR_YAML_UNMARSHAL_ERROR, requestID, err).PrintfError()
		}
	default:
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Allowed only 'format'='json', 'yaml', 'xml'", format).PrintfError()
	}
	_metrics.IncUnMarshalingDurationVec(format, operation, name, time.Now().Sub(tic))

	//_log.Debug("SUCCESS: requestID, operation, name, duration", requestID, operation, name, time.Now().Sub(tic))
	return nil
}

// UnmarshalSingleObject - распарсить одну строку в структуру
func (s *Service) UnmarshalSingleObject(requestID uint64, entity *_meta.Entity, options *_meta.Options, inBuf []byte, inFormat string, ignoreExtra bool) (row *_meta.Object, err error) {
	if s != nil && entity != nil && inBuf != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entity.Name", requestID, entity.Name)

		if inFormat == "json1" { // Временно отключен

			var fieldsVal = make(map[string]any)

			// Первым проходом проверяем валидный ли JSON
			// Вторым проходом сканируем JSON, чтобы понять состав полей - допустимая ситуация, когда передается не все поля
			// Для переданных полей формируем структуру и уже в нее третьим проходом парсим JSON

			// Если не валидный JSON, то распарсим его, чтобы подсветить ошибку
			if !gjson.ValidBytes(inBuf) {
				if err = codec.NewDecoderBytes(inBuf, &codec.JsonHandle{}).Decode(&fieldsVal); err != nil {
					// Сформировать структуру как образец JSON
					myErr := _err.WithCauseTyped(_err.ERR_JSON_UNMARSHAL_ERROR, requestID, err)
					if rowExpected, err := s.NewRowAll(requestID, entity, options); err == nil {
						//myErr.Extra = rowExpected.PtrValue
						myErr.Extra = rowExpected.Value
					}
					return nil, myErr.PrintfError()
				}
			}

			{ // Формирование ограниченного набора полей по тегам, которые есть в запросе
				var inFieldsName []string

				{ // считаем данные из JSON для анализа тегов
					gjson.ParseBytes(inBuf).ForEach(func(key, value gjson.Result) bool {
						if key.Str != "" {
							inFieldsName = append(inFieldsName, key.Str)
						}
						return true // keep iterating
					})
				} // считаем данные из JSON для анализа тегов

				//// Подготовим map с полями, которые нужны, лишние поля игнорируем
				//inFields, err := s.constructFieldsMap(requestID, entity, inFieldsName, "json", ignoreExtra, true)
				//if err != nil {
				//	return nil, err
				//}
				//
				//// Сформировать входную структуру
				//if rowInPtr, rowInType, err = s.newRowPtrRestrict(requestID, entity, inFields); err != nil {
				//	return nil, err
				//}

			} // Формирование ограниченного набора полей по тегам, которые есть в запросе

			// Сформировать входную структуру
			if row, err = s.NewRowAll(requestID, entity, options); err != nil {
				return nil, err
			}

			rowInValue := reflect.Indirect(row.RV) //  собственно структура

			{ // парсинг JSON в структуру
				// https://www.perimeterx.com/tech-blog/2022/boosting-up-json-performance-of-unstructured-structs-in-go/
				// https://stackoverflow.com/questions/24348184/get-pointer-to-value-using-reflection

				// сформируем маппинг для итогового разбора
				for i := 0; i < rowInValue.NumField(); i++ {
					fieldIn := rowInValue.Field(i)
					jsonTag := row.StructType.Field(i).Tag.Get("json")
					if !fieldIn.CanAddr() {
						return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("if !valueField.CanAddr() {}")).PrintfError()
					}
					fieldsVal[jsonTag] = fieldIn.Addr().Interface() // адрес поля структуры в которую записать данные
				}

				// парсинг JSON в структуру правильно заполняются поля NULL, time.Time
				if err = codec.NewDecoderBytes(inBuf, &codec.JsonHandle{}).Decode(&fieldsVal); err != nil {
					return nil, _err.WithCauseTyped(_err.ERR_JSON_UNMARSHAL_ERROR, requestID, err).PrintfError()
				}
			} // парсинг JSON в структуру
		} else {
			// Еси формат не JSON, то разбираем стандартным UnmarshalEntity

			//Сформировать входную структуру
			if row, err = s.NewRowAll(requestID, entity, options); err != nil {
				return nil, err
			}

			err = s.UnmarshalEntity(requestID, entity, inBuf, row.Value, "UnmarshalSingleObject", entity.Name, inFormat, false)
			if err != nil {
				return nil, err
			}
		}

		_log.Debug("SUCCESS: requestID, entity.Name, duration", requestID, entity.Name, time.Now().Sub(tic))
		return row, nil
	}

	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && inBuf != nil {}", []interface{}{s, entity.Name}).PrintfError()
}

// UnmarshalMultiObject - распарсить много строк в структуру
func (s *Service) UnmarshalMultiObject(requestID uint64, entity *_meta.Entity, options *_meta.Options, inBuf []byte, inFormat string, ignoreExtra bool) (rows *_meta.Object, err error) {
	if s != nil && entity != nil && inBuf != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entity.Name", requestID, entity.Name)

		//Сформировать входную структуру
		if rows, err = s.NewSliceAll(requestID, entity, options, 0, 64); err != nil {
			return nil, err
		}

		err = s.UnmarshalEntity(requestID, entity, inBuf, rows.Value, "UnmarshalMultiObject", entity.Name, inFormat, true)
		if err != nil {
			return nil, err
		}

		_log.Debug("SUCCESS: requestID, entity.Name, duration", requestID, entity.Name, time.Now().Sub(tic))
		return rows, nil
	}

	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && inBuf != nil {}", []interface{}{s, entity.Name}).PrintfError()
}

// UnmarshalMultiInterface - распарсить много строк в []interface{}
func (s *Service) UnmarshalMultiInterface(requestID uint64, entity *_meta.Entity, inBuf []byte, inFormat string) (rowsI []interface{}, err error) {
	if s != nil && entity != nil && inBuf != nil {

		//tic := time.Now()
		//_log.Debug("START: requestID, entity.Name", requestID, entity.Name)

		err = s.UnmarshalEntity(requestID, entity, inBuf, &rowsI, "UnmarshalMultiInterface", entity.Name, inFormat, true)
		if err != nil {
			return nil, err
		}

		//_log.Debug("SUCCESS: requestID, entity.Name, duration", requestID, entity.Name, time.Now().Sub(tic))
		return rowsI, nil
	}

	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && inBuf != nil {}", []interface{}{s, entity.Name}).PrintfError()
}

// UnmarshalSingleInterface - распарсить строку в interface{}
func (s *Service) UnmarshalSingleInterface(requestID uint64, entity *_meta.Entity, inBuf []byte, inFormat string) (rowI interface{}, err error) {
	if s != nil && entity != nil && inBuf != nil {

		//tic := time.Now()
		//_log.Debug("START: requestID, entity.Name", requestID, entity.Name)

		err = s.UnmarshalEntity(requestID, entity, inBuf, &rowI, "UnmarshalSingleInterface", entity.Name, inFormat, false)
		if err != nil {
			return nil, err
		}

		//_log.Debug("SUCCESS: requestID, entity.Name, duration", requestID, entity.Name, time.Now().Sub(tic))
		return rowI, nil
	}

	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && entity != nil && inBuf != nil {}", []interface{}{s, entity.Name}).PrintfError()
}
