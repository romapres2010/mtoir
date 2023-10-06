package meta

import (
	"fmt"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

func (entity *Entity) init(meta *Meta) (err error) {
	if entity != nil && meta != nil {
		entity.mx.Lock()
		defer entity.mx.Unlock()

		if entity.isInit {
			return nil
		}

		entity.meta = meta

		if entity.Status == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty 'status'", entity.Name))
		}

		if entity.InheritFromName != "" {
			if err = entity.inheritFromName(entity.InheritFromName); err != nil {
				return err
			}
		}

		countFields := len(entity.Fields) + len(entity.ReferencesDef) + 1
		entity.fieldsMap = make(FieldsMap, countFields)
		entity.structFields = make(Fields, 0, countFields)
		entity.structFieldsMap = make(FieldsMap, countFields)
		entity.associationMap = make(FieldsMap, len(entity.ReferencesDef))
		entity.compositionMap = make(FieldsMap, len(entity.ReferencesDef))
		entity.dbNameMap = make(FieldsMap, countFields)
		entity.jsonNameMap = make(FieldsMap, countFields)
		entity.xmlNameMap = make(FieldsMap, countFields)
		entity.yamlNameMap = make(FieldsMap, countFields)
		entity.xlsNameMap = make(FieldsMap, countFields)
		entity.exprNameMap = make(FieldsMap, countFields)
		entity.keyFieldsMap = make(FieldsMap)

		// Разберем tag и заполним имена
		entity.Tag.init()

		if err = entity.initReferencesSetEntity(); err != nil {
			return err
		}

		// Для отношений добавим виртуальные поля, в которые будем парсить данные
		if err = entity.initReferencesAddFields(); err != nil {
			return err
		}

		// Системные поля
		entity.initXMLNameField()
		entity.initErrorsField()
		entity.initValidField()
		//entity.initCacheValidField()
		//entity.initMxField()

		if err = entity.initFields(); err != nil {
			return err
		}

		entity.validationRules = entity.Fields.GetValidationRules()

		if err = entity.initKeys(); err != nil {
			return err
		}

		entity.initKeyFieldsMap()

		if err = entity.initFieldsTagMap(); err != nil {
			return err
		}

		if err = entity.initExpr(); err != nil {
			return err
		}

		if err = entity.initFieldsExprs(); err != nil {
			return err
		}

		if err = entity.initExprs(); err != nil {
			return err
		}

		entity.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && meta != nil {}", []interface{}{entity, meta}).PrintfError()
}

func (entity *Entity) initExpr() (err error) {
	if entity != nil {

		for _, expr := range entity.Exprs {
			if expr != nil {
				if err = expr.init(entity, nil); err != nil {
					return err
				}
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("empty Entity"))
}

func (entity *Entity) initXMLNameField() {
	// Для XML нужно дополнительно задать XMLName https://pkg.go.dev/encoding/xml#Marshal
	if entity.Fields.getUnsafe("XMLName") == nil {
		// Если такого поля еще нет, оно может быть задано явно в meta
		field := &Field{
			Status:       STATUS_ENABLED,
			Order:        len(entity.Fields) + 1,
			Name:         "XMLName",
			InternalType: FIELD_TYPE_XML_NAME,
			Format:       "",
			ValidateRule: "-",
			System:       true,
			Alias:        Alias{},
			Tag: Tag{
				Db:   "-",
				Json: "-",
				Yaml: "-",
				Xls:  "-",
				//Xml:  entity.GetTag("xml", true), // Задает имя элемента в XML, имеет приоритет над значением поля XMLName
			},
		}
		entity.addField(field)
		entity.xmlNameField = field
	}
}

func (entity *Entity) initErrorsField() {
	if entity.Fields.getUnsafe("INTERNAL_ERROR") == nil {
		// Если такого поля еще нет, оно может быть задано явно в meta
		field := &Field{
			Status:       STATUS_ENABLED,
			Order:        len(entity.Fields) + 1,
			Name:         "INTERNAL_ERROR",
			InternalType: FIELD_TYPE_INTERNAL_ERROR,
			Format:       "",
			ValidateRule: "-",
			System:       true,
			Alias:        Alias{},
			Tag: Tag{
				Db:   "-",
				Json: "__ERRORS__,omitempty",
				Yaml: "__ERRORS__,omitempty",
				Xls:  "__ERRORS__",
				Xml:  "__ERRORS__>__ERROR__,omitempty",
			},
		}
		entity.addField(field)
		entity.errorsField = field
	}
}

//func (entity *Entity) initExternalIdField() {
//	if entity.fields.getUnsafe("INTERNAL_EXTERNAL_ID") == nil {
//		// Если такого поля еще нет, оно может быть задано явно в meta
//		field := &field{
//			Status:       STATUS_ENABLED,
//			Order:        len(entity.fields) + 1,
//			Name:         "INTERNAL_EXTERNAL_ID",
//			InternalType: FIELD_TYPE_INTERNAL_EXTERNAL_ID,
//			Format:       "",
//			Reference:    nil,
//			ValidateRule: "-",
//			System:       true,
//			DoNotCopy:    false,
//			Alias:        Alias{},
//			Tag: Tag{
//				DbStorage:   "-",
//				Json: "-",
//				Yaml: "-",
//				Xls:  "-",
//				Xml:  "-",
//			},
//		}
//		entity.addField(field)
//		entity.cacheInvalidField = field
//	}
//}

//func (entity *Entity) initCacheValidField() {
//	if entity.fields.getUnsafe("INTERNAL_CACHE_INVALID") == nil {
//		// Если такого поля еще нет, оно может быть задано явно в meta
//		field := &field{
//			Status:       STATUS_ENABLED,
//			Order:        len(entity.fields) + 1,
//			Name:         "INTERNAL_CACHE_INVALID",
//			InternalType: FIELD_TYPE_CACHE_INVALID,
//			Format:       "",
//			Reference:    nil,
//			ValidateRule: "-",
//			System:       true,
//			Alias:        Alias{},
//			Tag: Tag{
//				Db:   "-",
//				Json: "-",
//				Yaml: "-",
//				Xls:  "-",
//				Xml:  "-",
//			},
//			Modify: Modify{
//				CopyRestrict: true,
//			},
//		}
//		entity.addField(field)
//		entity.cacheInvalidField = field
//	}
//}

func (entity *Entity) initValidField() {
	if entity.Fields.getUnsafe("INTERNAL_VALID") == nil {
		// Если такого поля еще нет, оно может быть задано явно в meta
		field := &Field{
			Status:       STATUS_ENABLED,
			Order:        len(entity.Fields) + 1,
			Name:         "INTERNAL_VALID",
			InternalType: FIELD_TYPE_VALIDATION_VALID,
			Format:       "",
			ValidateRule: "-",
			System:       true,
			Alias:        Alias{},
			Tag: Tag{
				Db:   "-",
				Json: "-",
				Yaml: "-",
				Xls:  "-",
				Xml:  "-",
			},
		}
		entity.addField(field)
		entity.validField = field
	}
}

//func (entity *Entity) initMxField() {
//	if entity.fields.getUnsafe("INTERNAL_MX") == nil {
//		// Если такого поля еще нет, оно может быть задано явно в meta
//		field := &field{
//			Status:       STATUS_ENABLED,
//			Order:        len(entity.fields) + 1,
//			Name:         "INTERNAL_MX",
//			InternalType: FIELD_TYPE_RWMUTEX,
//			Format:       "",
//			Reference:    nil,
//			ValidateRule: "-",
//			System:       true,
//			Alias:        Alias{},
//			Tag: Tag{
//				Db:   "-",
//				Json: "-",
//				Yaml: "-",
//				Xls:  "-",
//				Xml:  "-",
//			},
//			Modify: Modify{
//				CopyRestrict: true,
//			},
//		}
//		entity.addField(field)
//		entity.mxField = field
//	}
//}

func (entity *Entity) initFields() error {
	for _, field := range entity.Fields {
		if err := entity.initField(field); err != nil {
			return err
		}
	}

	return nil
}

func (entity *Entity) initFieldsExprs() error {
	for _, field := range entity.Fields {
		if err := field.initExprs(); err != nil {
			return err
		}
		entity.fieldsExprs = append(entity.fieldsExprs, field.Exprs...)
	}

	if len(entity.fieldsExprs) > 0 {
		entity.fieldsExprsByAction = make(ExprsByAction)

		for _, expr := range entity.fieldsExprs {
			exprsByAction, ok := entity.fieldsExprsByAction[expr.Action]
			if !ok {
				exprsByAction = &Exprs{}
				entity.fieldsExprsByAction[expr.Action] = exprsByAction
			}
			*exprsByAction = append(*exprsByAction, expr)
		}
	}

	return nil
}

func (entity *Entity) initExprs() error {

	entity.exprsByAction = make(ExprsByAction)
	for _, expr := range entity.Exprs {
		if expr != nil {
			if err := expr.init(entity, nil); err != nil {
				return err
			}
		}

		exprsByAction, ok := entity.exprsByAction[expr.Action]
		if !ok {
			exprsByAction = &Exprs{}
			entity.exprsByAction[expr.Action] = exprsByAction
		}
		*exprsByAction = append(*exprsByAction, expr)

	}

	return nil
}

func (entity *Entity) initField(field *Field) error {
	if _, ok := entity.fieldsMap[field.Name]; ok {
		return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - duplicate field '%s'", entity.Name, field.Name)).PrintfError()
	} else {

		if err := field.init(entity); err != nil {
			return err
		}

		entity.fieldsMap[field.Name] = field

		if field.reflectType != nil {
			entity.structFieldsMap[field.Name] = field
			entity.structFields = append(entity.structFields, field)
		}

		return nil
	}
}

func (entity *Entity) initFieldsTagMap() error {
	var errDetail []string

	for _, field := range entity.Fields {
		entity.initFieldTagMap(field, &errDetail)
	}

	if len(errDetail) > 0 {
		err := _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' found duplicate tag", entity.Name)).PrintfError()
		err.Details = errDetail
		return err
	}
	return nil
}

func (entity *Entity) initFieldTagMap(field *Field, errDetail *[]string) {
	if reflectType, err := field.getReflectType(); err != nil || reflectType == nil {
		return // не включать, обработку ошибки в этом месте не делать
	}

	if field.Tag.DbName != "" && field.Tag.DbName != "-" {
		if _, ok := entity.dbNameMap[field.Tag.DbName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.db'='%s'", entity.Name, field.Name, field.Tag.DbName))
		} else {
			entity.dbNameMap[field.Tag.DbName] = field
		}
	}

	if field.Tag.JsonName != "" && field.Tag.JsonName != "-" {
		if _, ok := entity.jsonNameMap[field.Tag.JsonName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.json'='%s'", entity.Name, field.Name, field.Tag.JsonName))
		} else {
			entity.jsonNameMap[field.Tag.JsonName] = field
		}
	}

	if field.Tag.XmlName != "" && field.Tag.XmlName != "-" {
		if _, ok := entity.xmlNameMap[field.Tag.XmlName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.xml'='%s'", entity.Name, field.Name, field.Tag.XmlName))
		} else {
			entity.xmlNameMap[field.Tag.XmlName] = field
		}
	}

	if field.Tag.YamlName != "" && field.Tag.YamlName != "-" {
		if _, ok := entity.yamlNameMap[field.Tag.YamlName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.yaml'='%s'", entity.Name, field.Name, field.Tag.YamlName))
		} else {
			entity.yamlNameMap[field.Tag.YamlName] = field
		}
	}

	if field.Tag.XlsName != "" && field.Tag.XlsName != "-" {
		if _, ok := entity.xlsNameMap[field.Tag.XlsName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.xls'='%s'", entity.Name, field.Name, field.Tag.XlsName))
		} else {
			entity.xlsNameMap[field.Tag.XlsName] = field
		}
	}

	if field.Tag.ExprName != "" && field.Tag.ExprName != "-" {
		if _, ok := entity.xlsNameMap[field.Tag.ExprName]; ok {
			*errDetail = append(*errDetail, fmt.Sprintf("Entity '%s' field '%s' found duplicate 'tag.expr'='%s'", entity.Name, field.Name, field.Tag.ExprName))
		} else {
			entity.xlsNameMap[field.Tag.ExprName] = field
		}
	}
}

func (entity *Entity) initKeys() error {
	entity.keysMap = make(KeysMap, len(entity.KeysDef))
	entity.keys = make(Keys, 0, len(entity.KeysDef))
	entity.keysUk = make(Keys, 0, len(entity.KeysDef))

	keysUk := make(Keys, 0, len(entity.KeysDef))

	for _, key := range entity.KeysDef {

		// Только включенные ключи
		if key.Status == STATUS_ENABLED {

			if _, ok := entity.keysMap[key.Name]; ok {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - duplicate Key '%s'", entity.Name, key.Name)).PrintfError()
			} else {

				if err := key.init(entity); err != nil {
					return err
				}

				entity.keysMap[key.Name] = key

				if key.Type == KEY_TYPE_PK {
					entity.pkKey = key
					entity.keys = append(entity.keys, key)   // Первым в списке идет PK
					entity.keysUk = append(entity.keys, key) // Первым в списке идет PK
				} else if key.Type == KEY_TYPE_UK {
					keysUk = append(keysUk, key) // UK собираем отдельно и добавим в список после PK
				} else if key.Type == KEY_TYPE_FK {
					entity.keys = append(entity.keys, key)
				} else {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - Key '%s' has invalide type '%s'", entity.Name, key.Name, key.Type)).PrintfError()
				}
			}
		}
	}
	entity.keys = append(entity.keys, keysUk...)   // UK добавим в список после PK
	entity.keysUk = append(entity.keys, keysUk...) // UK добавим в список после PK
	return nil
}

// initReferencesAddFields - для отношений добавим виртуальные поля, в которые будем парсить данные
func (entity *Entity) initReferencesAddFields() error {
	for _, ref := range entity.references {
		if ref != nil {

			if ref.Status == "" {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - empty 'status'", ref.entity.Name, ref.Name))
			}

			// Только включенные ключи
			if ref.Status == STATUS_ENABLED {

				// Если поля еще не существует
				if entity.Fields.getUnsafe(ref.Name) == nil {

					// Для каждого ссылочного поля REFERENCE_TYPE_ASSOCIATION создаем одноименное дополнительно поле с указателем
					if ref.Type == REFERENCE_TYPE_ASSOCIATION {
						refField := &Field{
							Status:       STATUS_ENABLED,
							Order:        len(entity.Fields) + 1,
							Name:         ref.Name,
							Required:     ref.Required,
							InternalType: FIELD_TYPE_ASSOCIATION,
							Format:       "",
							reference:    ref,
							System:       true,
							ValidateRule: ref.ValidateRule,
							Alias:        ref.Alias, // Имена берем с отношения
							Tag:          ref.Tag,   // tag для разбора берем с отношения
						}
						entity.Fields = append(entity.Fields, refField)
						entity.associationMap[refField.Name] = refField
						ref.field = refField

					} else if ref.Type == REFERENCE_TYPE_COMPOSITION {
						refField := &Field{
							Status:       STATUS_ENABLED,
							Order:        len(entity.Fields) + 1,
							Name:         ref.Name,
							Required:     ref.Required,
							InternalType: FIELD_TYPE_COMPOSITION,
							Format:       "",
							reference:    ref,
							System:       true,
							ValidateRule: ref.ValidateRule,
							Alias:        ref.Alias, // Имена берем с отношения
							Tag:          ref.Tag,   // tag для разбора берем с отношения
						}
						entity.Fields = append(entity.Fields, refField)
						entity.compositionMap[refField.Name] = refField
						ref.field = refField
					} else {
						return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Reference '%s' - incorrect or empty 'type' '%s'", entity.Name, ref.Name, ref.Type))
					}
				}
			}
		} else {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if ref != nil {}", []interface{}{ref}).PrintfError()
		}
	}
	return nil
}

func (entity *Entity) initReferencesSetEntity() error {
	entity.referencesMap = make(ReferencesMap, len(entity.ReferencesDef))
	entity.referencesBy = make(References, 0)
	entity.references = make(References, 0, len(entity.ReferencesDef))

	for _, ref := range entity.ReferencesDef {

		// Только включенные ключи
		if ref.Status == STATUS_ENABLED {

			if _, ok := entity.referencesMap[ref.Name]; ok {
				return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - duplicate Reference '%s'", entity.Name, ref.Name)).PrintfError()
			} else {
				ref.setEntity(entity)
				entity.referencesMap[ref.Name] = ref
				entity.references = append(entity.references, ref)
			}
		}
	}
	return nil
}

func (entity *Entity) initReferencesComplete() error {
	for _, ref := range entity.references {
		// Только включенные ключи
		if ref.Status == STATUS_ENABLED {
			if err := ref.init(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (entity *Entity) initKeyFieldsMap() FieldsMap {
	if entity != nil {
		for _, key := range entity.keys {
			if key != nil {
				for _, field := range key.fields {
					if field != nil {
						entity.keyFieldsMap[field.Name] = field
					}
				}
			}
		}
	}
	return nil
}

func (entity *Entity) initTypeCache() error {
	if entity != nil {
		entity.typeCache = NewTypeCache()

		{ // Все поля включаем в ключ
			entity.defTypeCacheKey = entity.TypeCacheKey(TYPE_CACHE_KEY_PREFIX_ALL, nil) // Все поля включаем в ключ

			rowType, err := entity.StructOf(nil, false, true, true, true, true, true, false, true)
			if err != nil {
				return err
			} else {
				_ = entity.SetTypeCache(entity.defTypeCacheKey, rowType)
			}
		} // Все поля включаем в ключ

		{ // Все поля включаем в ключ без tag
			entity.defTypeCacheKeyEmptyTag = entity.TypeCacheKey(TYPE_CACHE_KEY_PREFIX_EMPTY_TAG, nil) // Все поля включаем в ключ без tag

			rowType, err := entity.StructOf(nil, false, false, false, false, false, false, false, false)
			if err != nil {
				return err
			} else {
				_ = entity.SetTypeCache(entity.defTypeCacheKeyEmptyTag, rowType)
			}
		} // Все поля включаем в ключ без tag

		{ // Все поля включаем в ключ без tag
			entity.defTypeCacheKeyEmptyRef = entity.TypeCacheKey(TYPE_CACHE_KEY_PREFIX_EMPTY_REF, nil) // Все поля включаем в ключ без tag

			rowType, err := entity.StructOf(nil, false, true, true, true, true, true, false, false)
			if err != nil {
				return err
			} else {
				_ = entity.SetTypeCache(entity.defTypeCacheKeyEmptyRef, rowType)
			}
		} // Все поля включаем в ключ без tag

	}
	return nil
}
