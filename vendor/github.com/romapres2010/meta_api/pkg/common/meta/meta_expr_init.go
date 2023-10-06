package meta

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

func nullString() null.String {
	var val null.String
	return val
}

func nullFloat() null.Float {
	var val null.Float
	return val
}

func nullInt() null.Int {
	var val null.Int
	return val
}

func nullBool() null.Bool {
	var val null.Bool
	return val
}

func nullTime() null.Time {
	var val null.Time
	return val
}

func nullTimeNow() null.Time {
	return null.TimeFrom(time.Now())
}

func timeNow() time.Time {
	return time.Now()
}

func newUUID() uuid.UUID {
	return uuid.New()
}

func ptrEmptyUUID() *uuid.UUID {
	//uuidVal := uuid.UUID{}
	return nil
}

func emptyUUID() uuid.UUID {
	uuidVal := uuid.UUID{}
	return uuidVal
}

func ptrNewUUID() *uuid.UUID {
	uuidVal := uuid.New()
	return &uuidVal
}

func ptrUUID(uuidValFrom *uuid.UUID) *uuid.UUID {
	if uuidValFrom != nil {
		return uuidValFrom
	} else {
		uuidVal := uuid.New()
		return &uuidVal
	}
}

func NewExpr(entity *Entity, field *Field, status string, name string, type_ ExprType, action ExprAction, code string, fieldsArsName []string, fieldsDestName []string, doInit bool) (ex *Expr, err error) {
	if entity != nil {

		// Создадим и инициируем новое Expression
		ex = &Expr{
			Status:         status,
			Name:           name,
			Type:           type_,
			Action:         action,
			Code:           code,
			FieldsArgsName: fieldsArsName,
			FieldsDestName: fieldsDestName,
		}

		if doInit {
			if err = ex.init(entity, field); err != nil {
				return nil, err
			}
			return ex, nil
		} else {
			return ex, nil
		}

	}
	return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Expression '%s' - empty Entity or field", name))
}

