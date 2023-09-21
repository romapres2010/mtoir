package meta

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"encoding/xml"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

const FIELD_NOT_FOUND = "FIELD_NOT_FOUND"

type Field struct {
	Status          string `yaml:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`                      // Статус поля ENABLED, DEPRECATED, ...
	Order           int    `yaml:"order,omitempty" json:"order,omitempty" xml:"order,omitempty"`                         // Порядковый номер поля в сущности
	Required        bool   `yaml:"required,omitempty" json:"required,omitempty" xml:"required,omitempty"`                // Признак, что поле обязательное в сущности, это не означает что оно должно быть заполнено
	Name            string `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`                            // Имя поля
	InheritFromName string `yaml:"inherit_from,omitempty" json:"inherit_from,omitempty" xml:"inherit_from,omitempty"`    // Наследовать характеристики
	CopyFromName    string `yaml:"copy_from,omitempty" json:"copy_from,omitempty" xml:"copy_from,omitempty"`             // Скопировать характеристики поля с другой сущности
	InternalType    string `yaml:"internal_type,omitempty" json:"internal_type,omitempty" xml:"internal_type,omitempty"` // Тип поля - внутренний
	Format          string `yaml:"format,omitempty" json:"format,omitempty" xml:"format,omitempty"`                      // Формат поля
	ValidateRule    string `yaml:"validate_rule,omitempty" json:"validate_rule,omitempty" xml:"validate_rule,omitempty"` // Правила валидации поля
	System          bool   `yaml:"system,omitempty" json:"system,omitempty" xml:"system,omitempty"`                      // Признак, системного поля

	Alias     Alias     `yaml:"alias,omitempty" json:"alias,omitempty" xml:"alias,omitempty"`                   // Дополнительные имена, коды и описание
	DbStorage DbStorage `yaml:"db_storage,omitempty" json:"db_storage,omitempty" xml:"db_storage,omitempty"`    // Параметры хранения в БД
	Tag       Tag       `yaml:"tag,omitempty" json:"tag,omitempty" xml:"tag,omitempty"`                         // Имена различных тегов для парсинга данных
	Modify    Modify    `yaml:"modify,omitempty" json:"modify,omitempty" xml:"modify,omitempty"`                // Разрешенные операции с полем
	Exprs     Exprs     `yaml:"expressions,omitempty" json:"expressions,omitempty" xml:"expressions,omitempty"` // Выражение для вычисления или проверки

	entity      *Entity                // Сущность к которой относится поле
	copyFrom    *Field                 // Копировать характеристики с возможностью переопределения
	inheritFrom *Field                 // Наследовать характеристики
	reference   *Reference             // Поле может быть виртуальным, например как ссылка на родительскую сущность
	references  References             // Поле может входить в состав нескольких ссылок
	reflectType reflect.Type           // Go тип данных поля
	indexMap    map[reflect.Type][]int // Одно и то же поле может иметь разный индекс в структурах с разным составом полей
	isInit      bool                   // признак, что инициация успешная
}

func (entity *Entity) newField(name string) *Field {
	if entity != nil {

		field := &Field{
			Status: STATUS_ENABLED,
			Order:  len(entity.Fields) + 1,
			Name:   name,
			entity: entity,
		}

		return field
	}

	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
	return nil
}

func (field *Field) clearInternal() {
	if field != nil {
		field.inheritFrom = nil
		field.copyFrom = nil
		field.reference = nil
		field.references = nil
		field.reflectType = nil
		field.indexMap = nil
		field.isInit = false
	}
}

func (field *Field) init(entity *Entity) (err error) {
	if field != nil && entity != nil {

		if field.isInit {
			return nil
		}

		field.entity = entity // TODO - определить место присвоения

		if field.Status == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - empty 'status'", field.entity.Name, field.Name))
		}

		if field.Name == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - empty 'name'", field.entity.Name, field.Name))
		}

		if !isValidFieldName(field.Name) {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - invalid field name", field.entity.Name, field.Name))
		}

		if field.CopyFromName != "" {
			if err = field.copyFromFieldName(field.CopyFromName); err != nil {
				return err
			}
		}

		field.Tag.init()

		field.indexMap = make(map[reflect.Type][]int)

		if field.reflectType, err = field.getReflectType(); err != nil {
			return err
		}

		// TODO - если не указан tag db, заполнять из db_name?

		field.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if field != nil && entity != nil {}", []interface{}{field, entity}).PrintfError()
}

