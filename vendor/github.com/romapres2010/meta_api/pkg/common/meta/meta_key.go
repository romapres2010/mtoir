package meta

import (
	"fmt"
	"strings"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

type KeyType string

const (
	KEY_TYPE_PK KeyType = "PK"
	KEY_TYPE_UK KeyType = "UK"
	KEY_TYPE_FK KeyType = "FK"
)

// Keys - ключи сущности
type Keys []*Key

func (keys Keys) getUnsafe(name string) *Key {
	for _, key := range keys {
		if key.Name == name {
			return key
		}
	}
	return nil
}

// KeysMap - ключи сущности
type KeysMap map[string]*Key

func (keysMap KeysMap) Get(name string) *Key {
	if keysMap != nil {
		if v, ok := keysMap[name]; ok {
			return v
		}
	}
	return nil
}

// Key - ключ сущности
type Key struct {
	Status          string    `yaml:"status" json:"status" xml:"status"`                                                 // Статус поля ENABLED, DEPRECATED, ...
	Name            string    `yaml:"name" json:"name" xml:"name"`                                                       // Имя ключа
	InheritFromName string    `yaml:"inherit_from,omitempty" json:"inherit_from,omitempty" xml:"inherit_from,omitempty"` // Наследовать характеристики
	CopyFromName    string    `yaml:"copy_from,omitempty" json:"copy_from,omitempty" xml:"copy_from,omitempty"`          // Скопировать характеристики с другой сущности
	Type            KeyType   `yaml:"type" json:"type" xml:"type"`                                                       // Тип ключа
	FieldsName      []string  `yaml:"fields" json:"fields" xml:"fields"`                                                 // Поля ключа
	Alias           Alias     `yaml:"alias,omitempty" json:"alias,omitempty" xml:"alias,omitempty"`                      // Дополнительные имена, коды и описание
	DbStorage       DbStorage `yaml:"db_storage,omitempty" json:"db_storage,omitempty" xml:"db_storage,omitempty"`       // Параметры хранения в БД
	Modify          Modify    `yaml:"modify,omitempty" json:"modify,omitempty" xml:"modify,omitempty"`                   // Разрешенные операции с полями ключа

	entity       *Entity   // Сущность к которой относится ключ
	copyFrom     *Key      // Копировать характеристики с возможностью переопределения
	inheritFrom  *Key      // Наследовать характеристики
	fields       Fields    // Поля ключа, после разбора
	fieldsMap    FieldsMap // Поля ключа, после разбора для быстрого поиска
	fieldsString string    // Форматированный список полей ключа
	isInit       bool      // признак, что инициация успешная
}

func (entity *Entity) newKey(name string) *Key {
	if entity != nil {

		key := &Key{
			Status: STATUS_ENABLED,
			Name:   name,
			entity: entity,
		}

		return key
	}

	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
	return nil
}

func (key *Key) clearInternal() {
	if key != nil {
		//key.entity = nil
		key.copyFrom = nil
		key.inheritFrom = nil
		key.fields = nil
		key.fieldsMap = nil
		key.fieldsString = ""
		key.isInit = false
	}
}

func (key *Key) Entity() *Entity {
	if key != nil {
		return key.entity
	}
	return nil
}

func (key *Key) Fields() Fields {
	if key != nil {
		return key.fields
	}
	return nil
}

func (key *Key) FieldsString() string {
	if key != nil {
		return key.fieldsString
	}
	return ""
}

func (key *Key) AddFieldsToMap(fieldsMap *FieldsMap) {
	if key != nil && fieldsMap != nil {
		for _, field := range key.fields {
			(*fieldsMap)[field.Name] = field
		}
	}
}

func (key *Key) FieldsMap() FieldsMap {
	if key != nil && key.fieldsMap != nil {
		return key.fieldsMap.Copy()
	}
	return nil
}

func (key *Key) setEntity(entity *Entity) {
	if key != nil && entity != nil {
		if !key.isInit {
			key.entity = entity
		}
	}
}

func (key *Key) init(entity *Entity) error {
	if entity != nil {

		if key.isInit {
			return nil
		}

		key.entity = entity // TODO - определить место присвоения

		if key.Name == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - empty 'name'", key.entity.Name, key.Name))
		}

		if key.Status == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - empty 'status'", key.entity.Name, key.Name))
		}

		if key.CopyFromName != "" {
			if err := key.copyFromKeyName(key.CopyFromName); err != nil {
				return err
			}
		}

		if key.Type != KEY_TYPE_PK && key.Type != KEY_TYPE_UK && key.Type != KEY_TYPE_FK {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - incorrect type '%s'", key.entity.Name, key.Name, key.Type))
		}

		if len(key.FieldsName) == 0 {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - empty fields list", key.entity.Name, key.Name))
		}

		key.fieldsMap = make(FieldsMap, len(key.FieldsName))
		key.fields = make(Fields, 0, len(key.FieldsName))

		for _, fieldName := range key.FieldsName {
			if field := key.entity.fieldByNameUnsafe(fieldName); field != nil {

				if field.Status != STATUS_ENABLED {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - field '%s' must be enabled", key.entity.Name, key.Name, fieldName))
				}

				if _, ok := key.fieldsMap[field.Name]; ok {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s'  Key '%s'  - duplicate field '%s'", key.entity.Name, key.Name, fieldName))
				} else {
					key.fieldsMap[field.Name] = field
				}

				key.fields = append(key.fields, field)
				key.fieldsString = key.fieldsString + ", " + fieldName + "(" + field.InternalType + ")"

			} else {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - field '%s' not found", key.entity.Name, key.Name, fieldName))
			}
			key.fieldsString = strings.TrimLeft(key.fieldsString, ", ")
		}

		key.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Key '%s' - empty Entity", key.Name))
}

