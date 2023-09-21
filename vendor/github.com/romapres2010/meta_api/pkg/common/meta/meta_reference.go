package meta

import (
	"fmt"
	"strings"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

type ReferenceType string

const (
	REFERENCE_TYPE_ASSOCIATION ReferenceType = "Association" // Отношение от подчиненной записи на родителя
	REFERENCE_TYPE_COMPOSITION ReferenceType = "Composition" // Отношение от родителя на подчиненные записи
)

type ReferenceCardinality string

const (
	REFERENCE_CARDINALITY_EMPTY ReferenceCardinality = ""  // Отношение 0
	REFERENCE_CARDINALITY_0     ReferenceCardinality = "0" // Отношение 0
	REFERENCE_CARDINALITY_1     ReferenceCardinality = "1" // Отношение 1
	REFERENCE_CARDINALITY_M     ReferenceCardinality = "M" // Отношение M
)

// References - ссылки между сущностями
type References []*Reference

func (refs References) getUnsafe(name string) *Reference {
	for _, ref := range refs {
		if ref.Name == name {
			return ref
		}
	}
	return nil
}

// ReferencesMap - ссылки между сущностями
type ReferencesMap map[string]*Reference

func (referencesMap ReferencesMap) Get(name string) *Reference {
	if referencesMap != nil {
		if v, ok := referencesMap[name]; ok {
			return v
		}
	}
	return nil
}

// Reference - ссылка между сущностями
type Reference struct {
	Status          string               `yaml:"status" json:"status" xml:"status"`                                                 // Статус поля ENABLED, DEPRECATED, ...
	Name            string               `yaml:"name" json:"name" xml:""`                                                           // Имя ключа
	InheritFromName string               `yaml:"inherit_from,omitempty" json:"inherit_from,omitempty" xml:"inherit_from,omitempty"` // Наследовать характеристики
	CopyFromName    string               `yaml:"copy_from,omitempty" json:"copy_from,omitempty" xml:"copy_from,omitempty"`          // Скопировать характеристики с другой сущности
	Required        bool                 `yaml:"required" json:"required" xml:"required"`                                           // Признак, что ссылка обязательная в сущности
	ToEntityName    string               `yaml:"to_entity" json:"to_entity" xml:"to_entity"`                                        // Имя сущности, на которую ссылаемся
	ToKeyName       string               `yaml:"to_key" json:"to_key" xml:"to_key"`                                                 // Имя ключа, на который ссылаемся
	ToReferenceName string               `yaml:"to_reference" json:"to_reference" xml:"to_reference"`                               // Имя reference - зеркального данному Composition-Association
	Type            ReferenceType        `yaml:"type" json:"type" xml:"type"`                                                       // Тип ссылки
	Cardinality     ReferenceCardinality `yaml:"cardinality" json:"cardinality" xml:"cardinality"`                                  // cardinality ссылки
	Embed           bool                 `yaml:"embed" json:"embed" xml:"embed"`                                                    // Признак, что подчиненный объект встраивается по ссылке
	FieldsName      []string             `yaml:"fields" json:"fields" xml:"fields"`                                                 // Поля сущности, которая ссылается, должны совпадать по списку и типу
	ValidateRule    string               `yaml:"validate_rule" json:"validate_rule" xml:"validate_rule"`                            // Правила валидации поля
	Alias           Alias                `yaml:"alias,omitempty" json:"alias,omitempty" xml:"alias,omitempty"`                      // Дополнительные имена, коды и описание
	DbStorage       DbStorage            `yaml:"db_storage,omitempty" json:"db_storage,omitempty" xml:"db_storage,omitempty"`       // Параметры хранения в БД
	Tag             Tag                  `yaml:"tag,omitempty" json:"tag,omitempty" xml:"tag,omitempty"`                            // Имена различных тегов для парсинга данных
	Exprs           Exprs                `yaml:"expressions,omitempty" json:"expressions,omitempty" xml:"expressions,omitempty"`    // Выражение для вычисления или проверки

	entity       *Entity    // Сущность, которая ссылается
	copyFrom     *Reference // Скопировать характеристики с возможностью переопределения
	inheritFrom  *Reference // Наследовать характеристики
	toEntity     *Entity    // Сущность, на которую ссылаемся
	toKey        *Key       // Ключ, на который ссылаемся
	toReference  *Reference // reference - зеркального данному Composition-Association
	field        *Field     // Виртуальное поле, создаваемое для каждой reference
	fields       Fields     // Поля ключа, после разбора
	fieldsMap    FieldsMap  // Поля ключа, после разбора для быстрого поиска
	fieldsString string     // Форматированный список полей ключа для вывода в сообщениях
	isInit       bool       // признак, что инициация успешная
}

func (entity *Entity) newReference(name string) *Reference {
	if entity != nil {

		ref := &Reference{
			Status: STATUS_ENABLED,
			Name:   name,
			entity: entity,
		}

		return ref
	}

	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
	return nil
}

func (ref *Reference) clearInternal() {
	if ref != nil {
		//ref.field = nil TODO - не работает - проблема с автоматически создаваемыми полями для Reference
		//ref.entity = nil
		ref.copyFrom = nil
		ref.inheritFrom = nil
		ref.toEntity = nil
		ref.toKey = nil
		ref.toReference = nil
		ref.fields = nil
		ref.fieldsMap = nil
		ref.fieldsString = ""
		ref.isInit = false
	}
}

func (ref *Reference) ToEntity() *Entity {
	if ref != nil {
		return ref.toEntity
	}
	return nil
}

func (ref *Reference) ToKey() *Key {
	if ref != nil {
		return ref.toKey
	}
	return nil
}

func (ref *Reference) Field() *Field {
	if ref != nil {
		return ref.field
	}
	return nil
}

func (ref *Reference) ToReference() *Reference {
	if ref != nil {
		return ref.toReference
	}
	return nil
}

func (ref *Reference) FieldsString() string {
	if ref != nil {
		return ref.fieldsString
	}
	return ""
}

func (ref *Reference) Entity() *Entity {
	if ref != nil {
		return ref.entity
	}
	return nil
}

func (ref *Reference) AddFieldsToMap(fieldsMap *FieldsMap) {
	if ref != nil && fieldsMap != nil {
		(*fieldsMap)[ref.field.Name] = ref.field
		for _, field := range ref.fields {
			(*fieldsMap)[field.Name] = field
		}
	}
}

func (ref *Reference) FieldsMap() FieldsMap {
	if ref != nil && ref.fieldsMap != nil {
		return ref.fieldsMap.Copy()
	}
	return nil
}

func (ref *Reference) GetTag(format string, useNameAsDefault bool) string {
	if ref != nil {

		tag := ref.Tag.GetTag(format)
		if tag == "" && useNameAsDefault {
			tag = ref.Name
		}

		return tag
	}
	return ""
}

func (ref *Reference) GetTagName(format string, useNameAsDefault bool) string {
	if ref != nil {

		tag := ref.Tag.GetName(format)
		if tag == "" && useNameAsDefault {
			tag = ref.Name
		}

		return tag
	}
	return ""
}

func (ref *Reference) setEntity(entity *Entity) {
	if ref != nil && entity != nil {
		if !ref.isInit {
			ref.entity = entity
		}
	}
}

// func (ref *Reference) init(entity *Entity) error {
func (ref *Reference) init() error {
	if ref != nil && ref.entity != nil {

		if ref.isInit {
			return nil
		}

		//ref.entity = entity // TODO - определить место присвоения

		if ref.Name == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - empty 'name'", ref.entity.Name, ref.Name))
		}

		if ref.Status == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - empty 'status'", ref.entity.Name, ref.Name))
		}

		if ref.CopyFromName != "" {
			if err := ref.copyFromReferenceName(ref.CopyFromName); err != nil {
				return err
			}
		}

		if ref.Type != REFERENCE_TYPE_ASSOCIATION && ref.Type != REFERENCE_TYPE_COMPOSITION {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - incorrect 'type' '%s'", ref.entity.Name, ref.Name, ref.Type))
		}

		if ref.Cardinality == REFERENCE_CARDINALITY_EMPTY {
			// по умолчанию используем 1:M
			ref.Cardinality = REFERENCE_CARDINALITY_M
		}

		if ref.Cardinality != REFERENCE_CARDINALITY_0 && ref.Cardinality != REFERENCE_CARDINALITY_1 && ref.Cardinality != REFERENCE_CARDINALITY_M {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - incorrect 'cardinality' '%s'", ref.entity.Name, ref.Name, ref.Cardinality))
		}

		if ref.Cardinality == REFERENCE_CARDINALITY_M && ref.Embed {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - incorrect 'cardinality' '%s' for embed", ref.entity.Name, ref.Name, ref.Cardinality))
		}

		if len(ref.FieldsName) == 0 {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - empty fields list", ref.entity.Name, ref.Name))
		}

		ref.Tag.init()
		ref.fieldsMap = make(FieldsMap, len(ref.FieldsName))
		ref.fields = make(Fields, 0, len(ref.FieldsName))

		for _, fieldName := range ref.FieldsName {
			if field := ref.entity.fieldByNameUnsafe(fieldName); field != nil {

				if field.Status != STATUS_ENABLED {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - field '%s' must be enabled", ref.entity.Name, ref.Name, fieldName))
				}

				if ref.Required && !field.Required {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' is reqired - field '%s' must be also reqired", ref.entity.Name, ref.Name, fieldName))
				}

				if _, ok := ref.fieldsMap[field.Name]; ok {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s'  Reference '%s'  - duplicate field '%s'", ref.entity.Name, ref.Name, fieldName))
				} else {
					ref.fieldsMap[field.Name] = field
					field.references = append(field.references, ref) // поле содержит список reference, в состав которых оно входит
				}

				ref.fields = append(ref.fields, field)
				ref.fieldsString = ref.fieldsString + ", " + fieldName + "(" + field.InternalType + ")"

			} else {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - field '%s' not found", ref.entity.Name, ref.Name, fieldName))
			}
			ref.fieldsString = strings.TrimLeft(ref.fieldsString, ", ")
		}

		{ // Проверим наличие связанной сущности правильность типов полей

			if ref.entity.meta == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - empty meta pointer", ref.entity.Name, ref.Name, ref.fieldsString))
			}

			if ref.ToEntityName == "" {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - empty 'to_entity_name'", ref.entity.Name, ref.Name, ref.fieldsString))
			}

			if ref.toEntity = ref.entity.meta.GetEntityUnsafe(ref.ToEntityName); ref.toEntity == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' does not exist", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName))
			}

			if ref.toKey = ref.toEntity.KeyByName(ref.ToKeyName); ref.toKey == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' Key '%s' does not exist", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName, ref.ToKeyName))
			}

			if len(ref.fields) != len(ref.toKey.fields) {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' Key '%s'['%s'] has non equal field quantity", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName, ref.ToKeyName, ref.toKey.fieldsString))
			}

			// сравним типы полей
			for i, _ := range ref.fields {
				if !ref.fields[i].reflectType.AssignableTo(ref.toKey.fields[i].reflectType) {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' Key '%s'['%s'] has non equal field type in position '%d'", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName, ref.ToKeyName, ref.toKey.fieldsString, i))
				}
			}

			// Найти связанный reference
			if ref.ToReferenceName != "" {
				if ref.toReference = ref.toEntity.referenceDefByNameUnsafe(ref.ToReferenceName); ref.toReference == nil {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' ToReference '%s' does not exist", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName, ref.ToReferenceName))
				}

				if ref.toReference.Status != STATUS_ENABLED {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s'['%s'] - reference Entity '%s' ToReference '%s' has not incorrect status='%s'", ref.entity.Name, ref.Name, ref.fieldsString, ref.ToEntityName, ref.ToReferenceName, ref.toReference.Status))
				}
			}

			// Добавим в список для обратной связи сущностей
			ref.toEntity.addReferenceBy(ref)

		} // Проверим наличие связанной сущности правильность типов полей

		// Инициируем выражения
		if err := ref.initExpr(); err != nil {
			return err
		}

		ref.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Reference '%s' - empty Entity", ref.Name))
}