func (field *Field) copyFromFieldName(name string) (err error) {
	if field != nil && field.entity != nil && name != "" {

		if name != "" {
			fromFieldNameSplit := strings.Split(name, ".")

			// Формат должен быть 'EntityName'.'FieldName',
			if len(fromFieldNameSplit) != 2 {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - 'copy_from'=['%s'] must be 'EntityName.FieldName'", field.entity.Name, field.Name, name))
			}

			// Найти сущность
			fromEntity := field.entity.meta.GetEntityUnsafe(fromFieldNameSplit[0])
			if fromEntity == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - 'copy_from'='%s' 'EntityName'='%s' was not found", field.entity.Name, field.Name, name, fromFieldNameSplit[0]))
			}

			// Найти поле
			if fromField := fromEntity.fieldByNameUnsafe(fromFieldNameSplit[1]); fromField == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - 'copy_from'='%s' 'FieldName'='%s' was not found", field.entity.Name, field.Name, name, fromFieldNameSplit[1]))
			} else {
				return field.copyFromField(fromField)
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if field != nil && field.entity != nil && name != \"\" {}", []interface{}{field, name}).PrintfError()
}

func (field *Field) copyFromField(from *Field) (err error) {
	if field != nil && field.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - 'copy_from.FieldName'='%s' has not incorrect status='%s'", field.entity.Name, field.Name, from.Name, from.Status))
		}

		if field.Status == "" {
			field.Status = from.Status
		}

		//field.Required = from.Required
		//field.KeyType = from.KeyType
		//field.Modify = from.Modify

		if field.InternalType == "" {
			field.InternalType = from.InternalType
		}

		if field.Format == "" {
			field.Format = from.Format
		}

		if field.ValidateRule == "" {
			field.ValidateRule = from.ValidateRule
		}

		field.Alias.copyFrom(from.Alias, false)

		field.DbStorage.copyFrom(from.DbStorage, false)

		field.Tag.copyFrom(from.Tag, false)

		field.Exprs.copyFrom(from.Exprs, false)

		// Очистить все внутренние поля и сбросить признак инициации
		field.clearInternal()
		field.CopyFromName = from.entity.Name + "." + from.Name
		field.InheritFromName = ""

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if field != nil && field.entity != nil && from != nil {}", []interface{}{field, from}).PrintfError()
}

func (field *Field) inheritFromField(from *Field) (err error) {
	if field != nil && field.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' - 'inherit_from.FieldName'='%s' has not incorrect status='%s'", field.entity.Name, field.Name, from.Name, from.Status))
		}

		field.Status = from.Status
		field.Required = from.Required
		//field.KeyType = from.KeyType
		field.Modify = from.Modify
		field.InternalType = from.InternalType
		field.Format = from.Format
		field.ValidateRule = from.ValidateRule

		field.Alias.copyFrom(from.Alias, true)

		field.DbStorage.copyFrom(from.DbStorage, true)

		field.Tag.copyFrom(from.Tag, true)

		field.Exprs.copyFrom(from.Exprs, true)

		// Очистить все внутренние поля и сбросить признак инициации
		field.clearInternal()
		field.CopyFromName = ""
		field.InheritFromName = from.entity.Name + "." + from.Name

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if field != nil && field.entity != nil && from != nil {}", []interface{}{field, from}).PrintfError()
}

