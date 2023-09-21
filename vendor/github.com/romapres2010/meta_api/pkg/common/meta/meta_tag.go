package meta

import (
	"strings"
)

// Tag - Имена различных тегов для парсинга данных
type Tag struct {
	Db       string `yaml:"db,omitempty" json:"db,omitempty" xml:"db,omitempty"`                      // Имя поля в таблице для хранения
	Json     string `yaml:"json,omitempty" json:"json,omitempty" xml:"json,omitempty"`                // Tag для вывода в Json
	Xml      string `yaml:"xml,omitempty" json:"xml,omitempty" xml:"xml,omitempty"`                   // Tag для вывода в XML
	XmlSpace string `yaml:"xml_space,omitempty" json:"xml_space,omitempty" xml:"xml_space,omitempty"` // Tag для вывода в XML
	Yaml     string `yaml:"yaml,omitempty" json:"yaml,omitempty" xml:"yaml,omitempty"`                // Tag для вывода в YAML
	Xls      string `yaml:"xls,omitempty" json:"xls,omitempty" xml:"xls,omitempty"`                   // Tag для вывода в XLS - заголовок поля
	XlsSheet string `yaml:"xls_sheet,omitempty" json:"xls_sheet,omitempty" xml:"xls_sheet,omitempty"` // Tag для вывода в XLS - имя закладки
	Expr     string `yaml:"expr,omitempty" json:"expr,omitempty" xml:"expr,omitempty"`                // Tag для выполнения выражений
	Sql      string `yaml:"sql,omitempty" json:"sql,omitempty" xml:"sql,omitempty"`                   // Tag для обработки sql

	DbName   string `yaml:"-" json:"-" xml:"-"` // Имя поля в таблице для хранения
	JsonName string `yaml:"-" json:"-" xml:"-"` // Имя для вывода в Json
	XmlName  string `yaml:"-" json:"-" xml:"-"` // Имя для вывода в XML
	YamlName string `yaml:"-" json:"-" xml:"-"` // Имя для вывода в YAML
	XlsName  string `yaml:"-" json:"-" xml:"-"` // Имя для вывода в XLS
	ExprName string `yaml:"-" json:"-" xml:"-"` // Имя для выполнения выражений
}

func (tag *Tag) copyFrom(from Tag, overwrite bool) {
	if tag.Db == "" || overwrite {
		tag.Db = from.Db
	}

	if tag.Json == "" || overwrite {
		tag.Json = from.Json
	}

	if tag.Xml == "" || overwrite {
		tag.Xml = from.Xml
	}

	if tag.XmlSpace == "" || overwrite {
		tag.XmlSpace = from.XmlSpace
	}

	if tag.Yaml == "" || overwrite {
		tag.Yaml = from.Yaml
	}

	if tag.Xls == "" || overwrite {
		tag.Xls = from.Xls
	}

	if tag.Expr == "" || overwrite {
		tag.Expr = from.Expr
	}

	if tag.Sql == "" || overwrite {
		tag.Sql = from.Sql
	}

	if tag.XlsSheet == "" || overwrite {
		tag.XlsSheet = from.XlsSheet
	}
}

func (tag *Tag) init() {
	if tag != nil {
		var index int

		index = strings.IndexByte(tag.Db, ',')
		if index != -1 {
			tag.DbName = tag.Db[:index]
		} else {
			tag.DbName = tag.Db
		}

		index = strings.IndexByte(tag.Json, ',')
		if index != -1 {
			tag.JsonName = tag.Json[:index]
		} else {
			tag.JsonName = tag.Json
		}

		index = strings.IndexByte(tag.Xml, ',')
		if index != -1 {
			tag.XmlName = tag.Xml[:index]
		} else {
			tag.XmlName = tag.Xml
		}

		index = strings.IndexByte(tag.Yaml, ',')
		if index != -1 {
			tag.YamlName = tag.Yaml[:index]
		} else {
			tag.YamlName = tag.Yaml
		}

		index = strings.IndexByte(tag.Xls, ',')
		if index != -1 {
			tag.XlsName = tag.Xls[:index]
		} else {
			tag.XlsName = tag.Xls
		}

		index = strings.IndexByte(tag.Expr, ',')
		if index != -1 {
			tag.ExprName = tag.Expr[:index]
		} else {
			tag.ExprName = tag.Expr
		}

	}
}

func (tag *Tag) GetName(format string) string {
	if tag != nil {
		switch format {
		case "db":
			return tag.DbName
		case "json":
			return tag.JsonName
		case "xml":
			return tag.XmlName
		case "yaml":
			return tag.YamlName
		case "xls":
			return tag.XlsName
		case "expr":
			return tag.ExprName
		default:
			return ""
		}
	}
	return ""
}

func (tag *Tag) GetTag(format string) string {
	if tag != nil {
		switch format {
		case "db":
			return tag.Db
		case "json":
			return tag.Json
		case "xml":
			return tag.Xml
		case "xmlSpace":
			return tag.XmlSpace
		case "yaml":
			return tag.Yaml
		case "xls":
			return tag.Xls
		case "expr":
			return tag.Expr
		default:
			return ""
		}
	}
	return ""
}
