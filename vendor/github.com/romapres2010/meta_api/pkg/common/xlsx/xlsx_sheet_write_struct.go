package xlsx

import (
    "math"
    "reflect"

    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

func (s *Sheet) writeTitlesFromStruct(fields []*Field, newRow bool, titlePrefix string, groupTitles bool, titleGroupName string, autoFilter bool, calcPosition bool, titleRowIndex int) (maxColIndex int, maxRowIndex int, err error) {
    if s != nil && s.xlsx != nil && s.xlsx.file != nil {

        _log.Debug("START: reqID, SheetName, newRow, titlePrefix, groupTitles, titleGroupName", s.reqID, s.name, newRow, titlePrefix, groupTitles, titleGroupName)

        var prevColIndex = s.colIndex                              // запомним предыдущее состояние, чтобы вернуться на него
        var prevRowIndex = s.rowIndex                              // запомним предыдущее состояние, чтобы вернуться на него
        var titleIndexBeg = s.getCellIndex(s.colIndex, s.rowIndex) // Запомни начальную ячейку, с которой начали делать заголовок для группировки
        var titleIndexEnd string
        var count int

        // на одну строку вниз от текущего положения, колонку не меняем
        s.incRowIndex()

        for _, field := range fields {
            //_log.Debug("Process field: reqID, FieldName", s.reqID, field.name)

            // Если есть вложенные объекты, то рекурсия
            if field.fields != nil {
                s.setColIndex(maxColIndex) // Установим на последний заполненный столбец
                _log.Debug("Process cascade field: reqID, FieldName", s.reqID, field.name)
                // Групповым заголовком является xls поля со структурой
                cascadeColIndex, cascadeRowIndex, err := s.writeTitlesFromStruct(field.fields, false, titlePrefix, true, field.title, false, calcPosition, titleRowIndex)
                if err != nil {
                    return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "writeTitlesFromStruct", field).PrintfError()
                }

                maxColIndex = int(math.Max(float64(s.colIndex), float64(cascadeColIndex)))
                maxRowIndex = int(math.Max(float64(s.rowIndex), float64(cascadeRowIndex)))

            } else { // обработка не структурных полей

                if calcPosition { // только вычисляем позицию
                    _ = s.setCellIndex(field)
                    maxColIndex = int(math.Max(float64(s.colIndex), float64(maxColIndex)))
                    maxRowIndex = int(math.Max(float64(s.rowIndex), float64(maxRowIndex)))
                } else {

                    var curRowIndex = s.rowIndex
                    var title string
                    var cellIndexBeg = s.getCellIndex(s.colIndex, s.rowIndex) // Запомни начальную ячейку, с которой начали делать заголовок для группировки
                    var cellIndexEnd string

                    // Если задан уровень выводя заголовка, то выводим на указанной строке
                    if titleRowIndex > 0 {
                        s.setRowIndex(titleRowIndex)
                    }

                    cellIndexEnd = s.getCellIndex(s.colIndex, s.rowIndex)

                    _ = s.setCellIndex(field)
                    maxColIndex = int(math.Max(float64(s.colIndex), float64(maxColIndex)))
                    maxRowIndex = int(math.Max(float64(s.rowIndex), float64(maxRowIndex)))

                    if len(titlePrefix) != 0 && titlePrefix != "" {
                        title = titlePrefix + field.title
                    } else {
                        title = field.title
                    }

                    // установим заголовок
                    err = s.xlsx.file.SetCellStr(s.name, cellIndexBeg, title)
                    if err != nil {
                        return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellStr", title).PrintfError()
                    }

                    // Сгруппируем заголовки по высоте
                    err = s.xlsx.file.MergeCell(s.name, cellIndexBeg, cellIndexEnd)
                    if err != nil {
                        return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "MergeCell", cellIndexBeg, cellIndexEnd).PrintfError()
                    }

                    if s.titleStyle != 0 {
                        err = s.xlsx.file.SetCellStyle(s.name, cellIndexBeg, cellIndexEnd, s.titleStyle)
                        if err != nil {
                            return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellStyle", title).PrintfError()
                        }
                    }

                    count++

                    // Если задан уровень выводя заголовка, то вернем предыдущее состояние
                    if titleRowIndex > 0 {
                        s.setRowIndex(curRowIndex)
                    }
                }
            }
        }

        //if autoFilter {
        //    err = s.xlsx.file.AutoFilter(s.name, titleIndexBeg, titleIndexEnd, "")
        //    if err != nil {
        //        return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, err, s.reqID, "AutoFilter", titleIndexBeg, titleIndexEnd).PrintfError()
        //    }
        //}

        if !calcPosition {
            titleIndexEnd = s.getCellIndex(maxColIndex-1, s.rowIndex-1) // Номер строки берем -1 под группировку заголовка

            if groupTitles && count > 0 {

                _log.Debug("Merge Cells: reqID, SheetName, titleIndexBeg, titleIndexEnd", s.reqID, s.name, titleIndexBeg, titleIndexEnd)
                if len(titleGroupName) != 0 && titleGroupName != "" {
                    err = s.xlsx.file.SetCellStr(s.name, titleIndexBeg, titleGroupName)
                    if err != nil {
                        return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellStr - titleGroupName", titleGroupName).PrintfError()
                    }
                }
                // Сгруппируем заголовки
                err = s.xlsx.file.MergeCell(s.name, titleIndexBeg, titleIndexEnd)
                if err != nil {
                    return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "MergeCell", titleIndexBeg, titleIndexEnd).PrintfError()
                }
                if s.titleStyle != 0 {
                    err = s.xlsx.file.SetCellStyle(s.name, titleIndexBeg, titleIndexEnd, s.titleStyle)
                    if err != nil {
                        return 0, 0, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellStyle", titleIndexBeg, titleIndexEnd).PrintfError()
                    }
                }
            }
        }

        s.setRowIndex(prevRowIndex) // вернем предыдущее положение
        s.setColIndex(prevColIndex) // вернем предыдущее положение

        return maxColIndex, maxRowIndex, nil
    }
    return 0, 0, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, s.reqID, "if s != nil && s.xlsx != nil && s.xlsx.file != nil {}", []interface{}{s}).PrintfError()
}

