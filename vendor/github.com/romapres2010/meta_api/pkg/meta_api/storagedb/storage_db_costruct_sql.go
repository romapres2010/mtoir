package storagedb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// constructSqlSelect сформировать sql для запроса на основании параметров внешнего запроса
func (s *Storage) constructWhereByKey(requestID uint64, rowOut *_meta.Object, parameterChar string, argCnt int, key *_meta.Key, keyArgs ...interface{}) (wheres []string, sqlArgs []interface{}, err error) {
	if s != nil && rowOut != nil {

		tic := time.Now()
		entity := rowOut.Entity

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Добавим условие поиска по ключу
		if key != nil {
			if len(key.Fields()) != len(keyArgs) {
				return nil, nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Key '%s' - incorrect number of argument. Requed fields '%s'", entity.Name, key.Name, key.FieldsString()))
			}

			var wheresKey = make([]string, 0, len(key.Fields()))

			// Добавим поля ключа в Where
			for i, field := range key.Fields() {
				keyFieldDbColumnName := field.DbStorage.ColumnName
				if keyFieldDbColumnName != "" {
					if _meta.ArgNotEmpty(keyArgs[i]) {
						// не пустой аргумент
						wheresKey = append(wheresKey, keyFieldDbColumnName+"="+parameterChar+strconv.Itoa(argCnt))
						sqlArgs = append(sqlArgs, keyArgs[i])
						//sqlArgs = append(sqlArgs, _meta.ArgToString(keyArgs[i])) // в Oracle не работает UUID -> string
						argCnt++
					} else {
						// пустой аргумент
						wheresKey = append(wheresKey, keyFieldDbColumnName+" IS NULL")
					}
				} else {
					return nil, nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Key '%s' field '%s' - 'db.column_name' was not defined", entity.Name, key.Name, field.Name))
				}
			}

			wheres = append(wheres, wheresKey...)

			// Фиксированный блок из meta
			if key.DbStorage.DirectSqlWhere != "" {
				wheres = append(wheres, "("+key.DbStorage.DirectSqlWhere+")")
			}

		} else {
			return nil, nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - Key was not defined", entity.Name))
		}

		_log.Debug("SUCCESS: requestID, entityName, wheres, args, duration", requestID, entity.Name, wheres, sqlArgs, time.Now().Sub(tic))

		return wheres, sqlArgs, err
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && && rowOut != nil {}", []interface{}{s, rowOut}).PrintfError()
}

// constructSqlSelect сформировать sql для запроса на основании параметров внешнего запроса
func (s *Storage) constructSqlSelect(requestID uint64, rowOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (sqlFull string, sqlArgs []interface{}, err error) {
	if s != nil && rowOut != nil {

		tic := time.Now()
		entity := rowOut.Entity
		outFields := rowOut.Fields
		options := rowOut.Options
		argCnt := 1
		sqlFrom := ""
		sqlColumn := ""
		sqlTable := ""
		sqlWhere := "WHERE 1=1"
		sqlOrder := ""
		sqlLimit := ""
		sqlOffset := ""
		columns := make([]string, 0, len(entity.StructFields()))
		whereOption := options.DbWhere
		orderOption := options.DbOrder
		limitOption := options.DbLimit
		offsetOption := options.DbOffset
		nameFormatOption := options.Global.NameFormat
		restrict := outFields != nil && len(outFields) > 0
		directFrom := entity.DbStorage.DirectSqlSelect != ""
		parameterChar := "$"

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		switch s.dbCfg.DriverName {
		case "postgres", "pgx":
			parameterChar = "$"
		case "godror", "oracle":
			parameterChar = ":"
		default:
			parameterChar = "$"
		}

		// Если не задано, то формат берем из запроса, иначе "name"
		if nameFormatOption == "" {
			// по умолчанию в именах полей
			nameFormatOption = "name"
		}

		if directFrom {
			//sqlTable = "(" + entity.DbStorage.DirectSqlSelect + ") AS " + entity.Name
			sqlTable = "(" + entity.DbStorage.DirectSqlSelect + ")"
		} else {
			if entity.DbStorage.TableName == "" {
				err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' has not defined 'db.table_name'", entity.Name))
				return "", nil, err
			}

			if entity.DbStorage.SchemaName != "" {
				sqlTable = entity.DbStorage.SchemaName + "." + entity.DbStorage.TableName
			} else {
				sqlTable = entity.DbStorage.TableName
			}
		}

		{ // SELECT clause
			for _, field := range entity.StructFields() {
				// Если не заполнен DB - из БД не извлекать
				if field.Tag.Db != "" && field.Tag.Db != "-" && field.DbStorage.ColumnName != "" {

					// Разрешено ли запрашивать поля
					if field.Modify.RetrieveRestrict {
						continue
					}

					// Ограничить поля определенным списком
					if restrict {
						if _, ok := outFields[field.Name]; !ok {
							continue
						}
					}

					// Имена в БД и в структуре могут отличаться
					columns = append(columns, field.DbStorage.ColumnName+" as \""+field.Tag.Db+"\"")
				}
			}

			sqlColumn = strings.Join(columns, ", ")
			sqlFrom = "SELECT " + sqlColumn + " FROM " + sqlTable

		} // SELECT clause

		{ // WHERE clause
			var wheres []string

			// Добавим условие поиска по ключу
			if keyArgs != nil && len(keyArgs) != 0 {
				if wheresKey, wheresKeyArgs, err := s.constructWhereByKey(requestID, rowOut, parameterChar, 1, key, keyArgs...); err != nil {
					return "", nil, err
				} else {
					wheres = append(wheres, wheresKey...)
					sqlArgs = append(sqlArgs, wheresKeyArgs...)
				}
			}

			// Добавим условия поиска по остальным полям
			for fieldName, val := range options.DbFieldsWhere {
				if len(val) != 0 {

					if field := entity.FieldByTagName(nameFormatOption, fieldName); field == nil {
						// не найденные поля игнорируем, только если в них есть "." - это может относиться к вложенным сущностям
						if !strings.ContainsAny(fieldName, ".") {
							return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' has not have enabled field '%s'", entity.Name, fieldName))
						} else {
							continue
						}
					} else {
						if field.DbStorage.ColumnName != "" {
							wheres = append(wheres, field.DbStorage.ColumnName+"="+parameterChar+strconv.Itoa(argCnt))
						} else {
							return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' field '%s' has not have 'db.column_name'", entity.Name, fieldName))
						}
					}

					sqlArgs = append(sqlArgs, val)
					argCnt++
				}
			}

			// Фиксированный блок из meta
			if entity.DbStorage.DirectSqlWhere != "" {
				wheres = append(wheres, "("+entity.DbStorage.DirectSqlWhere+")")
			}

			// Вставка из запроса
			if whereOption != "" {
				wheres = append(wheres, whereOption)
			}

			if len(wheres) > 0 {
				sqlWhere = sqlWhere + " and " + strings.Join(wheres, " and ")
			}

		} // WHERE clause

		{ // ORDER BY clause
			if orderOption != "" {
				var orderFields = strings.Split(orderOption, ",")
				var orderDirection string
				var fieldName string
				var orders = make([]string, 0, len(orderFields))

				for _, orderField := range orderFields {

					if strings.HasSuffix(orderField, "-") {
						fieldName = strings.TrimSuffix(orderField, "-")
						orderDirection = "desc"
					} else if strings.HasSuffix(orderField, "+") {
						fieldName = strings.TrimSuffix(orderField, "+")
						orderDirection = "asc"
					} else if strings.HasSuffix(orderField, " desc") {
						fieldName = strings.TrimSuffix(orderField, " desc")
						orderDirection = "desc"
					} else if strings.HasSuffix(orderField, " asc") {
						fieldName = strings.TrimSuffix(orderField, " asc")
						orderDirection = "asc"
					} else {
						fieldName = orderField
						orderDirection = "asc"
					}

					if field := entity.FieldByTagName(nameFormatOption, fieldName); field == nil {
						// не найденные поля игнорируем, только если в них есть "." - это может относиться к вложенным сущностям
						if !strings.ContainsAny(fieldName, ".") {
							err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' has not have enabled field '%s'", entity.Name, fieldName))
							return "", nil, err
						}
					} else {
						if field.DbStorage.ColumnName != "" {
							orders = append(orders, strings.Join([]string{field.DbStorage.ColumnName, orderDirection}, " "))
							//} else if field.Tag.DbStorage != "" && field.Tag.DbStorage != "-" {
							//	orders = append(orders, strings.Join([]string{field.Tag.DbStorage, orderDirection}, " "))
						} else {
							err = _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' field '%s' has not have 'db.column_name'", entity.Name, fieldName))
							return "", nil, err
						}
					}

				}

				if len(orders) > 0 {
					sqlOrder = strings.Join([]string{"ORDER BY", strings.Join(orders, ", ")}, " ")
				}
			}
		} // ORDER BY clause

		if limitOption != "" {
			sqlLimit = strings.Join([]string{"LIMIT", limitOption}, " ")
		}

		if offsetOption != "" {
			sqlOffset = strings.Join([]string{"OFFSET", offsetOption}, " ")
		}

		sqlFull = strings.Join([]string{sqlFrom, sqlWhere, sqlOrder, sqlOffset, sqlLimit}, " ")

		_log.Debug("SUCCESS: requestID, entityName, sqlFull, args, duration", requestID, entity.Name, sqlFull, sqlArgs, time.Now().Sub(tic))

		return sqlFull, sqlArgs, err
	}
	return "", nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && && rowOut != nil {}", []interface{}{s, rowOut}).PrintfError()
}

// constructSqlInsert сформировать sql для создания строки
func (s *Storage) constructSqlInsert(requestID uint64, returning bool, rowIn *_meta.Object, rowOut *_meta.Object) (sqlFull string, sqlArgs []interface{}, err error) {
	if s != nil && rowIn != nil {

		tic := time.Now()
		entity := rowIn.Entity
		inFields := rowIn.Fields
		argCnt := 1
		sqlVal := ""
		sqlTable := ""
		sqlInto := ""
		sqlReturning := ""
		columnsOut := make([]string, 0, len(entity.StructFields()))
		columnsInto := make([]string, 0, len(entity.StructFields()))
		columnsVal := make([]string, 0, len(entity.StructFields()))
		restrictIn := inFields != nil && len(inFields) > 0
		restrictOut := rowOut != nil && rowOut.Fields != nil && len(rowOut.Fields) > 0
		parameterChar := "$"

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		switch s.dbCfg.DriverName {
		case "postgres", "pgx":
			parameterChar = "$"
		case "godror", "oracle":
			parameterChar = ":"
		default:
			parameterChar = "$"
		}

		if entity.DbStorage.TableName == "" {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', does not defined 'db.table_name'", entity.Name))
		}

		if entity.DbStorage.SchemaName != "" {
			sqlTable = entity.DbStorage.SchemaName + "." + entity.DbStorage.TableName
		} else {
			sqlTable = entity.DbStorage.TableName
		}

		for _, field := range entity.StructFields() {
			if field.Tag.Db != "" && field.Tag.Db != "-" && field.DbStorage.ColumnName != "" {

				if returning && rowOut != nil {
					// Ограничить поля возврата определенным списком - вернуть можно больше полей, чем было помещено в БД
					if restrictOut {
						if _, ok := rowOut.Fields[field.Name]; ok {
							columnsOut = append(columnsOut, field.DbStorage.ColumnName+" as "+field.Tag.Db)
						}
					} else {
						columnsOut = append(columnsOut, field.DbStorage.ColumnName+" as "+field.Tag.Db)
					}
				}

				// Пробросить поля, действие с которыми запрещено
				if field.Modify.CreateRestrict {
					continue
				}

				// Ограничить помещаемые поля определенным списком
				if restrictIn {
					if _, ok := inFields[field.Name]; !ok {
						continue
					}
				}

				// Аргумент для обработки
				if rv, err := rowIn.FieldRV(field); err != nil {
					return "", nil, err
				} else {
					sqlArgs = append(sqlArgs, rv.Interface())
				}

				columnsInto = append(columnsInto, field.DbStorage.ColumnName)
				columnsVal = append(columnsVal, parameterChar+strconv.Itoa(argCnt))
				argCnt++
			}
		}

		if len(columnsOut) > 0 {
			sqlReturning = "RETURNING " + strings.Join(columnsOut, ", ")
		}

		if len(columnsInto) > 0 {
			sqlInto = "INSERT INTO " + sqlTable + " (" + strings.Join(columnsInto, ", ") + ")"
		} else {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty INTO column list for persist", entity.Name))
		}

		if len(columnsVal) > 0 {
			sqlVal = "VALUES (" + strings.Join(columnsVal, ", ") + ")"
		} else {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty VALUES column list for persist", entity.Name))
		}

		sqlFull = strings.Join([]string{sqlInto, sqlVal, sqlReturning}, " ")

		_log.Debug("SUCCESS: requestID, entityName, duration, sqlFull", requestID, entity.Name, time.Now().Sub(tic), sqlFull)

		return sqlFull, sqlArgs, err
	}
	return "", nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil {}", []interface{}{s, rowIn}).PrintfError()
}

// constructSqlInsert сформировать sql для создания строки
func (s *Storage) constructSqlUpdate(requestID uint64, returning bool, rowIn *_meta.Object, rowOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (sqlFull string, sqlArgs []interface{}, err error) {
	if s != nil && rowIn != nil {

		tic := time.Now()
		entity := rowIn.Entity
		inFields := rowIn.Fields
		argCnt := 1
		sqlTable := ""
		sqlSet := ""
		sqlWhere := ""
		sqlReturning := ""
		columnsOut := make([]string, 0, len(entity.StructFields()))
		columnsSet := make([]string, 0, len(entity.StructFields()))
		restrictIn := inFields != nil && len(inFields) > 0
		restrictOut := rowOut != nil && rowOut.Fields != nil && len(rowOut.Fields) > 0
		parameterChar := "$"

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		switch s.dbCfg.DriverName {
		case "postgres", "pgx":
			parameterChar = "$"
		case "godror", "oracle":
			parameterChar = ":"
		default:
			parameterChar = "$"
		}

		if key == nil {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - Key was not defined", entity.Name))
		}

		if len(keyArgs) == 0 {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty keyArgs", entity.Name))
		}

		if entity.DbStorage.TableName == "" {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', does not defined 'db.table_name'", entity.Name))
		}

		if entity.DbStorage.SchemaName != "" {
			sqlTable = entity.DbStorage.SchemaName + "." + entity.DbStorage.TableName
		} else {
			sqlTable = entity.DbStorage.TableName
		}

		for _, field := range entity.StructFields() {
			if field.Tag.Db != "" && field.Tag.Db != "-" && field.DbStorage.ColumnName != "" {

				if returning && rowOut != nil {
					// Ограничить поля возврата определенным списком - вернуть можно больше полей, чем было помещено в БД
					if restrictOut {
						if _, ok := rowOut.Fields[field.Name]; ok {
							columnsOut = append(columnsOut, field.DbStorage.ColumnName+" as "+field.Tag.Db)
						}
					} else {
						columnsOut = append(columnsOut, field.DbStorage.ColumnName+" as "+field.Tag.Db)
					}
				}

				// Пробросить поля, действие с которыми запрещено
				if field.Modify.UpdateRestrict {
					continue
				}

				// Ограничить помещаемые поля определенным списком
				if restrictIn {
					if _, ok := inFields[field.Name]; !ok {
						continue
					}
				}

				// Аргумент для обработки
				if rv, err := rowIn.FieldRV(field); err != nil {
					return "", nil, err
				} else {
					sqlArgs = append(sqlArgs, rv.Interface())
				}

				columnsSet = append(columnsSet, field.DbStorage.ColumnName+"="+parameterChar+strconv.Itoa(argCnt))
				argCnt++
			}
		}

		if len(columnsOut) > 0 {
			sqlReturning = "RETURNING " + strings.Join(columnsOut, ", ")
		}

		if len(columnsSet) > 0 {
			sqlSet = "UPDATE " + sqlTable + " SET " + strings.Join(columnsSet, ", ")
		} else {
			return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty SET column list for persist", entity.Name))
		}

		// Добавим условие поиска по ключу
		if wheresKey, wheresKeyArgs, err := s.constructWhereByKey(requestID, rowIn, parameterChar, argCnt, key, keyArgs...); err != nil {
			return "", nil, err
		} else {
			sqlArgs = append(sqlArgs, wheresKeyArgs...)

			if len(wheresKey) > 0 {
				sqlWhere = "WHERE " + strings.Join(wheresKey, " and")
			} else {
				return "", nil, _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - empty wheresKey", entity.Name))
			}
		}

		sqlFull = strings.Join([]string{sqlSet, sqlWhere, sqlReturning}, " ")

		_log.Debug("SUCCESS: requestID, entityName, duration, sqlFull", requestID, entity.Name, time.Now().Sub(tic), sqlFull)

		return sqlFull, sqlArgs, err
	}
	return "", nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowIn != nil && key != nil {}", []interface{}{s, rowIn, key}).PrintfError()
}
