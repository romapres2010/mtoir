package meta

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

const TYPE_CACHE_KEY_PREFIX_ALL = "ALL"
const TYPE_CACHE_KEY_PREFIX_EMPTY_TAG = "EMPTY_TAG"
const TYPE_CACHE_KEY_PREFIX_EMPTY_REF = "EMPTY_REF"

func (entity *Entity) StructOf(restrictFields FieldsMap, validate, processJson, processDb, processXml, processYaml, processXls bool, addXmlName bool, processReference bool) (structType reflect.Type, err error) {
	if entity != nil {

		var cnt int
		var errDetail []string
		var structFields []reflect.StructField
		var isRestrictFields = restrictFields != nil && len(restrictFields) > 0

		// Функция восстановления после паники в reflect
		defer func() {
			if r := recover(); r != nil {
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "StructOf", entity.Name, structFields)
			}
		}()

		if isRestrictFields {
			structFields = make([]reflect.StructField, 0, len(restrictFields))
		} else {
			structFields = make([]reflect.StructField, 0, len(entity.structFields))
		}

		// Создадим структуру для разбора
		for _, field := range entity.structFields {

			// Ограничить поля определенным списком
			if isRestrictFields {
				if _, ok := restrictFields[field.Name]; !ok {
					continue
				}
			}

			if field.reflectType == nil {
				return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - empty reflectType", field.entity.Name, field.Name))
			}

			cnt++ // порядковый номер поля при добавлении структуру

			structTag := ""
			{ // structTag

				// Для исключения бесконечного цикла для парсинге reference
				if !(field.reference != nil && !processReference) {

					if processJson {
						if field.Tag.Json != "" {
							structTag = structTag + ` json:"` + field.Tag.Json + `"`
						}
					} else {
						structTag = structTag + ` json:"-"` // принудительно выключаем tag
					}

					if processXml {
						if field.Tag.Xml != "" {
							structTag = structTag + ` xml:"` + field.Tag.Xml + `"`
							//
							//if field.Name == "XMLName" {
							//    if addXmlName { // Если запрещено добавлять tag для XMLName
							//        structTag = structTag + ` xml:"` + field.Tag.Xml + `"`
							//    }
							//} else {
							//    structTag = structTag + ` xml:"` + field.Tag.Xml + `"`
							//}
						}
					} else {
						structTag = structTag + ` xml:"-"` // принудительно выключаем tag
					}

					if processYaml {
						if field.Tag.Yaml != "" {
							structTag = structTag + ` yaml:"` + field.Tag.Yaml + `"`
						}
					} else {
						structTag = structTag + ` yaml:"-"` // принудительно выключаем tag
					}

					if processDb {
						if field.Tag.Db != "" {
							structTag = structTag + ` db:"` + field.Tag.Db + `"`
							if field.Tag.Sql != "" {
								structTag = structTag + ` sql:"` + field.Tag.Sql + `"`
							}
						}
					} else {
						structTag = structTag + ` db:"-"` // принудительно выключаем tag
					}

					if processXls {
						if field.Tag.Xls != "" && field.Tag.Xls != "-" {
							structTag = structTag + ` title:"` + field.Tag.Xls + `"`
							//structTag = structTag + ` cell:"` + strconv.Itoa(cnt) + `"`
							structTag = structTag + ` cell:"auto"`
							//if i == 0 {
							// Имя sheet на уровне поля имеет приоритет над Entity
							if field.Tag.XlsSheet != "" && field.Tag.XlsSheet != "-" {
								structTag = structTag + ` sheet:"` + field.Tag.XlsSheet + `"`
							} else {
								structTag = structTag + ` sheet:"` + entity.Tag.XlsSheet + `"`
							}
							//}
						}
					} else {
						structTag = structTag + ` title:"-"` // принудительно выключаем tag
					}

				} else {
					// Для исключения бесконечного цикла для парсинге reference
					structTag = structTag + ` json:"-"`  // принудительно выключаем tag
					structTag = structTag + ` xml:"-"`   // принудительно выключаем tag
					structTag = structTag + ` yaml:"-"`  // принудительно выключаем tag
					structTag = structTag + ` title:"-"` // принудительно выключаем tag
				}

				if validate {
					if field.ValidateRule != "" {
						structTag = structTag + ` validate:"` + field.ValidateRule + `"`
					}
				}

				if field.Tag.Expr != "" {
					structTag = structTag + ` expr:"` + field.Tag.Expr + `"`
				}

				if field.Format != "" {
					structTag = structTag + ` format:"` + field.Format + `"`
				}

			} // structTag

			structField := reflect.StructField{
				Name: field.Name,
				Type: field.reflectType,
				Tag:  reflect.StructTag(strings.TrimSpace(structTag)),
			}

			// DO_NOT_REMOVE
			//if field.InternalType == FIELD_TYPE_ASSOCIATION {
			//    if toEntity := field.reference.toEntity; toEntity != nil {
			//        // ищем определение структуры в cache
			//        refTypeCacheEntry, cacheHit := toEntity.TypeCache(toEntity.defTypeCacheKey)
			//        if cacheHit {
			//            // Добавить реальный тип поля как *struct{} вместо interface{}
			//            field.reflectType = refTypeCacheEntry.structPtrType
			//            structField.Type = field.reflectType
			//        }
			//    }
			//}
			// DO_NOT_REMOVE

			structFields = append(structFields, structField)
		}

		if len(errDetail) > 0 {
			myErr := _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - has incorrect filed type", entity.Name))
			myErr.Details = errDetail
			return nil, myErr
		}

		// Создадим тип структуры
		structType = reflect.StructOf(structFields)

		// для каждого поля найдем его индекс, будем использовать для быстрого доступа к значению поля в структуре
		for _, field := range entity.structFields {
			if structField, ok := structType.FieldByName(field.Name); ok {
				field.indexMap[structType] = structField.Index
			} else {
				// поля можем не быть в структуре - это нормальная ситуация, если запрашивают не все поля
				//return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' not found - incorrect create reflect.StructOf", entity.Name, field.Name))
			}
		}

		return structType, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
}

