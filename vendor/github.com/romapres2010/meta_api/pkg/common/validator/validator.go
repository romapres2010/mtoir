package validator

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"gopkg.in/guregu/null.v4"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

type Validator struct {
	Validator  *validator.Validate // валидатор для проверки структуры
	Translator ut.Translator       // переводчик
	Meta       *_meta.Meta         // метаданные на которых работает валидатор

	mx sync.RWMutex
}

// NewValidator - создать новый валидатор и переводчик
func NewValidator(meta *_meta.Meta, applyEntityRule bool) (*Validator, error) {
	var valid = Validator{}
	var err error

	// создаем локаль для перевода ошибок
	eng := en.New()

	valid.Meta = meta

	// создаем новый транслятор
	uni := ut.New(eng, eng)
	valid.Translator, _ = uni.GetTranslator("en")

	// Создаем валидатор пересоздавать его каждый раз не эффективно
	valid.Validator = validator.New()

	// Регистрируем кастомную функцию на конкретный тег
	//err = valid.Validator.RegisterValidation("my_validation", ValidateMyVal)
	//if err != nil {
	//    return nil, err
	//}

	// register all sql.Null* types to use the ValidateValuer CustomTypeFunc
	valid.Validator.RegisterCustomTypeFunc(ValidateValuer, null.String{}, null.Float{}, null.Bool{}, null.Int{})

	// регистрируем переводчик
	err = en_translations.RegisterDefaultTranslations(valid.Validator, valid.Translator)
	if err != nil {
		return nil, err
	}

	// Настроим правила валидации для каждой сущности
	if meta != nil && applyEntityRule {
		for _, entity := range meta.Entities {
			rules, entType := entity.ValidationRulesType()
			_log.Debug("RegisterValidationRules: externalId, entity.Name, entity.Type, rules", _err.ERR_UNDEFINED_ID, entity.Name, reflect.TypeOf(entType), rules)
			valid.RegisterValidationRules(rules, entType)
		}
	}

	return &valid, nil
}

func (v *Validator) GetFieldDisplayName(entityName string, fieldName string) string {
	if v != nil && v.Meta != nil {
		return v.Meta.GetFieldDisplayName(entityName, fieldName)
	} else {
		return _meta.FIELD_NOT_FOUND
	}
}

func (v *Validator) GetFieldFullName(entityName string, fieldName string) string {
	if v != nil && v.Meta != nil {
		return v.Meta.GetFieldFullName(entityName, fieldName)
	} else {
		return _meta.FIELD_NOT_FOUND
	}
}

func (v *Validator) RegisterValidationRules(rules map[string]string, types interface{}) {
	if v != nil && rules != nil && len(rules) > 0 && types != nil {
		v.mx.Lock()
		defer v.mx.Unlock()

		v.Validator.RegisterStructValidationMapRules(rules, types)
	}
}

