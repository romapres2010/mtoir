package meta

import (
	"strings"
	"sync"

	"encoding/xml"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

const ENTITY_NOT_FOUND = "ENTITY_NOT_FOUND"

type Entity struct {
	Type            interface{} `yaml:"-" json:"-" xml:"-"`                                                                // Тип сущности в виде пустого интерфейса - для настройки валидатора
	Status          string      `yaml:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`                   // Статус ENABLED, DEPRECATED, ...
	StorageName     string      `yaml:"storage_name,omitempty" json:"storage_name,omitempty" xml:"storage_name,omitempty"` // Имя хранилища для сущности
	Name            string      `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`                         // Имя сущности
	InheritFromName string      `yaml:"inherit_from,omitempty" json:"inherit_from,omitempty" xml:"inherit_from,omitempty"` // Наследовать характеристики с другой сущности с возможностью переопределения
	SkipCache       bool        `yaml:"skip_cache,omitempty" json:"skip_cache,omitempty" xml:"skip_cache,omitempty"`       // Принудительно отключить Cache, по умолчанию включено
	Embed           bool        `yaml:"embed" json:"embed" xml:"embed"`                                                    // Признак, что объект всегда встраивается

	Alias         Alias         `yaml:"alias,omitempty" json:"alias,omitempty" xml:"alias,omitempty"`                   // Дополнительные имена, коды и описание
	DbStorage     DbStorage     `yaml:"db_storage,omitempty" json:"db_storage,omitempty" xml:"db_storage,omitempty"`    // Параметры хранения в БД
	Tag           Tag           `yaml:"tag,omitempty" json:"tag,omitempty" xml:"tag,omitempty"`                         // Имена различных тегов для парсинга данных
	Modify        Modify        `yaml:"modify,omitempty" json:"modify,omitempty" xml:"modify,omitempty"`                // Разрешенные операции с сущностью
	Exprs         Exprs         `yaml:"expressions,omitempty" json:"expressions,omitempty" xml:"expressions,omitempty"` // Выражение для вычисления или проверки
	exprsByAction ExprsByAction // Формулы вычисления сущности по Action

	KeysDef Keys    `yaml:"keys,omitempty" json:"keys,omitempty" xml:"keys,omitempty"` // Ключи сущности из определения
	keys    Keys    // Ключи сущности, созданные и доступные
	keysUk  Keys    // Ключи сущности, созданные и доступные
	pkKey   *Key    // Первичный ключ сущности
	keysMap KeysMap // Ключи сущности - map для быстрого поиска

	ReferencesDef References    `yaml:"references,omitempty" json:"references,omitempty" xml:"references,omitempty"` // Ссылки на родительские и дочерние сущности
	references    References    // Ссылки на родительские и дочерние сущности, созданные и доступные
	referencesMap ReferencesMap // Ссылки на родительские и дочерние сущности - map для быстрого поиска
	referencesBy  References    // TODO - Доделать - Кто ссылается на нас

	Fields          Fields    `yaml:"fields,omitempty" json:"fields,omitempty" xml:"fields,omitempty"` // Поля сущности, включая ссылки FK и отношения M:1
	fieldsMap       FieldsMap // Поля сущности, включая ссылки FK и отношения M:1
	structFields    Fields    // Поля сущности, только допустимые к обработке
	structFieldsMap FieldsMap // Поля сущности, только допустимые к обработке

	Definition *Definition `yaml:"definition,omitempty" json:"definition,omitempty" xml:"definition,omitempty"` // Определение сущности

	dbNameMap    FieldsMap // Для быстрого поиска и проверки дублей мета модели
	jsonNameMap  FieldsMap // Для быстрого поиска и проверки дублей мета модели
	xmlNameMap   FieldsMap // Для быстрого поиска и проверки дублей мета модели
	yamlNameMap  FieldsMap // Для быстрого поиска и проверки дублей мета модели
	xlsNameMap   FieldsMap // Для быстрого поиска и проверки дублей мета модели
	exprNameMap  FieldsMap // Для быстрого поиска и проверки дублей мета модели
	keyFieldsMap FieldsMap // Поля всех ключей сущности

	associationMap FieldsMap // Поля, которые имеют отношения 1:M
	compositionMap FieldsMap // Поля, которые имеют отношения 1:M

	fieldsExprs         Exprs         // Формулы вычисления полей
	fieldsExprsByAction ExprsByAction // Формулы вычисления полей по Action

	validationRules map[string]string // Правила проверки

	inheritFrom             []*Entity  // Наследовать характеристики с возможностью переопределения
	inheritTo               []*Entity  // Сущности, которые от нас наследованы
	meta                    *Meta      // Метаданные в состав которых входит сущность
	typeCache               *TypeCache // кэш типов для создания объектов
	defTypeCacheKey         string     // ключ кэша для полного набора полей, включает через ',' все разрешенные поля
	defTypeCacheKeyEmptyTag string     // ключ кэша для полного набора полей (без tag), включает через ',' все разрешенные поля
	defTypeCacheKeyEmptyRef string     // ключ кэша для полного набора полей (без tag для Reference), включает через ',' все разрешенные поля
	isInit                  bool       // признак, что инициация успешная

	// Системные поля
	xmlNameField      *Field // специальное "XMLName" поле для парсинга XML
	validField        *Field // специальное поле для проверки, что данные валидные
	errorsField       *Field // специальное поле для записи ошибок валидации
	cacheInvalidField *Field // специальное поле для проверки, что данные в cache валидные
	mxField           *Field // специальное поле для RWMutex

	mx sync.RWMutex
}

func (meta *Meta) newEntity(name string) *Entity {
	if meta != nil {

		entity := &Entity{
			Status: STATUS_ENABLED,
			Name:   name,
			meta:   meta,
		}

		return entity
	}

	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil {}", []interface{}{meta}).PrintfError()
	return nil
}

func (entity *Entity) clear() {
	if entity != nil {
		entity.Type = nil
		entity.Status = ""
		entity.StorageName = ""
		entity.Name = ""
		entity.InheritFromName = ""
		entity.SkipCache = false
		entity.Alias = Alias{}
		entity.DbStorage = DbStorage{}
		entity.Tag = Tag{}
		entity.Modify = Modify{}
		entity.KeysDef = nil
		entity.keys = nil
		entity.ReferencesDef = nil
		entity.references = nil
		entity.Fields = nil
		entity.Definition = nil

		entity.clearInternal()
	}
}

func (entity *Entity) clearInternal() {
	if entity != nil {
		entity.pkKey = nil
		entity.keysMap = nil
		entity.referencesMap = nil
		entity.referencesBy = nil
		entity.fieldsMap = nil
		entity.structFields = nil
		entity.structFieldsMap = nil
		entity.dbNameMap = nil
		entity.jsonNameMap = nil
		entity.xmlNameMap = nil
		entity.yamlNameMap = nil
		entity.xlsNameMap = nil
		entity.exprNameMap = nil
		entity.keyFieldsMap = nil
		entity.associationMap = nil
		entity.compositionMap = nil
		entity.fieldsExprs = nil
		entity.fieldsExprsByAction = nil
		entity.validationRules = nil
		entity.inheritFrom = nil
		entity.inheritTo = nil
		entity.meta = nil
		entity.typeCache = nil
		entity.defTypeCacheKey = ""
		entity.defTypeCacheKeyEmptyTag = ""
		entity.isInit = false
		entity.xmlNameField = nil
		entity.validField = nil
		entity.errorsField = nil
		entity.cacheInvalidField = nil
	}
}

func (entity *Entity) PKKey() *Key {
	if entity != nil {
		return entity.pkKey
	}
	return nil
}

func (entity *Entity) References() References {
	if entity != nil {
		return entity.references
	}
	return nil
}

func (entity *Entity) Keys() Keys {
	if entity != nil {
		return entity.keys
	}
	return nil
}

func (entity *Entity) KeysUK() Keys {
	if entity != nil {
		return entity.keysUk
	}
	return nil
}

func (entity *Entity) StructFields() Fields {
	if entity != nil {
		return entity.structFields
	}
	return nil
}

func (entity *Entity) HasAssociations() bool {
	if entity != nil {
		return entity.associationMap != nil && len(entity.associationMap) > 0
	}
	return false
}

func (entity *Entity) AssociationMap() FieldsMap {
	if entity != nil {
		return entity.associationMap
	}
	return nil
}

func (entity *Entity) GetAssociation(name string) *Reference {
	if entity != nil && entity.associationMap != nil {
		ref, ok := entity.associationMap[name]
		if ok && ref != nil && ref.reference != nil {
			return ref.reference
		}
	}
	return nil
}

func (entity *Entity) HasCompositions() bool {
	if entity != nil {
		return entity.compositionMap != nil && len(entity.compositionMap) > 0
	}
	return false
}

func (entity *Entity) CompositionMap() FieldsMap {
	if entity != nil {
		return entity.compositionMap
	}
	return nil
}

func (entity *Entity) GetComposition(name string) *Reference {
	if entity != nil && entity.compositionMap != nil {
		ref, ok := entity.compositionMap[name]
		if ok && ref != nil && ref.reference != nil {
			return ref.reference
		}
	}
	return nil
}

func (entity *Entity) FieldsExprsByAction(action ExprAction) *Exprs {
	if entity != nil && entity.fieldsExprsByAction != nil {
		return entity.fieldsExprsByAction[action]
	}
	return nil
}

func (entity *Entity) ExprsByAction(action ExprAction) *Exprs {
	if entity != nil && entity.exprsByAction != nil {
		return entity.exprsByAction[action]
	}
	return nil
}

func (entity *Entity) FieldsExprs() Exprs {
	if entity != nil {
		return entity.fieldsExprs
	}
	return nil
}

func (entity *Entity) addReferenceBy(ref *Reference) {
	if entity != nil && ref != nil {
		entity.referencesBy = append(entity.referencesBy, ref)
	}
}

func (entity *Entity) AddPkFieldsToMap(fieldsMap *FieldsMap) {
	if entity != nil && fieldsMap != nil {
		if entity.pkKey != nil {
			entity.pkKey.AddFieldsToMap(fieldsMap)
		}
	}
}

func (entity *Entity) AddSystemFieldsToMap(fieldsMap *FieldsMap) {
	if entity != nil && fieldsMap != nil {
		if entity.xmlNameField != nil {
			(*fieldsMap)[entity.xmlNameField.Name] = entity.xmlNameField
		}
		if entity.errorsField != nil {
			(*fieldsMap)[entity.errorsField.Name] = entity.errorsField
		}
		if entity.validField != nil {
			(*fieldsMap)[entity.validField.Name] = entity.validField
		}
		//if entity.cacheInvalidField != nil {
		//	(*argsFieldsMap)[entity.cacheInvalidField.Name] = entity.cacheInvalidField
		//}
		//if entity.mxField != nil {
		//	(*argsFieldsMap)[entity.mxField.Name] = entity.mxField
		//}
	}
}

func (entity *Entity) FieldsMap() FieldsMap {
	if entity != nil && entity.fieldsMap != nil {
		return entity.fieldsMap.Copy()
	}
	return nil
}

func (entity *Entity) StructFieldsMap() FieldsMap {
	if entity != nil && entity.structFieldsMap != nil {
		return entity.structFieldsMap.Copy()
	}
	return nil
}

func (entity *Entity) ValidationRulesType() (map[string]string, interface{}) {
	if entity != nil {
		//entity.mx.RLock()
		//defer entity.mx.RUnlock()

		return entity.validationRules, entity.Type
	}
	return nil, nil
}

func (entity *Entity) DefTypeCacheKey() string {
	if entity != nil {
		//entity.mx.RLock()
		//defer entity.mx.RUnlock()

		return entity.defTypeCacheKey
	}
	return ""
}

func (entity *Entity) DefTypeCacheKeyEmptyTag() string {
	if entity != nil {
		//entity.mx.RLock()
		//defer entity.mx.RUnlock()

		return entity.defTypeCacheKeyEmptyTag
	}
	return ""
}
func (entity *Entity) DefTypeCacheKeyEmptyRef() string {
	if entity != nil {
		//entity.mx.RLock()
		//defer entity.mx.RUnlock()

		return entity.defTypeCacheKeyEmptyRef
	}
	return ""
}

func (entity *Entity) ValidationRules() map[string]string {
	if entity != nil {
		//entity.mx.RLock()
		//defer entity.mx.RUnlock()

		return entity.validationRules
	}
	return nil
}

func (entity *Entity) GetXMLName() *Field {
	if entity != nil {
		return entity.xmlNameField
	}
	return nil
}

func (entity *Entity) GetErrors() *Field {
	if entity != nil {
		return entity.errorsField
	}
	return nil
}

func (entity *Entity) GetTag(format string, useNameAsDefault bool) string {
	if entity != nil {
		tag := entity.Tag.GetTag(format)
		if tag == "" && useNameAsDefault {
			tag = entity.Name
		}

		return tag
	}
	return ""
}
func (entity *Entity) GetTagName(format string, useNameAsDefault bool) string {
	if entity != nil {
		tag := entity.Tag.GetName(format)
		if tag == "" && useNameAsDefault {
			tag = entity.Name
		}

		return tag
	}
	return ""
}

func (entity *Entity) GetXmlNameFromTag(useNameAsDefault bool) *xml.Name {
	if entity != nil {

		xmlTag := entity.GetTag("xml", useNameAsDefault)
		if strings.ContainsAny(xmlTag, ">") {
			// Если указана вложенность элементов, то извлечь последний до >
			xmlTagIndex := strings.LastIndex(xmlTag, ">")
			if xmlTagIndex != -1 {
				xmlTag = xmlTag[xmlTagIndex+1:]
			}
		}
		return &xml.Name{
			Local: xmlTag,
			Space: entity.GetTag("xmlSpace", false),
		}
	}
	//return xml.Name{}
	return nil
}

func (entity *Entity) FieldByTagName(format, tag string) *Field {
	if entity != nil {
		entity.mx.RLock()
		defer entity.mx.RUnlock()

		switch format {
		case "name":
			if entity.structFields != nil {
				if field, ok := entity.structFieldsMap[tag]; ok {
					return field
				}
			}
		case "db":
			if entity.dbNameMap != nil {
				if field, ok := entity.dbNameMap[tag]; ok {
					return field
				}
			}
		case "json":
			if entity.jsonNameMap != nil {
				if field, ok := entity.jsonNameMap[tag]; ok {
					return field
				}
			}
		case "xml":
			if entity.xmlNameMap != nil {
				if field, ok := entity.xmlNameMap[tag]; ok {
					return field
				}
			}
		case "yaml":
			if entity.yamlNameMap != nil {
				if field, ok := entity.yamlNameMap[tag]; ok {
					return field
				}
			}
		case "xls":
			if entity.xlsNameMap != nil {
				if field, ok := entity.xlsNameMap[tag]; ok {
					return field
				}
			}
		case "expr":
			if entity.exprNameMap != nil {
				if field, ok := entity.exprNameMap[tag]; ok {
					return field
				}
			}
		default:
			return nil
		}
	}
	return nil
}

func (entity *Entity) fieldByNameUnsafe(name string) *Field {
	if entity.fieldsMap != nil {
		return entity.fieldsMap.Get(name)
	} else if entity.Fields != nil {
		return entity.Fields.getUnsafe(name)
	} else {
		return nil
	}
}

func (entity *Entity) keyDefByNameUnsafe(name string) *Key {
	//if entity.keysMap != nil {
	//	return entity.keysMap.Get(name)
	//} else
	if entity.KeysDef != nil {
		return entity.KeysDef.getUnsafe(name)
	} else {
		return nil
	}
}

func (entity *Entity) referenceDefByNameUnsafe(name string) *Reference {
	//if entity.referencesMap != nil {
	//	return entity.referencesMap.Get(name)
	//} else
	if entity.ReferencesDef != nil {
		return entity.ReferencesDef.getUnsafe(name)
	} else {
		return nil
	}
}

func (entity *Entity) FieldByName(name string) *Field {
	if entity != nil {
		entity.mx.RLock()
		defer entity.mx.RUnlock()

		return entity.fieldByNameUnsafe(name)
	} else {
		return nil
	}
}

func (entity *Entity) FieldDisplayName(name string) string {
	if entity != nil {
		entity.mx.RLock()
		defer entity.mx.RUnlock()

		return entity.fieldsMap.GetDisplayName(name)
	} else {
		return FIELD_NOT_FOUND
	}
}

func (entity *Entity) FieldFullName(name string) string {
	if entity != nil {
		entity.mx.RLock()
		defer entity.mx.RUnlock()

		return entity.fieldsMap.GetFullName(name)
	} else {
		return FIELD_NOT_FOUND
	}
}

func (entity *Entity) KeyByName(name string) *Key {
	if entity != nil {
		return entity.keysMap.Get(name)
		//if v, ok := entity.keysMap[name]; ok {
		//	return v
		//} else {
		//	return nil
		//}
	}
	return nil
}

func (entity *Entity) ReferenceByName(name string) *Reference {
	if entity != nil {
		return entity.referencesMap.Get(name)
		//if v, ok := entity.referencesMap[name]; ok {
		//	return v
		//} else {
		//	return nil
		//}
	}
	return nil
}

//// AllKeyFieldsMap - ограниченный набор полей - только ключа
//func (entity *Entity) AllKeyFieldsMap() argsFieldsMap {
//	if entity != nil {
//		return entity.keyFieldsMap
//	}
//	return nil
//}

//// PKFieldsMap - ограниченный набор полей - только PK
//func (entity *Entity) PKFieldsMap() argsFieldsMap {
//	if entity != nil {
//		if entity.pkKey != nil {
//			return entity.pkKey.argsFieldsMap
//		}
//	}
//	return nil
//}

// PKFieldsName - ограниченный набор полей - только PK
func (entity *Entity) PKFieldsName() []string {
	if entity != nil {
		if entity.pkKey != nil {
			return entity.pkKey.FieldsName
		}
	}
	return nil
}

func (entity *Entity) addField(field *Field) {
	if entity != nil && field != nil {
		entity.Fields = append(entity.Fields, field)
	}
}

func (entity *Entity) addReferenceDef(reference *Reference) {
	if entity != nil && reference != nil {
		entity.ReferencesDef = append(entity.ReferencesDef, reference)
	}
}

func (entity *Entity) addKeyDef(key *Key) {
	if entity != nil && key != nil {
		entity.KeysDef = append(entity.KeysDef, key)
	}
}

var entityNotFound = Entity{
	Name: ENTITY_NOT_FOUND,
	Alias: Alias{
		DisplayName: "ENTITY_NOT_FOUND",
	},
	DbStorage: DbStorage{
		SchemaName: "ENTITY_NOT_FOUND",
		TableName:  "ENTITY_NOT_FOUND",
	},
	Tag: Tag{
		Db:   "ENTITY_NOT_FOUND",
		Json: "ENTITY_NOT_FOUND",
		Xml:  "ENTITY_NOT_FOUND",
		Yaml: "ENTITY_NOT_FOUND",
		Xls:  "ENTITY_NOT_FOUND",
	},
	fieldsMap: FieldsMap{
		FIELD_NOT_FOUND: &fieldNotFound,
	},
}
