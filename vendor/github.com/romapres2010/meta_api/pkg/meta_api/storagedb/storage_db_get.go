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

// Get извлечь данные из БД
func (s *Storage) Get(ctx context.Context, requestID uint64, txId uint64, rowOut *_meta.Object, key *_meta.Key, keyArgs ...interface{}) (exists bool, err error) {
	if s != nil && s.db != nil && rowOut != nil && rowOut.Value != nil {

		_log.Debug("START: requestID, entityName, keyArgs", requestID, rowOut.Entity.Name, keyArgs)

		tic := time.Now()

		// На вход получаем только указатели на struct
		if rowOut.RV.Kind() != reflect.Ptr || reflect.Indirect(rowOut.RV).Kind() != reflect.Struct {
			return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "rowOut must be pointer to struct", []interface{}{rowOut.RV.Kind().String(), reflect.Indirect(rowOut.RV).Kind()}).PrintfError()
		}

		if len(keyArgs) == 0 {
			return false, nil
		}

		// Сформировать SQL
		sqlFull, sqlArgs, err := s.constructSqlSelect(requestID, rowOut, key, keyArgs...)
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
			exists, err = s.db.Get(requestID, tx, sqlFull, rowOut.Value, sqlArgs...)
			if err != nil {
				return false, err
			}

		} // Обработка в БД

		_log.Debug("SUCCESS: requestID, entityName, exists, duration", requestID, rowOut.Entity.Name, exists, time.Now().Sub(tic))

		return exists, nil
	}
	return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil && rowOut != nil && rowOut.Value != nil {}", []interface{}{s, rowOut}).PrintfError()
}
