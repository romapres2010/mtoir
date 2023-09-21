package meta

import (
    "sync"
)

const STATUS_ENABLED = "ENABLED"
const STATUS_DISABLE = "DISABLE"
const STATUS_DEPRECATED = "DEPRECATED"

type Meta struct {
    Status   string   `yaml:"status,omitempty" json:"status,omitempty" xml:"status,omitempty"`       // Статус  ENABLED, DEPRECATED, ...
    Name     string   `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`             // Имя меты
    Entities Entities `yaml:"entities,omitempty" json:"entities,omitempty" xml:"entities,omitempty"` // Сущности в "правильном" порядке для создания

    entityMap EntityMap // Сущности для быстрого поиска
    isInit    bool      // Признак, что инициация успешная

    mx sync.RWMutex
}

func NewMeta() *Meta {
    meta := &Meta{
        entityMap: newEntityMap(),
    }
    return meta
}

func (meta *Meta) Lock() {
    if meta != nil {
        meta.mx.Lock()
    }
}

func (meta *Meta) Unlock() {
    if meta != nil {
        meta.mx.Unlock()
    }
}

func (meta *Meta) RLock() {
    if meta != nil {
        meta.mx.RLock()
    }
}

func (meta *Meta) RUnlock() {
    if meta != nil {
        meta.mx.RUnlock()
    }
}

func (meta *Meta) GetEntitySafe(entityName string) *Entity {
    if meta != nil {
        return meta.entityMap.GetEntitySafe(entityName)
    }
    return nil
}

func (meta *Meta) GetEntity(entityName string) *Entity {
    if meta != nil {
        meta.mx.RLock()
        defer meta.mx.RUnlock()

        return meta.entityMap.GetEntity(entityName)
    }
    return nil
}

func (meta *Meta) GetEntityUnsafe(entityName string) *Entity {
    if meta != nil {
        return meta.entityMap.GetEntity(entityName)
    }
    return nil
}

func (meta *Meta) GetDisplayName(entityName string) string {
    if meta != nil {
        return meta.entityMap.GetDisplayName(entityName)
    }
    return ""
}

func (meta *Meta) GetFullName(entityName string) string {
    if meta != nil {
        return meta.entityMap.GetFullName(entityName)
    }
    return ""
}

func (meta *Meta) GetFieldDisplayName(entityName string, fieldName string) string {
    if meta != nil {
        return meta.entityMap.GetFieldDisplayName(entityName, fieldName)
    }
    return ""
}

func (meta *Meta) GetFieldFullName(entityName string, fieldName string) string {
    if meta != nil {
        return meta.entityMap.GetFieldFullName(entityName, fieldName)
    }
    return ""
}
