package storagedb

import (
	"context"
	"reflect"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_sql "github.com/romapres2010/meta_api/pkg/common/sqlxx"
)

// Select извлечь данные из БД
func (s *Storage) Select(ctx context.Context, requestID uint64, txId uint64, rowsOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (exists bool, err error) {
	if s != nil && s.db != nil && rowsOut != nil && rowsOut.Value != nil {

		_log.Debug("START: requestID, entityName", requestID, rowsOut.Entity.Name)

		tic := time.Now()

		// На вход получаем только указатели на slice
		if rowsOut.RV.Kind() != reflect.Ptr || reflect.Indirect(rowsOut.RV).Kind() != reflect.Slice {
			return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "rowsOut.Val must be pointer to slice", []interface{}{rowsOut.RV.Kind().String(), reflect.Indirect(rowsOut.RV).Kind()}).PrintfError()
		}

		// Сформировать SQL
		sqlFull, sqlArgs, err := s.constructSqlSelect(requestID, rowsOut, key, keyArgs...)
		if err != nil {
			return false, err
		}

		{ // Обработка в БД
			// При необходимости разобрать SQL
			if err = s.db.PreparexAddSql(sqlFull, true); err != nil {
				return false, err
			}

			var tx *_sql.Tx
			if txId == 0 {
				// работаем в контексте локальной транзакции - транзакцию не начинать
			} else {
				// Транзакция должна существовать в cache
				if tx, err = s.GetTxFromCache(txId); err != nil {
					return false, err
				}
			}

			// Запросить данные
			err = s.db.Select(requestID, tx, sqlFull, rowsOut.Value, sqlArgs...)
			if err != nil {
				return false, err
			}

		} // Обработка в БД

		rowCount := rowsOut.RV.Elem().Len()
		if rowCount > 0 {
			_log.Debug("SUCCESS: requestID, entityName, len, duration", requestID, rowsOut.Entity.Name, rowCount, time.Now().Sub(tic))
			return true, nil
		} else {
			_log.Debug("SUCCESS - NOT FOUND: requestID, entityName, len, duration", requestID, rowsOut.Entity.Name, rowCount, time.Now().Sub(tic))
			return false, nil
		}
	}
	return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil && rowsOut != nil && rowsOut.Value != nil {}", []interface{}{s, rowsOut}).PrintfError()
}