func (v *Validator) ValidateStruct(externalId uint64, val any) (err error) {
	// Функция восстановления после паники
	defer func() {
		r := recover()
		if r != nil {
			err = _recover.GetRecoverError(r, externalId, "validateStruct '")
		}
	}()

	if val != nil && v != nil && v.Validator != nil {

		v.mx.RLock()
		defer v.mx.RUnlock()

		// returns nil or ValidationErrors ( []FieldError )
		if err := v.Validator.Struct(val); err != nil {

			// this check is only needed when your code could produce
			// an invalid value for validation such as interface with nil
			// value most including myself do not usually have code like this.
			if _, ok := err.(*validator.InvalidValidationError); ok {
				return _err.WithCauseTyped(_err.ERR_VALIDATOR_COMMON_ERROR, externalId, err, "validator.InvalidValidationError", fmt.Sprintf("%+v", err)).PrintfError()
			}

			// Основное сообщения об ошибке
			causeMes := fmt.Sprintf("%+v", err)

			// Дополнительные сообщения об ошибке с переводом
			for _, validationError := range err.(validator.ValidationErrors) {
				causeMes += "\n" + validationError.Translate(v.Translator)
				//_log.PrintfDebugMsg("=================================================================")
				//_log.PrintfDebugMsg(fmt.Sprintf("Namespace %s", validationError.Namespace()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("field %s", validationError.field()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("StructNamespace %s", validationError.StructNamespace()), nil) // can differ when a custom TagNameFunc is registered or
				//_log.PrintfDebugMsg(fmt.Sprintf("StructType %s", validationError.StructType()), nil)         // by passing alt name to ReportError like below
				//_log.PrintfDebugMsg(fmt.Sprintf("Tag %s", validationError.Tag()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("ActualTag %s", validationError.ActualTag()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("Kind %s", validationError.Kind()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("Type %s", validationError.Type()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("Value %s", validationError.Value()), nil)
				//_log.PrintfDebugMsg(fmt.Sprintf("Param %s", validationError.Param()), nil)
			}
			valJson, _ := json.Marshal(val)

			return _err.NewTyped(_err.ERR_VALIDATOR_STRUCT_ERROR, externalId, causeMes, valJson)
		} else {
			return nil
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if val != nil && v != nil && v.Validator != nil {}", []interface{}{val, v}).PrintfError()
}

func (v *Validator) ValidateObject(externalId uint64, row *_meta.Object) (err error) {

	tic := time.Now()

	if v != nil && v.Validator != nil && row != nil && row.Entity != nil {

		// Обрабатываем только структуры
		if row.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if row.IsSlice {}", []interface{}{v, row}).PrintfError()
		}

		// Функция восстановления после паники
		defer func() {
			r := recover()
			if r != nil {
				errRecover := _recover.GetRecoverError(r, externalId, "validateStruct '"+row.Entity.Alias.FullName+"'")
				err = _err.WithCauseTyped(_err.ERR_VALIDATOR_COMMON_ERROR, externalId, errRecover, fmt.Sprintf("%+v", errRecover)).PrintfError()
			}
		}()

		v.mx.RLock()
		defer v.mx.RUnlock()

		// returns nil or ValidationErrors ( []FieldError )
		if err = v.Validator.Struct(row.Value); err != nil {

			entityFullName := row.Entity.Name

			if _, ok := err.(*validator.InvalidValidationError); ok {
				return _err.WithCauseTyped(_err.ERR_VALIDATOR_COMMON_ERROR, externalId, err, fmt.Sprintf("%+v", err)).PrintfError()
			}

			errors := _err.Errors{}

			// Дополнительные сообщения об ошибке с переводом
			for _, validationError := range err.(validator.ValidationErrors) {
				errorFieldName := validationError.StructField()
				field := row.Entity.FieldByName(errorFieldName)

				if field != nil && field.Status == _meta.STATUS_ENABLED {
					//entityFieldName := field.Name
					//errors.Append(externalId, _err.NewTypedTraceEmpty(_err.ERR_VALIDATOR_FIELD_COMMON_ERROR, externalId, entityFullName, entityFieldName, validationError.Namespace(), validationError.Tag(), validationError.Type(), validationError.Value(), validationError.Param(), validationError.Translate(v.Translator)))
					errors.Append(externalId, _err.NewTypedTraceEmpty(_err.ERR_VALIDATOR_FIELD_COMMON_ERROR, externalId, entityFullName, validationError.Namespace(), validationError.Namespace(), validationError.Tag(), validationError.Type(), validationError.Value(), validationError.Param(), validationError.Translate(v.Translator)))
				}
			}

			if errors.HasError() {
				return errors.Error(externalId, fmt.Sprintf("Entity '%s' - validation error", row.Entity.Name))
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", externalId, row.Entity.Name, time.Now().Sub(tic))
			}
			return nil
		} else {
			return nil
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if v != nil && v.Validator != nil && row != nil && row.Entity != nil {}", []interface{}{v, row}).PrintfError()
}

//// ValidateMyVal implements validator.Func
//func ValidateMyVal(fl validator.FieldLevel) bool {
//    return fl.field().String() == "awesome"
//}

// ValidateValuer implements validator.CustomTypeFunc
func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
		// handle the error how you want
	}

	return nil
}