func (ref *Reference) initExpr() (err error) {
	if ref != nil && ref.entity != nil && ref.field != nil {

		for _, expr := range ref.Exprs {
			if expr != nil {
				if err = expr.init(ref.entity, ref.field); err != nil {
					return err
				}
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if ref != nil && ref.entity != nil && ref.field != nil {}", []interface{}{ref}).PrintfError()
}

func (ref *Reference) copyFromReferenceName(name string) (err error) {
	if ref != nil && ref.entity != nil && name != "" {

		if name != "" {
			fromReferenceNameSplit := strings.Split(name, ".")

			// Формат должен быть 'EntityName'.'ReferenceName',
			if len(fromReferenceNameSplit) != 2 {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - 'copy_from'=['%s'] must be 'EntityName.ReferenceName'", ref.entity.Name, ref.Name, name))
			}

			// Найти сущность
			fromEntity := ref.entity.meta.GetEntityUnsafe(fromReferenceNameSplit[0])
			if fromEntity == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - 'copy_from'='%s' 'EntityName'='%s' was not found", ref.entity.Name, ref.Name, name, fromReferenceNameSplit[0]))
			}

			// Найти reference
			fromRef := fromEntity.ReferenceByName(fromReferenceNameSplit[1])
			if fromRef == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - 'copy_from'='%s' 'ReferenceName'='%s' was not found", ref.entity.Name, ref.Name, name, fromReferenceNameSplit[1]))
			} else {
				return ref.copyFromReference(fromRef)
			}

		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if ref != nil && ref.entity != nil && name != \"\" {}", []interface{}{ref, name}).PrintfError()
}

func (ref *Reference) copyFromReference(from *Reference) (err error) {
	if ref != nil && ref.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - 'copy_from'.ReferenceName'='%s' has not incorrect status='%s'", ref.entity.Name, ref.Name, from.Name, from.Status))
		}

		if ref.Status == "" {
			ref.Status = from.Status
		}

		if ref.Type == "" {
			ref.Type = from.Type
		}

		if len(ref.FieldsName) == 0 {
			ref.FieldsName = from.FieldsName
		}

		if ref.ValidateRule == "" {
			ref.ValidateRule = from.ValidateRule
		}

		// Для рекурсивных отношений подменим имя сущности на свое
		if ref.ToEntityName == "" {
			if from.ToEntityName == from.entity.Name {
				ref.ToEntityName = ref.entity.Name
			} else {
				ref.ToEntityName = from.ToEntityName
			}
		}

		if ref.ToKeyName == "" {
			ref.ToKeyName = from.ToKeyName
		}

		if ref.ToReferenceName == "" {
			ref.ToReferenceName = from.ToReferenceName
		}

		ref.Alias.copyFrom(from.Alias, false)

		ref.DbStorage.copyFrom(from.DbStorage, false)

		ref.Tag.copyFrom(from.Tag, false)

		// Очистить все внутренние поля и сбросить признак инициации
		ref.clearInternal()
		ref.CopyFromName = from.entity.Name + "." + from.Name
		ref.InheritFromName = ""

		//ref.copyFrom = from
		//ref.CopyFromName = from.entity.Name + "." + from.Name
		//
		////ref.entity = nil
		//ref.InheritFromName = ""
		//ref.inheritFrom = nil
		//ref.toEntity = nil
		//ref.toKey = nil
		//ref.toReference = nil
		////ref.field = nil TODO - не работает?
		//ref.fields = nil
		//ref.argsFieldsMap = nil
		//ref.argsFieldsString = ""
		//ref.isInit = false

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if ref != nil && ref.entity != nil && from != nil {}", []interface{}{ref, from}).PrintfError()
}

func (ref *Reference) inheritFromReference(from *Reference) (err error) {
	if ref != nil && ref.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - 'inherit_from'.ReferenceName'='%s' has not incorrect status='%s'", ref.entity.Name, ref.Name, from.Name, from.Status))
		}

		ref.Status = from.Status
		ref.Type = from.Type
		ref.Required = from.Required
		ref.FieldsName = from.FieldsName
		ref.ValidateRule = from.ValidateRule

		// Для рекурсивных отношений подменим имя сущности на свое
		if from.ToEntityName == from.entity.Name {
			ref.ToEntityName = ref.entity.Name
		} else {
			ref.ToEntityName = from.ToEntityName
		}
		ref.ToKeyName = from.ToKeyName
		ref.ToReferenceName = from.ToReferenceName

		ref.Alias.copyFrom(from.Alias, true)
		ref.DbStorage.copyFrom(from.DbStorage, true)
		ref.Tag.copyFrom(from.Tag, true)

		// Очистить все внутренние поля и сбросить признак инициации
		ref.clearInternal()
		ref.CopyFromName = ""
		ref.InheritFromName = from.entity.Name + "." + from.Name

		//ref.inheritFrom = from
		//ref.InheritFromName = from.entity.Name + "." + from.Name
		//
		////ref.entity = nil
		//ref.CopyFromName = ""
		//ref.copyFrom = nil
		//ref.toEntity = nil
		//ref.toKey = nil
		//ref.toReference = nil
		////ref.field = nil TODO - не работает?
		//ref.fields = nil
		//ref.argsFieldsMap = nil
		//ref.argsFieldsString = ""
		//ref.isInit = false // требуется инициация

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if ref != nil && ref.entity != nil && from != nil {}", []interface{}{ref, from}).PrintfError()
}