func (s *Sheet) writeRowFromStruct(fields []*Field, value reflect.Value, newRow bool, opt WriteOption) (err error) {
    if s != nil && s.xlsx != nil && s.xlsx.file != nil {
        _log.Debug("START: reqID, SheetName, newRow", s.reqID, s.name, newRow)

        valueIndirect := reflect.Indirect(value) // может быть передан указатель

        if valueIndirect.Kind() == reflect.Interface { // может быть передан интерфейс, содержащий указатель
            valueIndirect = reflect.Indirect(valueIndirect.Elem())
        }

        // нормальная ситуация - нет данных для некоторых ссылок
        if !valueIndirect.IsValid() {
            return nil
        }

        if newRow {
            s.incRowIndex()
            s.setColIndex(1)
        }

        var count int

        // TODO - цикл сделать по реальным данным, может быть разреженная структура
        for _, field := range fields {

            fieldValue := reflect.Indirect(valueIndirect.FieldByIndex(field.index))

            // Пустые поля не выводим
            if !fieldValue.IsValid() {
                continue
            }

            // Если есть вложенные объекты, то выводим их
            if field.fields != nil {
                _log.Debug("Process cascade field: reqID, FieldName", s.reqID, field.name)
                if field.isStruct {
                    err = s.writeRowFromStruct(field.fields, fieldValue, false, opt)
                    if err != nil {
                        return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "writeRowFromStruct", field).PrintfError()
                    }
                } else if field.isSlice {
                    // TODO - название sheet брать с объекта, а не полей
                    err = s.xlsx.WriteStruct(fieldValue.Interface(), field.sheetName, opt)
                    if err != nil {
                        return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "writeRowFromStruct", field).PrintfError()
                    }
                } else {
                    return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, s.reqID, "if field.fields != nil {} ", []interface{}{field.fieldType, fieldValue.Interface()}).PrintfError()
                }
            } else {
                cellIndex := s.setCellIndex(field)

                if cellIndex != "" {
                    if isNotStruct(fieldValue.Interface()) {
                        err = s.setCellFromValue(cellIndex, field.format, fieldValue, opt)
                        if err != nil {
                            return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "setCellFromValue", fieldValue).PrintfError()
                        } else {
                            count++
                        }
                    } else {
                        return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, s.reqID, "isNotStruct(fieldValueInterface)", []interface{}{field.fieldType, fieldValue.Interface()}).PrintfError()
                    }
                }
            }
        }

        if newRow && count == 0 {
            s.decRowIndex()
            s.setColIndex(1)
        }

        return nil
    }
    return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, s.reqID, "if s != nil && s.xlsx != nil && s.xlsx.file != nil {}", []interface{}{s}).PrintfError()
}