func (entity *Entity) TypeCacheKey(prefix string, fields FieldsMap) string {
	if entity != nil {

		// Если запросили все поля fields == nil, то вернем ранее сохраненный ключ
		if prefix == TYPE_CACHE_KEY_PREFIX_ALL && fields == nil && entity.defTypeCacheKey != "" {
			return entity.defTypeCacheKey
		}

		// Если запросили все поля fields == nil, то вернем ранее сохраненный ключ
		if prefix == TYPE_CACHE_KEY_PREFIX_EMPTY_TAG && fields == nil && entity.defTypeCacheKeyEmptyTag != "" {
			return entity.defTypeCacheKeyEmptyTag
		}

		fieldSlice := make([]string, 0, len(entity.structFields)+2)

		fieldSlice = append(fieldSlice, entity.Name)
		fieldSlice = append(fieldSlice, prefix)

		for _, field := range entity.structFields {
			if fields != nil {
				if _, ok := fields[field.Name]; ok {
					fieldSlice = append(fieldSlice, field.Name)
				}
			} else {
				fieldSlice = append(fieldSlice, field.Name)
			}
		}

		return strings.Join(fieldSlice, ",")
	}
	return ""
}

func (entity *Entity) NewStruct(fields FieldsMap, typeCacheKey string) (cacheHit bool, row *Object, err error) {
	if entity != nil {

		var typeCacheEntry *TypeCacheEntry

		// ищем определение структуры в cache
		typeCacheEntry, cacheHit = entity.TypeCache(typeCacheKey)

		if !cacheHit {

			// Создать тип структуры
			_log.Info("Not Found Entity Object Type Cache - create new one: entityName, typeCacheKey", entity.Name, typeCacheKey)
			newRowType, err := entity.StructOf(fields, false, true, true, true, true, true, false, true)
			if err != nil {
				return false, nil, err
			}

			// Использовать те типы, которые вернулись из кэша, могла возникнуть коллизия две горутины одновременно запросили новый тип
			typeCacheEntry = entity.SetTypeCache(typeCacheKey, newRowType)
		} else {
			//_log.Debug("Found Entity Object Type in Cache: entityName, typeCacheKey", entity.Name, typeCacheKey)
		}

		rowPtrRV := reflect.New(typeCacheEntry.structType)

		row = newObject(entity, fields, rowPtrRV.Interface(), rowPtrRV, typeCacheEntry.structType, typeCacheEntry.structPtrType, false)

		return cacheHit, row, nil
	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
}

func (entity *Entity) NewSliceAny(fields FieldsMap, typeCacheKey string, len, cap int) (cacheHit bool, slice *Object, err error) {
	if entity != nil {

		var typeCacheEntry *TypeCacheEntry

		// ищем определение структуры в cache
		typeCacheEntry, cacheHit = entity.TypeCache(typeCacheKey)

		if !cacheHit {

			// Создать тип структуры
			_log.Info("Not Found Entity Object Type Cache - create new one: entityName, typeCacheKey", entity.Name, typeCacheKey)
			newRowType, err := entity.StructOf(fields, false, true, true, true, true, true, false, true)
			if err != nil {
				return false, nil, err
			}

			// Использовать те типы, которые вернулись из кэша, могла возникнуть коллизия две горутины одновременно запросили новый тип
			typeCacheEntry = entity.SetTypeCache(typeCacheKey, newRowType)
		} else {
			//_log.Debug("Found Entity Object Type in Cache: entityName, typeCacheKey", entity.Name, typeCacheKey)
		}

		sliceRV := reflect.MakeSlice(typeCacheEntry.sliceAnyType, len, cap)
		slicePtrRV := reflect.New(typeCacheEntry.sliceAnyType)
		slicePtrRV.Elem().Set(sliceRV)

		slice = newObject(entity, fields, slicePtrRV.Interface(), slicePtrRV, typeCacheEntry.structType, typeCacheEntry.structPtrType, true)

		return cacheHit, slice, nil

	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
}

func (entity *Entity) NewSlice(fields FieldsMap, typeCacheKey string, len, cap int) (cacheHit bool, slice *Object, err error) {
	if entity != nil {

		var typeCacheEntry *TypeCacheEntry

		// ищем определение структуры в cache
		typeCacheEntry, cacheHit = entity.TypeCache(typeCacheKey)

		if !cacheHit {

			// Создать тип структуры
			_log.Info("Not Found Entity Object Type Cache - create new one: entityName, typeCacheKey", entity.Name, typeCacheKey)
			newRowType, err := entity.StructOf(fields, false, true, true, true, true, true, false, true)
			if err != nil {
				return false, nil, err
			}

			// Использовать те типы, которые вернулись из кэша, могла возникнуть коллизия две горутины одновременно запросили новый тип
			typeCacheEntry = entity.SetTypeCache(typeCacheKey, newRowType)
		} else {
			//_log.Debug("Found Entity Object Type in Cache: entityName, typeCacheKey", entity.Name, typeCacheKey)
		}

		sliceRV := reflect.MakeSlice(typeCacheEntry.sliceType, len, cap)
		slicePtrRV := reflect.New(typeCacheEntry.sliceType)
		slicePtrRV.Elem().Set(sliceRV)

		slice = newObject(entity, fields, slicePtrRV.Interface(), slicePtrRV, typeCacheEntry.structType, typeCacheEntry.structPtrType, true)

		return cacheHit, slice, nil

	}
	return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
}

func (entity *Entity) SetTypeCache(key string, inRowType reflect.Type) (typeCacheEntry *TypeCacheEntry) {
	if entity != nil && entity.typeCache != nil {
		return entity.typeCache.Set(key, inRowType)
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && entity.typeCache != nil {}", []interface{}{entity}).PrintfError()
	return nil
}

func (entity *Entity) TypeCache(key string) (*TypeCacheEntry, bool) {
	if entity != nil && entity.typeCache != nil {
		return entity.typeCache.Get(key)
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && entity.typeCache != nil {}", []interface{}{entity}).PrintfError()
	return nil, false
}

type TypeCacheEntry struct {
	structType    reflect.Type // struct{}
	structPtrType reflect.Type // *struct{}
	sliceType     reflect.Type // []*struct{}
	sliceAnyType  reflect.Type // []interface{}
}

type TypeCache struct {
	cache map[string]*TypeCacheEntry // различные варианты структур, в зависимости от ограничения на поля

	mx sync.RWMutex
}

func NewTypeCache() *TypeCache {
	typeCache := &TypeCache{
		cache: make(map[string]*TypeCacheEntry),
	}

	return typeCache
}

func (typeCache *TypeCache) Set(key string, inRowType reflect.Type) (typeCacheEntry *TypeCacheEntry) {
	if typeCache != nil && typeCache.cache != nil && inRowType != nil {
		typeCache.mx.Lock()
		defer typeCache.mx.Unlock()

		// Однажды создав тип по ключу, его нельзя менять, иначе разные горутины могут пересоздать и нельзя будет копировать объекты
		entry, ok := typeCache.cache[key]
		if !ok {
			entry = &TypeCacheEntry{
				structType:    inRowType,
				structPtrType: reflect.PointerTo(inRowType),
				//sliceType: reflect.SliceOf(inRowType),
				sliceType:    reflect.SliceOf(reflect.PointerTo(inRowType)),
				sliceAnyType: FIELD_TYPE_SLICE_ANY_RT,
			}
			typeCache.cache[key] = entry
		}

		return entry
	}
	return nil
}

func (typeCache *TypeCache) Get(key string) (*TypeCacheEntry, bool) {
	if typeCache != nil && typeCache.cache != nil {
		typeCache.mx.RLock()
		defer typeCache.mx.RUnlock()

		if entry, ok := typeCache.cache[key]; ok {
			return entry, ok
		}
	}
	return nil, false
}
