package storagedb

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_sql "github.com/romapres2010/meta_api/pkg/common/sqlxx"
)

// persist обновить / создать строку в БД
func (s *Storage) persist(ctx context.Context, requestID uint64, action Action, txId uint64, rowIn *_meta.Object, rowOut *_meta.Object) (err error) {
	if s != nil && s.db != nil && rowIn != nil && rowOut != nil {

		_log.Debug("START: requestID, entityName", requestID, rowIn.Entity.Name)

		tic := time.Now()
		returning := !(s.dbCfg.DriverName == "godror" || s.dbCfg.DriverName == "oracle") // Oracle не поддерживает returning
		sqlFull := ""
		var sqlArgs []interface{}
		var keyArgs []interface{}

		// Обрабатываем только структуры
		if rowIn.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowIn.IsSlice {}", []interface{}{s, rowIn}).PrintfError()
		}

		// Обрабатываем только структуры
		if rowOut.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, rowIn}).PrintfError()
		}

		pkKey := rowIn.PKKey()
		if pkKey == nil {
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - PK Key was not defined", rowIn.Entity.Name))
		}

		keyArgs, err = rowIn.KeyFieldsValue(pkKey)
		if err != nil {
			return err
		}

		// Проверить если все аргументы пустые - то сразу отказ
		if _meta.ArgsAllEmpty(keyArgs) {
			return _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Key '%s'['%s'] - error persist - got all empty values", rowIn.Entity.Name, pkKey.Name, pkKey.FieldsString()))
		}

		switch action {
		case PERSIST_ACTION_CREATE:
			sqlFull, sqlArgs, err = s.constructSqlInsert(requestID, returning, rowIn, rowOut)
			if err != nil {
				return err
			}
		case PERSIST_ACTION_UPDATE:
			sqlFull, sqlArgs, err = s.constructSqlUpdate(requestID, returning, rowIn, rowOut, pkKey, keyArgs...)
			if err != nil {
				return err
			}
		default:
			return _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Keys [%s] - incorrect action '%s'", rowOut.Entity.Name, rowOut.KeysValueString(), action))
		}

		{ // Обработка в БД
			exists := false
			rowsAffected := int64(0)

			// При необходимости разобрать SQL
			if err = s.db.PreparexAddSql(sqlFull, true); err != nil {
				return err
			}

			var tx *_sql.Tx
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

			if returning {
				// Обновить строку и считать назад данные
				exists, err = s.db.QueryRowxStructScan(requestID, tx, sqlFull, rowOut.Value, sqlArgs...)
				if err != nil {
					if txId == 0 { // работаем в контексте локальной транзакции
						_ = s.db.Rollback(requestID, tx)
					}
					return err
				}
			} else {
				// Oracle не поддерживает returning - отдельно запросим после обработки
				rowsAffected, _, err = s.db.Exec(requestID, tx, sqlFull, sqlArgs...)
				exists = rowsAffected > 0
				if err != nil {
					if txId == 0 { // работаем в контексте локальной транзакции
						_ = s.db.Rollback(requestID, tx)
					}
					return err
				}
			}

			// 0 строк обновлено - трактовать как ошибку
			if !exists {
				return _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Key ['%s'] - error '%s' - row does not exists or 0 row affected", rowIn.Entity.Name, rowIn.KeysValueString(), action))
			}

			// Oracle не поддерживает returning - отдельно запросим после обработки
			if !returning {
				exists, err = s.Get(ctx, requestID, txId, rowOut, pkKey, keyArgs...)
				if err != nil {
					return err
				}
				if !exists {
					return _err.NewTypedTraceEmpty(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Key ['%s'] - error query after '%s' - row does not exists", rowIn.Entity.Name, rowIn.KeysValueString(), action))
				}
			}

			if txId == 0 { // работаем в контексте локальной транзакции
				if err = s.db.Commit(requestID, tx); err != nil {
					return err
				}
			}
		} // Обработка в БД

		_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowIn.Entity.Name, time.Now().Sub(tic))

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && s.db != nil && rowIn != nil && rowOut != nil {}", []interface{}{s, rowIn, rowOut}).PrintfError()
}
