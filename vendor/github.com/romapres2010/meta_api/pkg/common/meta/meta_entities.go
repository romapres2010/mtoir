package meta

type Entities []*Entity

type EntityMap map[string]*Entity

func newEntityMap() EntityMap {
	return make(EntityMap)
}

func (ent EntityMap) Set(entityName string, entity *Entity) {
	if ent != nil && entity != nil {
		ent[entityName] = entity
	}
}

func (ent EntityMap) Delete(entityName string) {
	if ent != nil {
		delete(ent, entityName)
	}
}

func (ent EntityMap) GetEntitySafe(entityName string) *Entity {
	if v, ok := ent[entityName]; ok {
		return v
	} else {
		return &entityNotFound
	}
}

func (ent EntityMap) GetEntity(entityName string) *Entity {
	if v, ok := ent[entityName]; ok {
		return v
	} else {
		return nil
	}
}

func (ent EntityMap) GetDisplayName(entityName string) string {
	entity := ent.GetEntitySafe(entityName)
	return entity.Alias.DisplayName
}

func (ent EntityMap) GetFullName(entityName string) string {
	entity := ent.GetEntitySafe(entityName)
	if entity.Alias.FullName == ENTITY_NOT_FOUND {
		return entityName
	} else {
		return entity.Alias.FullName
	}
}

func (ent EntityMap) GetFieldDisplayName(entityName string, fieldName string) string {
	return ent.GetEntitySafe(entityName).FieldDisplayName(fieldName)
}

func (ent EntityMap) GetFieldFullName(entityName string, fieldName string) string {
	return ent.GetEntitySafe(entityName).FieldFullName(fieldName)
}
