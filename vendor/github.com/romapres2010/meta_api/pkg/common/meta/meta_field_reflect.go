package meta

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"encoding/xml"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

const (
	FIELD_TYPE_COMPOSITION      = "COMPOSITION"
	FIELD_TYPE_ASSOCIATION      = "ASSOCIATION"
	FIELD_TYPE_XML_NAME         = "XMLName"
	FIELD_TYPE_INTERNAL_ERROR   = "*_err.Errors"
	FIELD_TYPE_CACHE_INVALID    = "bool"
	FIELD_TYPE_VALIDATION_VALID = "bool"
	FIELD_TYPE_RWMUTEX          = "sync.RWMutex"
)

var (
	FIELD_TYPE_ASSOCIATION_RT = reflect.TypeOf((*interface{})(nil)).Elem()
	FIELD_TYPE_COMPOSITION_RT = reflect.TypeOf((*interface{})(nil)).Elem()
	FIELD_TYPE_SLICE_ANY_RT   = reflect.SliceOf(reflect.TypeOf((*interface{})(nil)).Elem())
	FIELD_TYPE_XML_NAME_RT    = reflect.TypeOf(xml.Name{})
	//FIELD_TYPE_INTERNAL_ERROR_RT = reflect.TypeOf((*_err.Errors)(nil))
	FIELD_TYPE_INTERNAL_ERROR_RT = reflect.TypeOf((*_err.Errors)(nil)).Elem()
	FIELD_TYPE_RWMUTEX_RT        = reflect.TypeOf(sync.RWMutex{})
)

var reflectTypeMap = map[string]reflect.Type{
	"bool":                    reflect.TypeOf((*bool)(nil)).Elem(),
	"*bool":                   reflect.TypeOf((*bool)(nil)),
	"int":                     reflect.TypeOf((*int)(nil)).Elem(),
	"*int":                    reflect.TypeOf((*int)(nil)),
	"int8":                    reflect.TypeOf((*int8)(nil)).Elem(),
	"*int8":                   reflect.TypeOf((*int8)(nil)),
	"int16":                   reflect.TypeOf((*int16)(nil)).Elem(),
	"*int16":                  reflect.TypeOf((*int16)(nil)),
	"int32":                   reflect.TypeOf((*int32)(nil)).Elem(),
	"*int32":                  reflect.TypeOf((*int32)(nil)),
	"int64":                   reflect.TypeOf((*int64)(nil)).Elem(),
	"*int64":                  reflect.TypeOf((*int64)(nil)),
	"uint":                    reflect.TypeOf((*uint)(nil)).Elem(),
	"*uint":                   reflect.TypeOf((*uint)(nil)),
	"uint8":                   reflect.TypeOf((*uint8)(nil)).Elem(),
	"*uint8":                  reflect.TypeOf((*uint8)(nil)),
	"uint16":                  reflect.TypeOf((*uint16)(nil)).Elem(),
	"*uint16":                 reflect.TypeOf((*uint16)(nil)),
	"uint32":                  reflect.TypeOf((*uint32)(nil)).Elem(),
	"*uint32":                 reflect.TypeOf((*uint32)(nil)),
	"uint64":                  reflect.TypeOf((*uint64)(nil)).Elem(),
	"*uint64":                 reflect.TypeOf((*uint64)(nil)),
	"uintptr":                 reflect.TypeOf((*uintptr)(nil)).Elem(),
	"float32":                 reflect.TypeOf((*float32)(nil)).Elem(),
	"*float32":                reflect.TypeOf((*float32)(nil)),
	"float64":                 reflect.TypeOf((*float64)(nil)).Elem(),
	"*float64":                reflect.TypeOf((*float64)(nil)),
	"complex64":               reflect.TypeOf((*complex64)(nil)).Elem(),
	"*complex64":              reflect.TypeOf((*complex64)(nil)),
	"complex128":              reflect.TypeOf((*complex128)(nil)).Elem(),
	"*complex128":             reflect.TypeOf((*complex128)(nil)),
	"string":                  reflect.TypeOf((*string)(nil)).Elem(),
	"*string":                 reflect.TypeOf((*string)(nil)),
	"null.String":             reflect.TypeOf((*null.String)(nil)).Elem(),
	"*null.String":            reflect.TypeOf((*null.String)(nil)),
	"null.Int":                reflect.TypeOf((*null.Int)(nil)).Elem(),
	"*null.Int":               reflect.TypeOf((*null.Int)(nil)),
	"null.Float":              reflect.TypeOf((*null.Float)(nil)).Elem(),
	"*null.Float":             reflect.TypeOf((*null.Float)(nil)),
	"null.Time":               reflect.TypeOf((*null.Time)(nil)).Elem(),
	"*null.Time":              reflect.TypeOf((*null.Time)(nil)),
	"null.Bool":               reflect.TypeOf((*null.Bool)(nil)).Elem(),
	"*null.Bool":              reflect.TypeOf((*null.Bool)(nil)),
	"time.Time":               reflect.TypeOf((*time.Time)(nil)).Elem(),
	"*time.Time":              reflect.TypeOf((*time.Time)(nil)),
	"[]interface{}":           reflect.TypeOf([]interface{}{}).Elem(),
	"interface{}":             reflect.TypeOf((*interface{})(nil)).Elem(),
	"*interface{}":            reflect.TypeOf((*interface{})(nil)),
	FIELD_TYPE_COMPOSITION:    FIELD_TYPE_COMPOSITION_RT,
	FIELD_TYPE_ASSOCIATION:    FIELD_TYPE_ASSOCIATION_RT,
	FIELD_TYPE_XML_NAME:       FIELD_TYPE_XML_NAME_RT,
	FIELD_TYPE_INTERNAL_ERROR: FIELD_TYPE_INTERNAL_ERROR_RT,
	FIELD_TYPE_RWMUTEX:        FIELD_TYPE_RWMUTEX_RT,
	"UIID":                    reflect.TypeOf((*uuid.UUID)(nil)).Elem(),
	"*UIID":                   reflect.TypeOf((*uuid.UUID)(nil)),
}

func (field *Field) getReflectType() (reflect.Type, error) {
	if field != nil && field.entity != nil {
		if field.Status != STATUS_ENABLED {
			return nil, nil
		}

		if field.InternalType == "" {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - empty 'type'", field.entity.Name, field.Name))
		}

		reflectType, ok := reflectTypeMap[field.InternalType]

		if !ok {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - has usupperted 'internal_type'='%s'", field.entity.Name, field.Name, field.InternalType))
		}
		return reflectType, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if field != nil && field.Entity != nil {}", []interface{}{field}).PrintfError()
}
