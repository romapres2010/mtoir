package meta

import (
	"fmt"
	"strings"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

func (entity *Entity) inheritFromName(name string) (err error) {
	if entity != nil && name != "" {

		if name != "" {

			//if entity.Name == "Asset" {
			//	_log.Debug("if name == \"Asset\" {")
			//}

			entityNames := strings.Split(name, ",")

			for _, entityName := range entityNames {

				// Найти сущность
				fromEntity := entity.meta.GetEntityUnsafe(entityName)
				if fromEntity == nil {
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - 'inherit_from'='%s' was not found", entity.Name, entityName))
				}

				{ // По всем объектам родительской сущности

					entity.Alias.copyFrom(fromEntity.Alias, false)

					entity.DbStorage.copyFrom(fromEntity.DbStorage, false)

					entity.Tag.copyFrom(fromEntity.Tag, false)

					{ // Наследуем поля
						for _, from := range fromEntity.Fields {

							// копируем только включенные и не системные
							if from.Status != STATUS_ENABLED || from.System {
								continue
							}

							// Если существует в текущей сущности, то скопировать незаполненное из наследованной, иначе создать как копию
							if field := entity.fieldByNameUnsafe(from.Name); field == nil {
								newField := entity.newField(from.Name)
								if err = newField.inheritFromField(from); err != nil {
									return err
								} else {
									entity.addField(newField)
								}
							} else {
								if field.CopyFromName == "" { // Явное CopyFrom имеет приоритет над InheritFrom
									field.entity = entity // TODO - определить место присвоения
									// TODO - если повторная инициация сущности, то дублируются Expr
									if err = field.copyFromField(from); err != nil {
										return err
									}
								}
							}

						}
					} // Наследуем поля

					{ // Наследуем key из определения
						for _, from := range fromEntity.KeysDef {

							// копируем только включенные
							if from.Status != STATUS_ENABLED {
								continue
							}

							// Если существует в текущей сущности, то скопировать незаполненное из наследованной, иначе создать как копию
							if key := entity.keyDefByNameUnsafe(from.Name); key == nil {
								newKey := entity.newKey(from.Name)
								if err = newKey.inheritFromKey(from); err != nil {
									return err
								} else {
									entity.addKeyDef(newKey)
								}
							} else {
								if key.CopyFromName == "" { // Явное CopyFrom имеет приоритет над InheritFrom
									key.setEntity(entity) // TODO - определить место присвоения
									if err = key.copyFromKey(from); err != nil {
										return err
									}
								}
							}

						}
					} // Наследуем key из определения

					{ // Наследуем reference
						for _, from := range fromEntity.ReferencesDef {

							// копируем только включенные
							if from.Status != STATUS_ENABLED {
								continue
							}

							// Если существует в текущей сущности, то скопировать незаполненное из наследованной, иначе создать как копию
							if ref := entity.referenceDefByNameUnsafe(from.Name); ref == nil {
								newReference := entity.newReference(from.Name)
								if err = newReference.inheritFromReference(from); err != nil {
									return err
								} else {
									entity.addReferenceDef(newReference)
								}
							} else {
								if ref.CopyFromName == "" { // Явное CopyFrom имеет приоритет над InheritFrom
									ref.setEntity(entity) // TODO - определить место присвоения
									if err = ref.copyFromReference(from); err != nil {
										return err
									}
								}
							}

						}
					} // Наследуем reference

				} // По всем объектам родительской сущности

				// Все сущности наследованные от данной
				entity.inheritFrom = append(entity.inheritFrom, fromEntity)
				fromEntity.inheritTo = append(fromEntity.inheritTo, entity)
			}
		}

		entity.isInit = false

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && name != \"\" {}", []interface{}{entity, name}).PrintfError()
}
