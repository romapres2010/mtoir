package meta

import (
	"reflect"
	"strings"
	"sync"

	"encoding/xml"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

type ObjectReferenceMap map[*Reference]*Object

type Objects []*Object

type Object struct {
	Entity        *Entity       // ссылка на сущность
	Reference     *Reference    // Заполняется, если объект является ссылочным полем
	Fields        FieldsMap     // реальный список полей в объекте - может отличаться в меньшую сторону
	Value         interface{}   // указатель на struct или slice с данными
	RV            reflect.Value // reflect.ValueOf(Value)
	StructType    reflect.Type  // тип struct{} или []*struct{} с данными
	StructPtrType reflect.Type  // тип *struct{} или []*struct{} с данными

	Options *Options // Опции, на которых был построен объект TODO при сохранении в кэш опции нужно очищать

	IsSlice bool    // признак, что объект является slice
	Objects Objects // для slice содержит массив объектов

	AssociationMap ObjectReferenceMap // ассоциации, повешенные на поля
	CompositionMap ObjectReferenceMap // композиции, повешенные на поля
	OrgObject      *Object            // указатель на struct или slice с данными, на основании которого построен данный объект

	cacheInvalid bool         // один объект может кешироваться с разными ключами - признак используется для инвалидации значения в кэше
	mx           sync.RWMutex // блокировка объекта - используется для целостности по чтению из кэша и для инвалидации данных
}

func newObject(entity *Entity, fields FieldsMap, value interface{}, reflectValue reflect.Value, structType reflect.Type, structPtrType reflect.Type, isSlice bool) *Object {
	o := &Object{
		Entity:        entity,
		StructType:    structType,
		StructPtrType: structPtrType,
		Fields:        fields,
		Value:         value,
		RV:            reflectValue,
		cacheInvalid:  false,
		IsSlice:       isSlice,
	}

	o.OrgObject = o

	return o
}

func (o *Object) NewFromRV(rv reflect.Value, isSlice bool) *Object {
	if o != nil && o.Entity != nil {
		val := rv.Interface()
		out := &Object{
			Entity:        o.Entity,
			StructType:    o.StructType,
			StructPtrType: o.StructPtrType,
			Fields:        o.Fields,
			Value:         val,
			RV:            rv,
			cacheInvalid:  false,
			IsSlice:       isSlice,
			Options:       o.Options,
		}

		o.OrgObject = o

		return out
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil {}", []interface{}{o}).PrintfError()
	return nil
}

func (o *Object) SetFromRV(rv reflect.Value) {
	if o != nil && o.Entity != nil {
		o.Value = rv.Interface()
		o.RV = rv
		return
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) PKKey() *Key {
	if o != nil && o.Entity != nil {
		if o.Entity.PKKey() != nil {
			return o.Entity.PKKey()
		}
	}
	return nil
}
func (o *Object) PKValue() []interface{} {
	if o != nil && o.Entity != nil {
		if o.Entity.PKKey() != nil {
			return o.KeyValue(o.Entity.PKKey())
		}
	}
	return nil
}

func (o *Object) KeyValue(key *Key) []interface{} {
	if o != nil && o.Entity != nil && key != nil {
		values, err := o.KeyFieldsValue(key)
		if err != nil {
			return nil
		}
		return values
	}
	return nil
}

func (o *Object) KeyValueString(key *Key) string {
	if o != nil && o.Entity != nil && key != nil {

		if key != nil && (key.Type == KEY_TYPE_PK || key.Type == KEY_TYPE_UK) {
			values, err := o.KeyFieldsValue(key)
			if err != nil {
				return ""
			}

			return key.Name + "[" + key.fieldsString + "] ('" + ArgsToString("','", values...) + "')"
		}

	}
	return ""
}

func (o *Object) KeysValueString() string {
	if o != nil && o.Entity != nil {

		var keysValues []string

		for _, key := range o.Entity.Keys() {
			if key != nil && (key.Type == KEY_TYPE_PK || key.Type == KEY_TYPE_UK) {
				keyValues := o.KeyValueString(key)
				keysValues = append(keysValues, keyValues)
			}
		}

		return strings.Join(keysValues, ",")
	}
	return ""
}

func (o *Object) FieldValue(field *Field) (value interface{}, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.FieldValue(field, o)
	}
	return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) FieldRV(field *Field) (value reflect.Value, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.FieldRV(field, o)
	}
	return reflect.Value{}, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

//func (o *Object) NotEmptyFieldRV(field *Field) (value reflect.Value, empty bool, err error) {
//	if o != nil && o.Entity != nil && field != nil {
//		value, err = o.FieldRV(field)
//		if err != nil {
//			return reflect.Value{}, false, err
//		} else {
//			if value.IsValid() && !value.IsNil() && !value.IsZero() {
//
//				var value2 reflect.Value
//
//				if value.Kind() == reflect.Ptr {
//					value2 = reflect.Indirect(value)
//				} else if value.Kind() == reflect.Interface {
//					value2 = reflect.Indirect(value)
//				}
//
//				if value.Kind() == reflect.Ptr {
//					value2 = reflect.Indirect(value)
//				}
//
//				switch value2.Kind() {
//				case reflect.Struct:
//					return value, true, nil
//
//				}
//
//			}
//		}
//	}
//	return reflect.Value{}, false, err
//}

func (o *Object) SetErrorValue(errors _err.Errors) (err error) {
	if o != nil && o.Entity != nil {
		if err = o.Entity.SetErrorValue(o, errors); err != nil {
			myErr := _err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err)
			myErr.Errors = errors
			return myErr
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetXmlNameValueSlice(xmlName *xml.Name) (err error) {
	if o != nil && o.Entity != nil {
		if o.Options.Global.OutFormat == "xml" {
			return o.Entity.SetXmlNameValueSlice(o, xmlName)
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetXmlNameValue(xmlName *xml.Name) (err error) {
	if o != nil && o.Entity != nil {
		if o.Options.Global.OutFormat == "xml" {
			return o.Entity.SetXmlNameValue(o, xmlName)
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetFieldRV(field *Field, val reflect.Value) (err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.SetFieldRV(field, o, val)
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) ZeroFieldRV(field *Field) (err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.ZeroFieldRV(field, o)
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) KeyFieldsValue(key *Key) (values []interface{}, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.KeyFieldsValue(key, o)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) KeyFieldsRV(key *Key) (values []reflect.Value, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.KeyFieldsRV(key, o)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) FieldsValue(fields Fields) (values []interface{}, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.FieldsValue(fields, o)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) FieldsRV(fields Fields) (values []reflect.Value, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.FieldsRV(fields, o)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetKeyFieldsRV(key *Key, values []reflect.Value) (err error) {
	if o != nil && o.Entity != nil && key != nil {

		if len(values) != len(key.fields) {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if len(values) != len(key.fields) {}").PrintfError()
		}

		for i, field := range key.fields {
			if err = o.SetFieldRV(field, values[i]); err != nil {
				return err
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil && key != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) ReferenceFieldsValue(reference *Reference) (values []interface{}, err error) {
	if o != nil && o.Entity != nil {
		return o.Entity.ReferenceFieldsValue(reference, o)
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetXmlNameValueFromTag() (err error) {
	if o != nil && o.Entity != nil {
		if o.Options.Global.OutFormat == "xml" {
			return o.Entity.SetXmlNameValueFromTag(o)
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) SetXmlNameValueFromTagSlice() (err error) {
	if o != nil && o.Entity != nil {
		if o.Options.Global.OutFormat == "xml" {
			return o.Entity.SetXmlNameValueFromTagSlice(o)
		}
		return nil

	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) Copy(copyAssociation bool, copyComposition bool) *Object {
	if o != nil && o.Entity != nil {
		out := &Object{
			//Key:        key,
			Entity:        o.Entity,
			StructType:    o.StructType,
			StructPtrType: o.StructPtrType,
			Fields:        o.Fields,
			Value:         o.Value,
			//PtrValue:     o.PtrValue,
			RV:           o.RV,
			cacheInvalid: false,
			IsSlice:      o.IsSlice,
		}

		if copyAssociation && len(o.Entity.associationMap) > 0 {
			out.CopyAssociationFrom(o)
		}

		if copyComposition && len(o.Entity.compositionMap) > 0 {
			out.CopyCompositionFrom(o)
		}

		return out
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
	return nil
}

func (o *Object) CopyAssociationFrom(from *Object) {
	if o != nil && o.Entity != nil {
		if len(from.Entity.associationMap) > 0 && len(from.AssociationMap) > 0 {
			o.AssociationMap = make(ObjectReferenceMap, len(from.AssociationMap))
			for key, object := range from.AssociationMap {
				o.AssociationMap[key] = object
			}
		}
		return
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) CopyCompositionFrom(from *Object) {
	if o != nil && o.Entity != nil {
		if len(from.Entity.compositionMap) > 0 && len(from.CompositionMap) > 0 {
			o.CompositionMap = make(ObjectReferenceMap, len(from.CompositionMap))
			for key, object := range from.CompositionMap {
				o.CompositionMap[key] = object
			}
		}
		return
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && o.Entity != nil {}", []interface{}{o}).PrintfError()
}

func (o *Object) ClearObjects() {
	if o != nil {
		o.Objects = make(Objects, 0)
	}
}

func (o *Object) AppendObject(obj *Object) {
	if o != nil && obj != nil {
		o.Objects = append(o.Objects, obj)
	}
}

func (o *Object) SetAssociationUnsafe(field *Field, obj *Object) error {
	if o != nil && field != nil && obj != nil && field.reference != nil {
		if o.AssociationMap == nil { // Создадим при первом использовании
			o.AssociationMap = make(ObjectReferenceMap, len(o.Entity.associationMap))
		}
		obj.Reference = field.reference
		o.AssociationMap[field.reference] = obj
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && field != nil && obj != nil && field.reference != nil {}", []interface{}{o, field, obj}).PrintfError()
}

func (o *Object) SetCompositionUnsafe(field *Field, obj *Object) error {
	if o != nil && field != nil && obj != nil && field.reference != nil {
		if o.CompositionMap == nil { // Создадим при первом использовании
			o.CompositionMap = make(ObjectReferenceMap, len(o.Entity.compositionMap))
		}
		obj.Reference = field.reference
		o.CompositionMap[field.reference] = obj
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if o != nil && field != nil && obj != nil && field.reference != nil {}", []interface{}{o, field, obj}).PrintfError()
}

func (o *Object) Lock() {
	if o != nil {
		o.mx.Lock()
	}
}

func (o *Object) RLock() {
	if o != nil {
		o.mx.RLock()
	}
}

func (o *Object) Unlock() {
	if o != nil {
		o.mx.Unlock()
	}
}

func (o *Object) RUnlock() {
	if o != nil {
		o.mx.RUnlock()
	}
}

func (o *Object) CacheInvalid() bool {
	if o != nil {
		return o.cacheInvalid
	}
	return false
}

func (o *Object) SetCacheInvalid(cacheInvalid bool) {
	if o != nil {
		if o.cacheInvalid == cacheInvalid {
			return
		}

		//_log.Info("START - before lock")
		o.mx.Lock()
		//_log.Info("START - after lock")
		o.SetCacheInvalidUnsafe(cacheInvalid)
		o.mx.Unlock()
		//_log.Info("START - after Unlock")
	}
}

func (o *Object) SetCacheInvalidUnsafe(cacheInvalid bool) {
	if o != nil {
		o.cacheInvalid = cacheInvalid
	}
}
