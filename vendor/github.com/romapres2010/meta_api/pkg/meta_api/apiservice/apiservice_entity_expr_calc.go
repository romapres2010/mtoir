package apiservice

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/antonmedv/expr/vm"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// processExprs выполнить расчет полей и сущности в целом
func (s *Service) processExprs(ctx context.Context, requestID uint64, rowIn *_meta.Object, row *_meta.Object, exprAction _meta.ExprAction) (err error) {
	if s != nil && row != nil {

		_log.Debug("START: requestID, entityName, exprAction", requestID, row.Entity.Name, exprAction)

		if err = s.processFieldsExprs(ctx, requestID, rowIn, row, exprAction); err != nil {
			return err
		}

		if err = s.processEntityExprs(ctx, requestID, rowIn, row, exprAction); err != nil {
			return err
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError()
}

// processExprs выполнить расчет полей
// TODO - добавить игнорирование ошибок и пустого результата
func (s *Service) processFieldsExprs(ctx context.Context, requestID uint64, rowIn *_meta.Object, row *_meta.Object, action _meta.ExprAction) (err error) {
	if s != nil && row != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName, action", requestID, row.Entity.Name, action)

		exprs := row.Entity.FieldsExprsByAction(action)
		if exprs != nil {

			// Функция восстановления после паники в reflect
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, requestID, "processExprs", row.Entity.Name)
				}
			}()

			restrict := len(row.Fields) > 0
			errors := _err.Errors{}
			exprVm := &vm.VM{} // https://expr.medv.io/docs/Tips

			// по всем выражениям, зарегистрированных на полях
			for _, ex := range *exprs {
				if ex.Field() != nil {

					if ex.Status != _meta.STATUS_ENABLED {
						continue
					}

					// Лишние поля не считаем
					if restrict {
						if _, ok := row.Fields[ex.Field().Name]; !ok {
							continue
						}
					}

					// Только выражения, которые применимы к типу запроса
					if !ex.CheckAction(action) {
						continue
					}

					if ex.Type == _meta.EXPR_CALCULATE {
						// Вычислим поля
						if output, err := ex.CalculateField(requestID, row, exprVm); err != nil {
							errors.Append(requestID, err)
							_log.Debug("Error calculate: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						} else {
							_log.Debug("Success calculate: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						}
					} else if ex.Type == _meta.EXPR_DB_CALCULATE {
						if output, err := s.calculateFieldDb(ctx, requestID, row, ex); err != nil {
							errors.Append(requestID, err)
							_log.Debug("Error DB calculate: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						} else {
							_log.Debug("Success DB calculate: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						}
					} else if ex.Type == _meta.EXPR_COPY && rowIn != nil {
						// Скопируем поля из входной структуры, если ее нет - то не считать ошибкой
						if output, err := ex.CopyField(requestID, rowIn, row, exprVm); err != nil {
							errors.Append(requestID, err)
							_log.Debug("Error copy: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						} else {
							_log.Debug("Success copy: entityName, exprName, action, output", row.Entity.Name, ex.Name, action, output)
						}
					}
				} else {
					return _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - empty field pointer", ex.Entity().Name, ex.Name, ex.Code)).PrintfError()
				}
			}

			if errors.HasError() {
				return errors.Error(requestID, fmt.Sprintf("Entity '%s' - calculation error", row.Entity.Name))
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, row.Entity.Name, time.Now().Sub(tic))
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError()
}

// processEntityExprs выполнить расчет сущностей
func (s *Service) processEntityExprs(ctx context.Context, requestID uint64, rowIn *_meta.Object, row *_meta.Object, action _meta.ExprAction) (err error) {
	if s != nil && row != nil {

		tic := time.Now()

		_log.Debug("START: requestID, entityName, action", requestID, row.Entity.Name, action)

		exprs := row.Entity.ExprsByAction(action)
		if exprs != nil {

			// Функция восстановления после паники в reflect
			defer func() {
				r := recover()
				if r != nil {
					err = _recover.GetRecoverError(r, requestID, "processEntityExprs", row.Entity.Name)
				}
			}()

			errors := _err.Errors{}

			for _, ex := range *exprs {
				if ex.Field() == nil {

					if ex.Status != _meta.STATUS_ENABLED {
						continue
					}

					// Только выражения, которые применимы к типу запроса
					if !ex.CheckAction(action) {
						continue
					}

					if ex.Type == _meta.EXPR_DB_CALCULATE {
						if err := s.calculateEntityDb(ctx, requestID, row, ex); err != nil {
							errors.Append(requestID, err)
						} else {
							_log.Debug("Success DB calculate: entityName, exprName, action", row.Entity.Name, ex.Name, action)
						}
					}
				} else {
					return _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - NOT empty field pointer", ex.Entity().Name, ex.Name, ex.Code)).PrintfError()
				}
			}

			if errors.HasError() {
				return errors.Error(requestID, fmt.Sprintf("Entity '%s' - calculation error", row.Entity.Name))
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, row.Entity.Name, time.Now().Sub(tic))
			}
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError()
}

func (s *Service) calculateFieldDb(ctx context.Context, externalId uint64, row *_meta.Object, ex *_meta.Expr) (output interface{}, err error) {
	if ex != nil && row != nil && row.RV.IsValid() {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, externalId, "calculateFieldDb", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", row.Entity.Name, ex.Name, ex.Code))
			}
		}()

		if !ex.IsInit() {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] is not init", ex.Entity().Name, ex.Field().Name, ex.Name, ex.Code)).PrintfError()
		}

		if ex.Type != _meta.EXPR_DB_CALCULATE {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.Entity().Name, ex.Field().Name, ex.Name, ex.Code, ex.Type, _meta.EXPR_DB_CALCULATE))
		}

		if ex.Field() == nil {
			return nil, _err.NewTyped(_err.ERR_ERROR, externalId, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - empty field pointer", ex.Entity().Name, ex.Name, ex.Code)).PrintfError()
		}

		// Получим БД для сущности
		storage, err := s.getStorageByEntity(ex.Entity())
		if err != nil {
			return nil, err
		}

		// Транзакцию получим через контекст
		txId := fromContextTxId(ctx)

		// Сформируем набор значений полей для обработки
		args, err := row.FieldsValue(ex.ArgsFields())
		if err != nil {
			return nil, err
		}

		// Поле, в которое поместить результат
		fieldRV, err := row.FieldRV(ex.Field())
		if err != nil {
			return nil, err
		}

		// Указатель на новый объект нужного типа, для scan результатов
		outputRVPtr := reflect.New(fieldRV.Type())
		outputRVPtrI := outputRVPtr.Interface()

		// Выполнить вычисление
		exists, err := storage.ExecuteScan(ctx, externalId, txId, []interface{}{outputRVPtrI}, ex.Code, args...)
		if err != nil {
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, externalId, err, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - error", ex.Entity().Name, ex.Field().Name, ex.Name, ex.Code))
		}

		if exists {
			// Установим поле по результатам расчета
			fieldRV.Set(outputRVPtr.Elem())
		} else {
			fieldRV.Set(reflect.Zero(fieldRV.Type())) // Очищаем поле
		}

		return outputRVPtr.Elem().Interface(), nil

	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil && row != nil && row.Value.IsValid() {}", []interface{}{ex, row}).PrintfError()
}

func (s *Service) calculateEntityDb(ctx context.Context, requestID uint64, row *_meta.Object, ex *_meta.Expr) (err error) {
	if ex != nil && row != nil {

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, requestID, "calculateFieldDb", fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - recover from panic", row.Entity.Name, ex.Name, ex.Code))
			}
		}()

		if !ex.IsInit() {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] is not init", ex.Entity().Name, ex.Field().Name, ex.Name, ex.Code)).PrintfError()
		}

		if ex.Type != _meta.EXPR_DB_CALCULATE {
			return _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' field '%s' Expression '%s' Code=['%s'] - incorrect 'type' '%s'. Must be '%s'", ex.Entity().Name, ex.Field().Name, ex.Name, ex.Code, ex.Type, _meta.EXPR_DB_CALCULATE))
		}

		// Получим БД для сущности
		storage, err := s.getStorageByEntity(ex.Entity())
		if err != nil {
			return err
		}

		// Транзакцию получим через контекст
		txId := fromContextTxId(ctx)

		// Сформируем набор значений полей для обработки
		args, err := row.FieldsValue(ex.ArgsFields())
		if err != nil {
			return err
		}

		// Сформируем набор полей для результатов
		var destFields _meta.Fields
		var destFieldsRV []reflect.Value
		var destFieldsOutRVPtr []reflect.Value
		var destFieldsOutPtrI []interface{}

		if ex.Field() != nil {
			destFields = append(destFields, ex.Field()) // Результат только одно поле
		} else {
			destFields = ex.DestFields() // Все результаты расчета
		}

		destFieldsRV, err = row.FieldsRV(destFields)
		if err != nil {
			return err
		}

		// Указатели на новый объект нужного типа, для scan результатов
		for _, fieldRV := range destFieldsRV {
			outputRVPtr := reflect.New(fieldRV.Type())
			destFieldsOutRVPtr = append(destFieldsOutRVPtr, outputRVPtr)
			destFieldsOutPtrI = append(destFieldsOutPtrI, outputRVPtr.Interface())
		}

		// Выполнить вычисление
		exists, err := storage.ExecuteScan(ctx, requestID, txId, destFieldsOutPtrI, ex.Code, args...)
		if err != nil {
			return _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - error", ex.Entity().Name, ex.Name, ex.Code))
		}

		if exists {
			for i, fieldRV := range destFieldsRV {
				fieldRV.Set(destFieldsOutRVPtr[i].Elem())
			}
		} else {
			for _, fieldRV := range destFieldsRV {
				fieldRV.Set(reflect.Zero(fieldRV.Type())) // Очищаем поле
			}
		}

		return nil

	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if ex != nil && row != nil {}", []interface{}{ex, row}).PrintfError()
}
