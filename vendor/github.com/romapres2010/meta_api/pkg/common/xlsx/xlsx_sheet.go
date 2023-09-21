package xlsx

import (
    "fmt"
    "reflect"
    "strconv"
    "time"

    "github.com/xuri/excelize/v2"
    "gopkg.in/guregu/null.v4"

    _err "github.com/romapres2010/meta_api/pkg/common/error"
)

// Sheet - закладка в Excel файле
type Sheet struct {
    xlsx        *Xlsx
    reqID       uint64
    name        string
    transpose   bool   // поменять столбцы и строки
    index       int    // номер закладки Excel
    rowIndex    int    // порядковый номер обрабатываемой строки
    colIndex    int    // порядковый номер обрабатываемой колонки
    cellIndex   string // буквенное наименование обрабатываемой колонки
    titleStyle  int    // номер зарегистрированного стиля Excel
    rowStyle    int    // номер зарегистрированного стиля Excel
    isTitlesSet bool   // признак, что заголовок уже установлен
    //titleRowIndex int    // до какой строки заголовок
}

func NewSheet(xlsx *Xlsx, sheetName string, rowOffset int, transpose bool) (*Sheet, error) {
    if xlsx != nil && xlsx.file != nil {
        var sheet Sheet
        var err error

        sheet.reqID = xlsx.reqID
        sheet.xlsx = xlsx
        sheet.name = sheetName
        sheet.transpose = transpose
        if sheet.index, err = xlsx.file.NewSheet(sheetName); err != nil {
            return nil, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Error create Excel sheet '%s'", sheetName))
        }
        sheet.setStartIndex(rowOffset, 1)

        return &sheet, nil
    } else {
        return nil, nil
    }
}

func (s *Sheet) SetTitleStyle(titleStyle int) {
    s.titleStyle = titleStyle
}

func (s *Sheet) SetRowStyle(rowStyle int) {
    s.rowStyle = rowStyle
}

func (s *Sheet) setColIndex(colIndex int) {
    if s != nil {
        if colIndex >= 1 {
            s.colIndex = colIndex
        } else {
            s.colIndex = 1
        }
    }
}

func (s *Sheet) setRowIndex(rowIndex int) {
    if s != nil {
        if rowIndex >= 0 {
            s.rowIndex = rowIndex
        } else {
            s.rowIndex = 0
        }
    }
}

func (s *Sheet) setStartIndex(rowIndex int, colIndex int) {
    if s != nil {
        s.setRowIndex(rowIndex)
        s.setColIndex(colIndex)
    }
}

func (s *Sheet) setRowIndexOffset(rowIndexOffset int) {
    if s != nil {
        s.rowIndex = s.rowIndex + rowIndexOffset
        if s.rowIndex < 0 {
            s.rowIndex = 0
        }
    }
}

func (s *Sheet) incRowIndex() {
    if s != nil {
        s.setRowIndexOffset(1)
    }
}

func (s *Sheet) decRowIndex() {
    if s != nil {
        s.setRowIndexOffset(-1)
    }
}

func (s *Sheet) getCellIndex(colIndex int, rowIndex int) string {
    if s != nil {

        var cellIndex string
        var err error

        if s.transpose {
            cellIndex, err = excelize.CoordinatesToCellName(rowIndex, colIndex, false) // returns "A1", nil
            if err == nil {
                return cellIndex
            }
        } else {
            cellIndex, err = excelize.CoordinatesToCellName(colIndex, rowIndex, false) // returns "A1", nil
            if err == nil {
                return cellIndex
            }
        }
    }
    return ""
}

func (s *Sheet) setCellIndex(field *Field) string {
    if s != nil {
        cellTag := field.cell

        if cellTag != "" && cellTag != XLSX_TAG_CELL_SKIP {
            if cellTag == XLSX_TAG_CELL_AUTO {
                s.cellIndex = s.getCellIndex(s.colIndex, s.rowIndex)
                s.colIndex++
            } else {
                if colIndex, err := strconv.Atoi(cellTag); err == nil {
                    s.cellIndex = s.getCellIndex(colIndex, s.rowIndex)
                    // TODO Если будут перемешаны нумерованные ячейки и auto, то может сбиться нумерация
                    s.colIndex = colIndex + 1
                } else {
                    s.cellIndex = s.getCellIndex(s.colIndex, s.rowIndex)
                    s.colIndex++
                }
            }
            return s.cellIndex
        }
    }
    return ""
}

func (s *Sheet) setCellFromValue(cellIndex string, format string, value reflect.Value, opt WriteOption) (err error) {
    if s != nil && s.xlsx != nil && s.xlsx.file != nil {
        //_log.Debug("START: reqID, SheetName, FieldName, CellIndex, Value", s.reqID, s.name, format, cellIndex, value)

        // Пустые поля не выводим, кроме стилей
        if value.IsValid() {

            val := value.Interface()

            if fv, ok := toFloat64(val); ok {
                err = s.xlsx.file.SetCellFloat(s.name, cellIndex, fv, opt.FloatPrecision, 64)
                if err != nil {
                    return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellFloat", val).PrintfError()
                }
            } else {
                switch t := val.(type) {
                case time.Time:
                    if format != "" {
                        err = s.xlsx.file.SetCellValue(s.name, cellIndex, t.Format(format))
                    } else {
                        err = s.xlsx.file.SetCellValue(s.name, cellIndex, t)
                    }
                case string:
                    err = s.xlsx.file.SetCellStr(s.name, cellIndex, t)
                case bool:
                    err = s.xlsx.file.SetCellBool(s.name, cellIndex, t)
                case null.String:
                    if t.Valid {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, t.String)
                    } else {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                    }
                case null.Float:
                    if t.Valid {
                        err = s.xlsx.file.SetCellFloat(s.name, cellIndex, t.Float64, opt.FloatPrecision, 64)
                    } else {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                    }
                case null.Int:
                    if t.Valid {
                        err = s.xlsx.file.SetCellInt(s.name, cellIndex, int(t.Int64))
                    } else {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                    }
                case null.Bool:
                    if t.Valid {
                        err = s.xlsx.file.SetCellBool(s.name, cellIndex, t.Bool)
                    } else {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                    }
                case null.Time:
                    if t.Valid {
                        if format != "" {
                            err = s.xlsx.file.SetCellValue(s.name, cellIndex, t.Time.Format(format))
                        } else {
                            err = s.xlsx.file.SetCellValue(s.name, cellIndex, t.Time)
                        }
                    } else {
                        err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                    }
                case nil:
                    err = s.xlsx.file.SetCellStr(s.name, cellIndex, "")
                default:
                    err = s.xlsx.file.SetCellValue(s.name, cellIndex, t)
                }
            }
            //_log.Debug("SUCCESS: reqID, SheetName, FieldName, CellIndex, Value", s.reqID, s.name, format, cellIndex, value)

            if err != nil {
                return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "setCellFromValue", val).PrintfError()
            }
        }

        if s.rowStyle != 0 {
            err = s.xlsx.file.SetCellStyle(s.name, cellIndex, cellIndex, s.rowStyle)
            if err != nil {
                return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, s.reqID, err, "SetCellStyle", cellIndex, cellIndex).PrintfError()
            }
        }
        return nil

    }
    return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, s.reqID, "if s != nil && s.xlsx != nil && s.xlsx.file != nil {}", []interface{}{s}).PrintfError()
}