func (ex *Expr) init(entity *Entity, field *Field) (err error) {
	if entity != nil {

		ex.entity = entity // TODO - определить место присвоения

		var baseFieldName string

		if ex.Name == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' - empty 'name'", ex.entity.Name, ex.Name))
		}

		if ex.Status == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' - empty 'status'", ex.entity.Name, ex.Name))
		}

		if field != nil {
			ex.field = field
			baseFieldName = field.Name // Выражение может быть привязано к сущности в целом
		}

		if ex.Type != EXPR_CALCULATE &&
			ex.Type != EXPR_DB_CALCULATE &&
			ex.Type != EXPR_VALIDATE &&
			ex.Type != EXPR_FILTER &&
			ex.Type != EXPR_COPY &&
			ex.Type != EXPR_CONVERT {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - incorrect 'type' '%s'", ex.entity.Name, baseFieldName, ex.Name, ex.Type))
		}

		if ex.Action != EXPR_ACTION_ALL &&
			ex.Action != EXPR_ACTION_GET &&
			ex.Action != EXPR_ACTION_POST_GET &&
			ex.Action != EXPR_ACTION_INSIDE_GET &&
			ex.Action != EXPR_ACTION_PUT &&
			ex.Action != EXPR_ACTION_PRE_CREATE &&
			ex.Action != EXPR_ACTION_PRE_UPDATE &&
			ex.Action != EXPR_ACTION_PRE_PUT &&
			ex.Action != EXPR_ACTION_POST_PUT &&
			ex.Action != EXPR_ACTION_MARSHAL &&
			ex.Action != EXPR_ACTION_PRE_MARSHAL &&
			ex.Action != EXPR_ACTION_INSIDE_MARSHAL &&
			ex.Action != EXPR_ACTION_POST_MARSHAL &&
			ex.Action != EXPR_ACTION_POST_FETCH {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - incorrect 'action' '%s'", ex.entity.Name, baseFieldName, ex.Name, ex.Action))
		}

		if ex.Code == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("field '%s' field '%s' Expression ['%s'] - code", ex.entity.Name, baseFieldName, ex.Name))
		}

		if ex.Name == "" {
			ex.Name = ex.Code // В качестве имени используем собственно выражение
		}

		if ex.Code != "-" && ex.Type != EXPR_DB_CALCULATE { // Не требуется настраивать компилятор

			// TODO - вынести все глобальные опции отдельно
			TimeNow := expr.Function(
				"TimeNow",
				func(params ...any) (any, error) {
					return timeNow(), nil
				},
				timeNow,
			)

			NullTimeNow := expr.Function(
				"NullTimeNow",
				func(params ...any) (any, error) {
					return nullTimeNow(), nil
				},
				nullTimeNow,
			)

			PtrNewUUID := expr.Function(
				"PtrNewUUID",
				func(params ...any) (any, error) {
					return ptrNewUUID(), nil
				},
				ptrNewUUID,
			)

			PtrUUID := expr.Function(
				"PtrUUID",
				func(params ...any) (any, error) {
					if val, ok := params[0].(*uuid.UUID); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*uuid.UUID)(nil)).String())).PrintfError()
					} else {
						return ptrUUID(val), nil
					}
				},
				ptrUUID,
			)

			PtrEmptyUUID := expr.Function(
				"PtrEmptyUUID",
				func(params ...any) (any, error) {
					return ptrEmptyUUID(), nil
				},
				ptrEmptyUUID,
			)

			EmptyUUID := expr.Function(
				"EmptyUUID",
				func(params ...any) (any, error) {
					return emptyUUID(), nil
				},
				emptyUUID,
			)

			NewUUID := expr.Function(
				"NewUUID",
				func(params ...any) (any, error) {
					return newUUID(), nil
				},
				newUUID,
			)

			NullString := expr.Function(
				"NullString",
				func(params ...any) (any, error) {
					return nullString(), nil
				},
				nullString,
			)

			StringFrom := expr.Function(
				"StringFrom",
				func(params ...any) (any, error) {
					if val, ok := params[0].(string); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*string)(nil)).Elem().String())).PrintfError()
					} else {
						return null.StringFrom(val), nil
					}
				},
				null.StringFrom,
			)

			NullFloat := expr.Function(
				"NullFloat",
				func(params ...any) (any, error) {
					return nullFloat(), nil
				},
				nullFloat,
			)

			FloatFrom := expr.Function(
				"FloatFrom",
				func(params ...any) (any, error) {
					if val, ok := params[0].(float64); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*float64)(nil)).Elem().String())).PrintfError()
					} else {
						return null.FloatFrom(val), nil
					}
				},
				null.FloatFrom,
			)

			NullBool := expr.Function(
				"NullBool",
				func(params ...any) (any, error) {
					return nullBool(), nil
				},
				nullBool,
			)

			BoolFrom := expr.Function(
				"BoolFrom",
				func(params ...any) (any, error) {
					if val, ok := params[0].(bool); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*bool)(nil)).Elem().String())).PrintfError()
					} else {
						return null.BoolFrom(val), nil
					}
				},
				null.BoolFrom,
			)

			NullInt := expr.Function(
				"NullInt",
				func(params ...any) (any, error) {
					return nullInt(), nil
				},
				nullInt,
			)

			IntFrom := expr.Function(
				"IntFrom",
				func(params ...any) (any, error) {
					if val, ok := params[0].(int64); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*int64)(nil)).Elem().String())).PrintfError()
					} else {
						return null.IntFrom(val), nil
					}
				},
				null.IntFrom,
			)

			NullTime := expr.Function(
				"NullTime",
				func(params ...any) (any, error) {
					return nullTime(), nil
				},
				nullTime,
			)

			TimeFrom := expr.Function(
				"TimeFrom",
				func(params ...any) (any, error) {
					if val, ok := params[0].(time.Time); !ok {
						return nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - computation error - incorrect argument type, recieved '%s', expected '%s'", ex.entity.Name, baseFieldName, ex.Name, reflect.ValueOf(params[0]).Type().String(), reflect.TypeOf((*time.Time)(nil)).Elem().String())).PrintfError()
					} else {
						return null.TimeFrom(val), nil
					}
				},
				null.TimeFrom,
			)

			options := []expr.Option{
				expr.AllowUndefinedVariables(), // Allow to use undefined variables https://pkg.go.dev/github.com/antonmedv/expr#AllowUndefinedVariables
				TimeNow,
				NullTimeNow,
				NewUUID,
				PtrEmptyUUID,
				EmptyUUID,
				PtrNewUUID,
				PtrUUID,
				NullString,
				StringFrom,
				NullFloat,
				FloatFrom,
				NullBool,
				BoolFrom,
				NullInt,
				IntFrom,
				NullTime,
				TimeFrom,
			}
			// TODO - вынести все глобальные опции отдельно

			if ex.program, err = expr.Compile(ex.Code, options...); err != nil {
				return _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - compilation error", ex.entity.Name, baseFieldName, ex.Name)).PrintfError()
			}
		}

		if len(ex.FieldsArgsName) > 0 {
			ex.argsFieldsMap = make(FieldsMap, len(ex.FieldsArgsName))
			ex.argsFields = make(Fields, 0, len(ex.FieldsArgsName))

			for _, fieldName := range ex.FieldsArgsName {
				if exField := ex.entity.fieldByNameUnsafe(fieldName); exField != nil {

					//if exField.Status != STATUS_ENABLED {
					//	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - field '%s' must be enabled", ex.entity.Name, baseFieldName, ex.Name, fieldName))
					//}

					if _, ok := ex.argsFieldsMap[exField.Name]; ok {
						return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s']  - duplicate field '%s'", ex.entity.Name, baseFieldName, ex.Name, fieldName))
					} else {
						ex.argsFieldsMap[exField.Name] = exField
					}

					ex.argsFields = append(ex.argsFields, exField)
					ex.argsFieldsString = ex.argsFieldsString + ", " + fieldName + "(" + exField.InternalType + ")"

				} else {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - field '%s' not found", ex.entity.Name, baseFieldName, ex.Name, fieldName))
				}
				ex.argsFieldsString = strings.TrimLeft(ex.argsFieldsString, ", ")
			}
		}

		if len(ex.FieldsDestName) > 0 {
			ex.destFieldsMap = make(FieldsMap, len(ex.FieldsDestName))
			ex.destFields = make(Fields, 0, len(ex.FieldsDestName))

			for _, fieldName := range ex.FieldsDestName {
				if exField := ex.entity.fieldByNameUnsafe(fieldName); exField != nil {

					//if exField.Status != STATUS_ENABLED {
					//	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - field '%s' must be enabled", ex.entity.Name, baseFieldName, ex.Name, fieldName))
					//}

					if _, ok := ex.destFieldsMap[exField.Name]; ok {
						return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s']  - duplicate destination field '%s'", ex.entity.Name, baseFieldName, ex.Name, fieldName))
					} else {
						ex.destFieldsMap[exField.Name] = exField
					}

					ex.destFields = append(ex.destFields, exField)
					ex.destFieldsString = ex.destFieldsString + ", " + fieldName + "(" + exField.InternalType + ")"

				} else {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression ['%s'] - destination field '%s' not found", ex.entity.Name, baseFieldName, ex.Name, fieldName))
				}
				ex.destFieldsString = strings.TrimLeft(ex.destFieldsString, ", ")
			}
		}

		ex.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Expression '%s' - empty Entity or field", ex.Name))
}
