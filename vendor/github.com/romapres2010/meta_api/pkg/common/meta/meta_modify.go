package meta

// Modify - параметры управления модификацией объекта
type Modify struct {
    CreateRestrict   bool `yaml:"create_restrict,omitempty" json:"create_restrict,omitempty" xml:"create_restrict,omitempty"`       // Признак, что запрещено вставлять данные в сущность или поле
    RetrieveRestrict bool `yaml:"retrieve_restrict,omitempty" json:"retrieve_restrict,omitempty" xml:"retrieve_restrict,omitempty"` // Признак, что запрещено считывать данные в сущность или поле
    UpdateRestrict   bool `yaml:"update_restrict,omitempty" json:"update_restrict,omitempty" xml:"update_restrict,omitempty"`       // Признак, что запрещено обновлять данные в сущность или поле
    DeleteRestrict   bool `yaml:"delete_restrict,omitempty" json:"delete_restrict,omitempty" xml:"delete_restrict,omitempty"`       // Признак, что запрещено удалять данные в сущность или поле
    CopyRestrict     bool `yaml:"copy_restrict,omitempty" json:"copy_restrict,omitempty" xml:"copy_restrict,omitempty"`             // Признак, что запрещено копировать сущность или поле при внутренней обработке
    VisibleRestrict  bool `yaml:"visible_restrict,omitempty" json:"visible_restrict,omitempty" xml:"visible_restrict,omitempty"`    // Признак, что запрещено показывать полей
}
