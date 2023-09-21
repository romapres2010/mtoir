package meta

// Alias - дополнительные имена и описание
type Alias struct {
	DisplayName  string `yaml:"display_name,omitempty" json:"display_name,omitempty" xml:"display_name,omitempty"`    // Имя для отображения
	FullName     string `yaml:"full_name,omitempty" json:"full_name,omitempty" xml:"full_name,omitempty"`             // Полное имя для вывода в логи и ошибки
	Description  string `yaml:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`       // Дополнительное описание
	ExternalName string `yaml:"external_name,omitempty" json:"external_name,omitempty" xml:"external_name,omitempty"` // Имя во внешней системе
}

func (alias *Alias) copyFrom(from Alias, overwrite bool) {
	if alias.DisplayName == "" || overwrite {
		alias.DisplayName = from.DisplayName
	}

	if alias.FullName == "" || overwrite {
		alias.FullName = from.FullName
	}

	if alias.Description == "" || overwrite {
		alias.Description = from.Description
	}

	if alias.ExternalName == "" || overwrite {
		alias.ExternalName = from.ExternalName
	}
}
