package xlsx

import (
    "reflect"

    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

const (
    XLSX_TAG_CELL      = "cell"
    XLSX_TAG_CELL_AUTO = "auto"
    XLSX_TAG_CELL_SKIP = "-"
    XLSX_TAG_TITLE     = "title"
    XLSX_TAG_SHEET     = "sheet"
    XLSX_TAG_FORMAT    = "format"
    XLSX_TAG_LINK      = "link"
)

// Field - информация об одном поле
type Field struct {
    name      string
    index     []int
    title     string
    sheetName string
    cell      string
    link      string
    format    string
    isStruct  bool     // признак, что поле из каскадной структуры
    isSlice   bool     // признак, что поле из каскадной структуры
    fields    []*Field // вложенная в поле структура
    fieldType reflect.StructField
    //value     reflect.Value
}

func firstSheetName(fields []*Field) string {
    for _, field := range fields {
        if field.sheetName != "" {
            return field.sheetName
        }
    }
    return ""
}

func inspectStructFields(value reflect.Value, doCascade bool) (fields []*Field) {

    valueIndirect := reflect.Indirect(value) // может быть передан указатель на структуру

    _log.Debug("START: valueKind", valueIndirect.String())

    if valueIndirect.Kind() == reflect.Interface { // может быть передан интерфейс, содержащий указатель на структуру
        valueIndirect = reflect.Indirect(valueIndirect.Elem())
    }

    valueType := valueIndirect.Type()
    // TODO - название sheet брать с объекта, а не полей

    if valueType.Kind() != reflect.Struct {
        return nil
    }

    fields = make([]*Field, 0, valueType.NumField())

    for i := 0; i < valueType.NumField(); i++ {

        fieldType := valueType.Field(i)
        cell := fieldType.Tag.Get(XLSX_TAG_CELL)

        if cell != "" && cell != XLSX_TAG_CELL_SKIP {

            fieldValue := reflect.Indirect(valueIndirect.FieldByIndex(fieldType.Index))

            if !fieldValue.IsValid() {
                _log.Debug("field is empty or not valid", fieldType.Name)
                continue
            }

            field := &Field{
                name:      fieldType.Name,
                index:     fieldType.Index,
                title:     fieldType.Tag.Get(XLSX_TAG_TITLE),
                sheetName: fieldType.Tag.Get(XLSX_TAG_SHEET),
                cell:      cell,
                link:      fieldType.Tag.Get(XLSX_TAG_LINK),
                format:    fieldType.Tag.Get(XLSX_TAG_FORMAT),
                fieldType: fieldType,
            }

            if fieldValue.Kind() == reflect.Interface && fieldValue.IsNil() {
                continue
            }

            fieldValueInterface := fieldValue.Interface()

            if doCascade {

                if !isNotStruct(fieldValueInterface) {

                    // вложенная структура или указатель на структуру
                    field.isStruct = true
                    field.fields = inspectStructFields(fieldValue, true)

                } else if IsSlice(fieldValueInterface) {

                    // вложенный массив или указатель на массив
                    if reflect.Indirect(reflect.ValueOf(fieldValueInterface)).Len() > 0 {
                        field.isSlice = true
                        sliceValue := reflect.ValueOf(fieldValue.Interface())
                        valueOne := reflect.Indirect(sliceValue).Index(0)
                        field.fields = inspectStructFields(valueOne, true)
                    }

                }
            }

            fields = append(fields, field)
        }
    }

    return fields
}

func inspectStructLevel(value reflect.Value) (level int) {

    valueIndirect := reflect.Indirect(value) // может быть передан указатель на структуру

    //_log.Debug("START: valueKind", reflect.Indirect(value).Kind().String())

    if valueIndirect.Kind() == reflect.Interface { // может быть передан интерфейс, содержащий указатель на структуру
        valueIndirect = reflect.Indirect(valueIndirect.Elem())
    }

    valueType := valueIndirect.Type()

    if valueType.Kind() != reflect.Struct {
        return 0
    }

    level = 1

    // проверим все поля, если есть вложенные указатели - пойдем в них
    for i := 0; i < valueType.NumField(); i++ {

        fieldType := valueType.Field(i)
        cell := fieldType.Tag.Get(XLSX_TAG_CELL)

        if cell != "" && cell != XLSX_TAG_CELL_SKIP {

            fieldValue := reflect.Indirect(valueIndirect.FieldByIndex(fieldType.Index))

            if !fieldValue.IsValid() {
                continue
            }

            if fieldValue.Kind() == reflect.Interface && fieldValue.IsNil() {
                continue
            }

            // вложенная структура или указатель на структуру
            if !isNotStruct(fieldValue.Interface()) {
                innerLevel := inspectStructLevel(fieldValue)
                level += innerLevel
            }
        }
    }

    return level
}
