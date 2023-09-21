package meta

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
)

func ArgToString(arg interface{}) string {
	if arg == nil {
		return ""
	} else {

		// Интерфейс может не содержать данные, если это указатель
		//val := reflect.Indirect(reflect.ValueOf(arg))
		val := reflect.ValueOf(arg)
		if !(val.IsValid() && !val.IsZero()) {
			return ""
		}

		switch f := arg.(type) {
		case string:
			return f
		case *string:
			return *f
		default:
			if argTM, ok := f.(encoding.TextMarshaler); ok {
				if argTM != nil {
					text, err := argTM.MarshalText()
					if err != nil {
						return fmt.Sprintf("%v", arg) // ключевое поле может быть любого типа - приводим к строке
					} else {
						return string(text)
					}
				} else {
					return ""
				}
			} else {
				return fmt.Sprintf("%v", arg) // ключевое поле может быть любого типа - приводим к строке
			}
		}
	}
}

// ArgNotEmpty - аргумент пуст
func ArgNotEmpty(arg interface{}) bool {
	val := reflect.ValueOf(arg)
	if val.IsValid() && !val.IsZero() {
		return true
	}
	return false
}

// ArgsAllEmpty - все аргументы пустые
func ArgsAllEmpty(args []interface{}) bool {
	if len(args) == 0 {
		return true
	}

	anyNotEmpty := false
	for _, arg := range args {
		val := reflect.ValueOf(arg)
		if val.IsValid() && !val.IsZero() {
			anyNotEmpty = true
			break
		}
	}
	return anyNotEmpty == false
}

func ArgsToStrings(args ...interface{}) []string {

	argsStr := make([]string, 0, len(args))

	// Преобразовать к строке
	for _, arg := range args {
		argsStr = append(argsStr, ArgToString(arg))
	}
	return argsStr
}

func ArgsToString(sep string, args ...interface{}) string {
	if len(args) == 0 {
		return ""
	} else {
		return strings.Join(ArgsToStrings(args...), sep)
	}
}

func ArgsSliceToStrings(args []interface{}) []string {

	argsStr := make([]string, 0, len(args))

	// Преобразовать к строке
	for _, arg := range args {
		argsStr = append(argsStr, ArgToString(arg))
	}
	return argsStr
}

func ArgsSliceToString(sep string, args []interface{}) string {
	if len(args) == 0 {
		return ""
	} else {
		return strings.Join(ArgsToStrings(args), sep)
	}
}

func ArgsMapToStrings(sep string, args map[string]interface{}) string {

	if len(args) == 0 {
		return ""
	}

	argsStr := make([]string, 0, len(args))

	for _, arg := range args {
		argsStr = append(argsStr, ArgToString(arg))
	}
	return strings.Join(argsStr, sep)
}

// unmarshal разобрать произвольную структуру из 'json', 'yaml', 'xml'
func unmarshal(buf []byte, val any, operation, name string, format string) (err error) {
	_log.Debug("START: name", name)

	tic := time.Now()

	switch format {
	case "json":
		if err = json.Unmarshal(buf, val); err != nil {
			return _err.WithCauseTyped(_err.ERR_JSON_UNMARSHAL_ERROR, _err.ERR_UNDEFINED_ID, err).PrintfError()
		}
	case "xml":
		if err = xml.Unmarshal(buf, val); err != nil {
			return _err.WithCauseTyped(_err.ERR_XML_UNMARSHAL_ERROR, _err.ERR_UNDEFINED_ID, err).PrintfError()
		}
	case "yaml":
		if err = yaml.Unmarshal(buf, val); err != nil {
			return _err.WithCauseTyped(_err.ERR_YAML_UNMARSHAL_ERROR, _err.ERR_UNDEFINED_ID, err).PrintfError()
		}
	default:
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "Allowed only 'format'='json', 'yaml', 'xml'", format).PrintfError()
	}
	_metrics.IncUnMarshalingDurationVec(format, operation, name, time.Now().Sub(tic))

	_log.Debug("SUCCESS: name, duration", name, time.Now().Sub(tic))
	return nil
}