func (key *Key) copyFromKeyName(name string) (err error) {
	if key != nil && key.entity != nil && name != "" {

		if name != "" {
			fromFieldNameSplit := strings.Split(name, ".")

			// Формат должен быть 'EntityName'.'KeyName',
			if len(fromFieldNameSplit) != 2 {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - 'copy_from'=['%s'] must be 'EntityName.KeyName'", key.entity.Name, key.Name, name))
			}

			// Найти сущность
			fromEntity := key.entity.meta.GetEntityUnsafe(fromFieldNameSplit[0])
			if fromEntity == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - 'copy_from'='%s' 'EntityName'='%s' was not found", key.entity.Name, key.Name, name, fromFieldNameSplit[0]))
			}

			// Найти ключ
			fromKey := fromEntity.KeyByName(fromFieldNameSplit[1])
			if fromKey == nil {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - 'copy_from'='%s' 'KeyName'='%s' was not found", key.entity.Name, key.Name, name, fromFieldNameSplit[1]))
			} else {
				return key.copyFromKey(fromKey)
			}

		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if key != nil && key.entity != nil && name != \"\" {}", []interface{}{key, name}).PrintfError()
}

func (key *Key) copyFromKey(from *Key) (err error) {
	if key != nil && key.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - 'copy_from.KeyName'='%s' has not incorrect status='%s'", key.entity.Name, key.Name, from.Name, from.Status))
		}

		if key.Status == "" {
			key.Status = from.Status
		}

		if key.Type == "" {
			key.Type = from.Type
		}

		if len(key.FieldsName) == 0 {
			key.FieldsName = from.FieldsName
		}

		key.Alias.copyFrom(from.Alias, false)

		key.DbStorage.copyFrom(from.DbStorage, false)

		// Очистить все внутренние поля и сбросить признак инициации
		key.clearInternal()
		key.CopyFromName = from.entity.Name + "." + from.Name
		key.InheritFromName = ""

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if key != nil && key.entity != nil && from != nil {}", []interface{}{key, from}).PrintfError()
}

func (key *Key) inheritFromKey(from *Key) (err error) {
	if key != nil && key.entity != nil && from != nil {

		if from.Status != STATUS_ENABLED {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Key '%s' - 'inherit_from.KeyName'='%s' has not incorrect status='%s'", key.entity.Name, key.Name, from.Name, from.Status))
		}

		key.Status = from.Status
		key.Type = from.Type
		key.FieldsName = from.FieldsName
		key.Modify = from.Modify
		key.Alias.copyFrom(from.Alias, true)
		key.DbStorage.copyFrom(from.DbStorage, true)

		// Очистить все внутренние поля и сбросить признак инициации
		key.clearInternal()
		key.CopyFromName = ""
		key.InheritFromName = from.entity.Name + "." + from.Name

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if key != nil && key.entity != nil && from != nil {}", []interface{}{key, from}).PrintfError()
}
