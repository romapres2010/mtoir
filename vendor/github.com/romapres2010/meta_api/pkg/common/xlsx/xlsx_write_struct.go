package xlsx

import (
	"reflect"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
)

func (xls *Xlsx) WriteStruct(val interface{}, sheetNameIn string, opt WriteOption) (err error) {
	if xls != nil {

		var sheet *Sheet
		var sheetName string
		var ok bool
		var titleStyle int
		var cellStyle int

		if titleStyle, err = xls.file.NewStyle(gTitleStyle); err != nil {
			return err
		}

		if cellStyle, err = xls.file.NewStyle(gCellStyle); err != nil {
			return err
		}

		value := reflect.Indirect(reflect.ValueOf(val))
		if value.Kind() == reflect.Interface { // может быть передан интерфейс, содержащий указатель на struct или slice
			value = reflect.Indirect(value.Elem())
		}

		_log.Debug("START: reqID, WriteOption, value.Type", xls.reqID, opt, value.Type().String())

		var valueOne reflect.Value
		var slice bool

		if value.Kind() == reflect.Slice {
			if opt.CascadeStruct {

				// Для анализа нужно выбрать элемент с максимальной вложенностью для отображения иерархических структур
				if value.Len() > 0 {
					slice = true
					maxLevelIndex := 0
					maxLevel := 0

					valueLen := value.Len()
					for i := 0; i < valueLen; i++ {
						v := value.Index(i)

						level := inspectStructLevel(v)
						if level > maxLevel {
							maxLevel = level
							maxLevelIndex = i
						}
					}

					valueOne = value.Index(maxLevelIndex)

					_log.Debug("Inspect level: maxLevelIndex", maxLevelIndex)
				} else {
					return nil // пустой Slice - нечего обрабатывать
				}
			} else {
				valueOne = value.Index(0) // Если отключена каскадная обработка, то достаточно первого элемента в массиве
			}
		} else {
			slice = false
			valueOne = value
		}

		_log.Debug("ValueOne:", valueOne.Type().String())

		fields := inspectStructFields(valueOne, opt.CascadeStruct)

		_log.Debug("Inspect fields:", fields)

		if sheetNameIn != "" {
			sheetName = sheetNameIn
		} else {
			//sheetName = findSheetName(valueOne.Interface())
			sheetName = firstSheetName(fields)
		}

		// Если находим первое упоминание Sheet, то создаем новый лист
		if sheetName != "" {
			// Если sheet уже существует, то второй раз не создавать
			if sheet, ok = xls.sheets[sheetName]; ok {
				_log.Debug("Sheet already exists: reqID, SheetName", xls.reqID, sheetName)
			} else {
				_log.Debug("Sheet does non exists - create one: reqID, SheetName", xls.reqID, sheetName)
				if sheet, err = NewSheet(xls, sheetName, 0, opt.Transpose); err != nil {
					return err
				}
				xls.sheets[sheetName] = sheet
			}

			// Установить стили для заголовка и строк
			sheet.SetTitleStyle(titleStyle)
			sheet.SetRowStyle(cellStyle)

			// TODO в русском языке может быть другое наименование
			_ = xls.file.DeleteSheet("Sheet1")
		} else {
			return _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, xls.reqID, err, "SheetName was not defined").PrintfError()
		}

		if sheet != nil {

			// Добавить titles, если они еще не установлены
			if opt.SetTitles && !sheet.isTitlesSet {

				// Заголовок выводим с левого верхнего угла
				sheet.setColIndex(1)
				sheet.setRowIndex(1)

				titleRowIndex := 1

				// первым проходом рекурсивно вычислить глубину заголовка
				_, titleRowIndex, err = sheet.writeTitlesFromStruct(fields, opt.NewRow, opt.TitlePrefix, true, sheetName, opt.AutoFilter, true, 0)
				if err != nil {
					return err
				}

				// Вторым проходом заполняем заголовок
				_, _, err = sheet.writeTitlesFromStruct(fields, opt.NewRow, opt.TitlePrefix, true, sheetName, opt.AutoFilter, false, titleRowIndex)
				if err != nil {
					return err
				}

				// Данные выводим сразу под заголовком
				sheet.setColIndex(1)
				sheet.setRowIndex(titleRowIndex)

				sheet.isTitlesSet = true
			}

			// Для массива обработать все элементы в цикле
			if slice {
				valueLen := value.Len()
				for i := 0; i < valueLen; i++ {
					_log.Debug("Write row: rowNum, sheet.Name, sheet.ColIndex, sheet.RowIndex", i, sheet.name, sheet.colIndex, sheet.rowIndex)

					err = sheet.writeRowFromStruct(fields, reflect.Indirect(value.Index(i)), true, opt)
					if err != nil {
						return err
					}
				}
			} else {
				_log.Debug("Write row: sheet.Name, sheet.ColIndex, sheet.RowIndex", sheet.name, sheet.colIndex, sheet.rowIndex)

				err = sheet.writeRowFromStruct(fields, reflect.Indirect(value), opt.NewRow, opt)
				if err != nil {
					return err
				}
			}
			return nil
		} else {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, xls.reqID, "sheet != nil", []interface{}{sheet}).PrintfError()
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "s != nil", []interface{}{xls}).PrintfError()
}