func (field *Field) initExprs() (err error) {
	if field.entity != nil {

		for _, expr := range field.Exprs {
			if expr != nil {
				if err = expr.init(field.entity, field); err != nil {
					return err
				}
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("field '%s' - empty Entity", field.Name))
}

func (field *Field) Entity() *Entity {
	if field != nil {
		return field.entity
	}
	return nil
}

func (field *Field) ReflectType() reflect.Type {
	if field != nil {
		return field.reflectType
	}
	return nil
}

func (field *Field) Index(t reflect.Type) []int {
	if field != nil && t != nil {
		if index, ok := field.indexMap[t]; ok {
			return index
		}
	}
	return nil
}

func (field *Field) References() References {
	if field != nil {
		return field.references
	}
	return nil
}

func (field *Field) Reference() *Reference {
	if field != nil {
		return field.reference
	}
	return nil
}

func (field *Field) CheckFieldReference() error {
	if field != nil {
		if field.Reference() == nil {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s' - empty 'Reference' pointer", field.entity.Name, field.Name)).PrintfError()
		}

		if field.Reference().ToEntity() == nil {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s', Reference '%s' - empty 'toEntity' pointer", field.entity.Name, field.Name, field.Reference().Name)).PrintfError()
		}

		if field.Reference().ToKey() == nil {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s', Field '%s', Reference '%s' - empty 'toKey' pointer", field.entity.Name, field.Name, field.Reference().Name)).PrintfError()
		}
	}
	return nil
}

func (field *Field) GetTagName(format string, useNameAsDefault bool) string {
	if field != nil {
		tag := field.Tag.GetName(format)
		if tag == "" && useNameAsDefault {
			tag = field.Name
		}
		return tag
	}
	return ""
}

func (field *Field) GetTag(format string, useNameAsDefault bool) string {
	if field != nil {
		tag := field.Tag.GetTag(format)
		if tag == "" && useNameAsDefault {
			tag = field.Name
		}
		return tag
	}
	return ""
}

func (field *Field) GetXmlNameFromTag(useNameAsDefault bool) *xml.Name {
	if field != nil {

		xmlTag := field.GetTag("xml", useNameAsDefault)
		if strings.ContainsAny(xmlTag, ">") {
			// Если указана вложенность элементов, то извлечь последний до >
			xmlTagIndex := strings.LastIndexByte(xmlTag, '>')
			if xmlTagIndex != -1 {
				xmlTag = xmlTag[xmlTagIndex+1:]
			}
		}
		return &xml.Name{
			Local: xmlTag,
			Space: field.GetTag("xmlSpace", false),
		}
	}
	return nil
}

var fieldNotFound = Field{
	Name: FIELD_NOT_FOUND,
	Alias: Alias{
		DisplayName: FIELD_NOT_FOUND,
	},
	Tag: Tag{
		Db:   FIELD_NOT_FOUND,
		Json: FIELD_NOT_FOUND,
		Xml:  FIELD_NOT_FOUND,
		Yaml: FIELD_NOT_FOUND,
		Xls:  FIELD_NOT_FOUND,
	},
}

type Fields []*Field

func (fields Fields) getUnsafe(fieldName string) *Field {
	for _, field := range fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

func (fields Fields) Get(fieldName string) *Field {
	for _, field := range fields {
		if field.Name == fieldName {
			return field
		}
	}
	return &fieldNotFound
}

func (fields Fields) GetDisplayName(fieldName string) string {
	field := fields.Get(fieldName)
	return field.Alias.DisplayName
}

func (fields Fields) GetFullName(fieldName string) string {
	field := fields.Get(fieldName)
	return field.Alias.FullName
}

func (fields Fields) GetValidationRules() map[string]string {
	var rules = make(map[string]string, len(fields))
	for _, field := range fields {
		if field.ValidateRule != "" {
			rules[field.Name] = field.ValidateRule
		} else {
			rules[field.Name] = "-"
		}
	}
	return rules
}

type FieldsMap map[string]*Field

func (fieldsMap FieldsMap) Get(name string) *Field {
	if fieldsMap != nil {
		if v, ok := fieldsMap[name]; ok {
			return v
		}
	}
	return nil
}

func (fieldsMap FieldsMap) String() string {
	if fieldsMap != nil {
		result := make([]string, len(fieldsMap))

		for fieldName, field := range fieldsMap {
			if field != nil {
				result = append(result, fieldName)
			}
		}

		return strings.Join(result, "-")
	}
	return ""
}

func (fieldsMap FieldsMap) Copy() FieldsMap {
	if fieldsMap != nil {

		result := make(FieldsMap, len(fieldsMap))

		for fieldName, field := range fieldsMap {
			if field != nil {
				result[fieldName] = field
			}
		}

		return result
	}
	return nil
}

func (fieldsMap FieldsMap) Merge(fieldsMap2 FieldsMap) FieldsMap {
	if fieldsMap != nil {

		result := make(FieldsMap, len(fieldsMap)+len(fieldsMap2))

		for fieldName, field := range fieldsMap {
			if field != nil {
				result[fieldName] = field
			}
		}

		for fieldName, field := range fieldsMap2 {
			if field != nil {
				result[fieldName] = field
			}
		}

		return result
	}
	return nil
}

func (fieldsMap FieldsMap) GetDisplayName(fieldName string) string {
	if field := fieldsMap.Get(fieldName); field != nil {
		return field.Alias.DisplayName
	} else {
		return fieldNotFound.Alias.DisplayName
	}
}

func (fieldsMap FieldsMap) GetFullName(fieldName string) string {
	if field := fieldsMap.Get(fieldName); field != nil {
		return field.Alias.FullName
	} else {
		return fieldNotFound.Alias.FullName
	}
}

func (fieldsMap FieldsMap) GetValidationRules() map[string]string {
	var rules = make(map[string]string, len(fieldsMap))
	for _, field := range fieldsMap {
		if field.ValidateRule != "" {
			rules[field.Name] = field.ValidateRule
		} else {
			rules[field.Name] = "-"
		}
	}
	return rules
}

// isLetter reports whether a given 'rune' is classified as a Letter.
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isValidFieldName(fieldName string) bool {
	for i, c := range fieldName {
		if i == 0 && !isLetter(c) {
			return false
		}

		if !(isLetter(c) || unicode.IsDigit(c)) {
			return false
		}
	}

	return len(fieldName) > 0
}
