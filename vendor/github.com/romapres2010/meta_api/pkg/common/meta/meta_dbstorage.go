package meta

// DbStorage - параметры для хранения в БД
type DbStorage struct {
	StorageName     string `yaml:"storage_name,omitempty" json:"storage_name,omitempty" xml:"storage_name,omitempty"`                // Имя хранилища для сущности
	SchemaName      string `yaml:"schema_name,omitempty" json:"schema_name,omitempty" xml:"schema_name,omitempty"`                   // Имя схемы таблицы для хранения сущности
	TableName       string `yaml:"table_name,omitempty" json:"table_name,omitempty" xml:"table_name,omitempty"`                      // Имя таблицы для хранения сущности
	ObjectName      string `yaml:"object_name,omitempty" json:"object_name,omitempty" xml:"object_name,omitempty"`                   // Имя объекта БД
	ColumnName      string `yaml:"column_name,omitempty" json:"column_name,omitempty" xml:"column_name,omitempty"`                   // Имя поля в таблице для хранения
	ColumnType      string `yaml:"column_type,omitempty" json:"column_type,omitempty" xml:"column_type,omitempty"`                   // Тип поля в таблице для хранения
	ColumnDefault   string `yaml:"column_default,omitempty" json:"column_default,omitempty" xml:"column_default,omitempty"`          // Значение по умолчанию поля в таблице для хранения
	ColumnMandatory bool   `yaml:"column_mandatory,omitempty" json:"column_mandatory,omitempty" xml:"column_mandatory,omitempty"`    // Обязательность поля в таблице для хранения
	DirectSqlSelect string `yaml:"direct_sql_select,omitempty" json:"direct_sql_select,omitempty" xml:"direct_sql_select,omitempty"` // Прямой SQL для доступа к данным таблицы
	DirectSqlWhere  string `yaml:"direct_sql_where,omitempty" json:"direct_sql_where,omitempty" xml:"direct_sql_where,omitempty"`    // Дополнительное правило Where, добавляемое через AND
}

func (storage *DbStorage) copyFrom(from DbStorage, overwrite bool) {
	if storage.StorageName == "" || overwrite {
		storage.StorageName = from.StorageName
	}

	if storage.SchemaName == "" || overwrite {
		storage.SchemaName = from.SchemaName
	}

	if storage.TableName == "" || overwrite {
		storage.TableName = from.TableName
	}

	if storage.ObjectName == "" || overwrite {
		storage.ObjectName = from.ObjectName
	}

	if storage.ColumnName == "" || overwrite {
		storage.ColumnName = from.ColumnName
	}

	if storage.ColumnType == "" || overwrite {
		storage.ColumnType = from.ColumnType
	}

	if storage.ColumnDefault == "" || overwrite {
		storage.ColumnDefault = from.ColumnDefault
	}

	if storage.ColumnMandatory != true || overwrite {
		storage.ColumnMandatory = from.ColumnMandatory
	}

	if storage.DirectSqlSelect == "" || overwrite {
		storage.DirectSqlSelect = from.DirectSqlSelect
	}

	if storage.DirectSqlWhere == "" || overwrite {
		storage.DirectSqlWhere = from.DirectSqlWhere
	}
}
