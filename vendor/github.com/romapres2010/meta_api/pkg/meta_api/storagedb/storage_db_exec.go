package storagedb

import (
	"context"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_sql "github.com/romapres2010/meta_api/pkg/common/sqlxx"
)

// Execute выполнить код в БД
func (s *Storage) Execute(ctx context.Context, requestID uint64, txId uint64, sqlQuery string, args ...interface{}) (err error) {
	if s != nil && s.db != nil {

		_log.Debug("START: requestID", requestID)

		tic := time.Now()
		var tx *_sql.Tx

		{ // Обработка в БД
			// При необходимости разобрать SQL
			if err = s.db.PreparexAddSql(sqlQuery, true); err != nil {
				return err
			}

			if txId == 0 {
				// работаем в контексте локальной транзакции - не помещаем ее в cache
				if tx, err = s.db.BeginTxx(ctx, requestID, nil); err != nil {
					return err
				}
			} else {
				// Транзакция должна существовать в cache
				if tx, err = s.GetTxFromCache(txId); err != nil {
					return err
				}
			}

			// Создать строку и считать назад данные
			if _, _, err = s.db.Exec(requestID, tx, sqlQuery, args...); err != nil {
				if txId == 0 { // работаем в контексте локальной транзакции
					_ = s.db.Rollback(requestID, tx)
				}
				return err
			}

			if txId == 0 { // работаем в контексте локальной транзакции
				if err = s.db.Commit(requestID, tx); err != nil {
					return err
				}
			}
		} // Обработка в БД

		_log.Debug("SUCCESS: requestID, duration", requestID, time.Now().Sub(tic))

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}

// ExecuteScan выполнить код в БД и вернуть результат
func (s *Storage) ExecuteScan(ctx context.Context, requestID uint64, txId uint64, dest []interface{}, sqlQuery string, args ...interface{}) (exists bool, err error) {
	if s != nil && s.db != nil {

		_log.Debug("START: requestID", requestID)

		tic := time.Now()
		var tx *_sql.Tx

		{ // Обработка в БД
			// При необходимости разобрать SQL
			if err = s.db.PreparexAddSql(sqlQuery, true); err != nil {
				return false, err
			}

			if txId == 0 {
				// работаем в контексте локальной транзакции - не помещаем ее в cache
				if tx, err = s.db.BeginTxx(ctx, requestID, nil); err != nil {
					return false, err
				}
			} else {
				// Транзакция должна существовать в cache
				if tx, err = s.GetTxFromCache(txId); err != nil {
					return false, err
				}
			}

			// Выполнить и считать назад данные
			if exists, err = s.db.QueryRowScan(requestID, tx, sqlQuery, dest, args...); err != nil {
				if txId == 0 { // работаем в контексте локальной транзакции
					_ = s.db.Rollback(requestID, tx)
				}
				return false, err
			}

			if txId == 0 { // работаем в контексте локальной транзакции
				if err = s.db.Commit(requestID, tx); err != nil {
					return false, err
				}
			}
		} // Обработка в БД

		_log.Debug("SUCCESS: requestID, duration", requestID, time.Now().Sub(tic))

		return exists, nil
	}
	return false, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil {}", []interface{}{s}).PrintfError()
}
