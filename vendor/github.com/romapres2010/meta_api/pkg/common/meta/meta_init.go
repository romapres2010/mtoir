package meta

import (
	"fmt"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
)

func (meta *Meta) Init() error {
	if meta != nil {
		meta.mx.Lock()
		defer meta.mx.Unlock()

		return meta.InitUnsafe()
	}
	return nil
}

func (meta *Meta) Backup() *Meta {
	if meta != nil {

		copyMeta := &Meta{
			Status:    meta.Status,
			Name:      meta.Name,
			Entities:  make(Entities, len(meta.Entities)),
			entityMap: newEntityMap(),
			isInit:    meta.isInit,
		}

		copy(copyMeta.Entities, meta.Entities)

		return copyMeta
	}
	return nil
}

func (meta *Meta) Restore(copyMeta *Meta) {
	if meta != nil && copyMeta != nil && meta.Name == copyMeta.Name {
		meta.Status = copyMeta.Status
		meta.Entities = copyMeta.Entities
		meta.entityMap = copyMeta.entityMap
		meta.isInit = copyMeta.isInit
	}
}

func (meta *Meta) InitUnsafe() error {
	if meta != nil {

		// Сохраним предыдущее состояние
		copyEntities := make(Entities, len(meta.Entities))
		copy(copyEntities, meta.Entities)

		// Создаем новые структуры для хранения
		meta.Entities = make(Entities, 0, len(meta.Entities))
		meta.entityMap = newEntityMap()

		// Пересоздаем сущность из definition и наполняем entityMap
		for _, oldEntity := range copyEntities {
			// Создаем новую сущность из определения старой
			entity, err := meta.newFromDefinition(oldEntity.Definition)
			if err != nil {
				return err
			}

			if err = entity.init(meta); err != nil {
				return err
			}

			meta.Entities = append(meta.Entities, entity)
			meta.entityMap.Set(entity.Name, entity)

			//if err = entity.initReferencesComplete(); err != nil {
			//	return err
			//}

			//if err = entity.initTypeCache(); err != nil {
			//	return err
			//}
		}

		// Пересоздаем ReferencesDef
		for _, entity := range meta.Entities {
			if err := entity.initReferencesComplete(); err != nil {
				return err
			}
		}

		// Пересоздаем Type Cache
		for _, entity := range meta.Entities {
			if err := entity.initTypeCache(); err != nil {
				return err
			}
		}

		meta.isInit = true

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil {}", []interface{}{meta}).PrintfError()
}

func (meta *Meta) SetEntityFromDefinitionUnsafe(definition *Definition, doReplace bool, doInit bool) (*Entity, error) {
	if meta != nil && definition != nil {

		if definition.Name == "" {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Meta '%s' - empty Entity name in definition '%s'", meta.Name, *definition))
		}

		_log.Debug("START: entityName, doReplace, doInit", definition.Name, doReplace, doInit)

		// проверка корректности definition
		if err := definition.check(); err != nil {
			return nil, err
		}

		// Создаем новую сущность из определения
		entity, err := meta.newFromDefinition(definition)
		if err != nil {
			return nil, err
		}

		if err = meta.setEntityUnsafe(entity, doReplace, doInit); err != nil {
			return nil, err
		}

		return entity, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil && definition != nil {}", []interface{}{meta, definition}).PrintfError()
}

func (meta *Meta) setEntityUnsafe(entity *Entity, doReplace bool, doInit bool) error {
	if meta != nil && entity != nil {

		if entity.Name == "" {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Meta '%s' - empty Entity name", meta.Name))
		}

		_log.Debug("START: entityName, doReplace, doInit", entity.Name, doReplace, doInit)

		meta.entityMap.Set(entity.Name, entity)

		if doInit {
			// Инициация сущности, кроме reference
			if err := entity.init(meta); err != nil {
				meta.entityMap.Delete(entity.Name)
				return err
			}

			// Инициация reference сущности
			if err := entity.initReferencesComplete(); err != nil {
				meta.entityMap.Delete(entity.Name)
				return err
			}

			// Инициация Type Cache
			if err := entity.initTypeCache(); err != nil {
				meta.entityMap.Delete(entity.Name)
				return err
			}
		}

		replacedId := -1
		for i, _ := range meta.Entities {
			if meta.Entities[i].Name == entity.Name {
				if doReplace {
					replacedId = i
					meta.Entities[i] = entity
					break
				} else {
					meta.entityMap.Delete(entity.Name)
					return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Meta '%s' Entity '%s' - already exists", meta.Name, entity.Name))
				}
			}
		}

		// Сущности не существует
		if replacedId == -1 {
			meta.Entities = append(meta.Entities, entity)
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil && entity != nil {}", []interface{}{meta, entity}).PrintfError()
}

func (meta *Meta) Set(entity *Entity, doReplace bool, doInit bool) error {
	if meta != nil && entity != nil {

		_log.Info("START: entityName, doReplace, doInit", entity.Name, doReplace, doInit)
		meta.mx.Lock()
		defer meta.mx.Unlock()
		_log.Info("START - after lock: entityName, doReplace, doInit", entity.Name, doReplace, doInit)

		return meta.setEntityUnsafe(entity, doReplace, doInit)
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil && entity != nil {}", []interface{}{meta, entity}).PrintfError()
}
