package apiservice

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// processPost - рекурсивная пост обработка, включая валидацию и вычисление полей
func (s *Service) processPost(ctx context.Context, requestID uint64, object *_meta.Object, action Action, exprAction _meta.ExprAction, validate bool, calculate bool, filter bool, checkExists bool, checkNotExists bool) (err error, errors _err.Errors) {
	if s != nil && object != nil && object.Options != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		entity := object.Entity

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		_log.Debug("START: requestID, entityName, object.IsSlice", requestID, entity.Name, object.IsSlice)

		if object.IsSlice {

			rowsOut := object // обработка массива - slice

			for _, rowOut := range rowsOut.Objects {
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPost(ctx, requestID, rowOut, action, exprAction, validate, calculate, filter, checkExists, checkNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}

			// Обработать ошибку - встраивать в slice некуда
			if errors.HasError() {
				_log.Debug("ERROR: requestID, entityName, duration, errors", requestID, entity.Name, time.Now().Sub(tic), len(errors))
				if rowsOut.Reference == nil {
					return errors.Error(requestID, fmt.Sprintf("Entity '%s' Slice - error 'Post process'", entity.Name)), errors
				} else {
					return errors.Error(requestID, fmt.Sprintf("Entity '%s', Referense '%s', Referense Entity '%s' - error 'Post process'", entity.Name, rowsOut.Reference.Name, rowsOut.Reference.ToEntity().Name)), errors
				}
			} else {
				_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
				return nil, errors
			}

		} else {

			rowOut := object // обработка одного объекта - struct

			// Обработать Association
			for _, associationRow := range rowOut.AssociationMap {

				associationCheckExists := false
				associationCheckNotExists := false

				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				// Для ассоциаций не проверяем дубли по ключу
				if err, localErrors = s.processPost(ctx, requestID, associationRow, action, exprAction, validate, calculate, filter, associationCheckExists, associationCheckNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}

			// Обработать Composition
			for composition, compositionRows := range rowOut.CompositionMap {

				compositionCheckExists := checkExists
				compositionCheckNotExists := checkNotExists

				if composition.Embed {
					compositionCheckExists = false
				}

				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPost(ctx, requestID, compositionRows, action, exprAction, validate, calculate, filter, compositionCheckExists, compositionCheckNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)

				// Фильтруем composition только Cardinality = M
				if filter && composition.Cardinality == _meta.REFERENCE_CARDINALITY_M {

					// Выполним статическую фильтрацию, встроенную в Meta
					if object.Options.Global.StaticFiltering {
						for _, expr := range composition.Exprs {

							// Только выражения, которые применимы к типу запроса
							if !expr.CheckAction(exprAction) {
								continue
							}

							// Фильтрация Composition
							if compositionRowsOut, err := s.processCompositionFiltering(requestID, expr, rowOut, compositionRows, composition); err != nil {
								errors.Append(requestID, err)
							} else {
								if compositionRowsOut != nil {
									// Переопределим текущую Composition после фильтрации
									rowOut.CompositionMap[composition] = compositionRowsOut
								} else {
									// Пустой набор после фильтрации - удаляем Composition
									delete(rowOut.CompositionMap, composition)
								}
							}
						}
					}

					// Если в опциях есть динамическая фильтрация, то отработаем ее
					if object.Options.FilterPostRefExprs != nil {

						// Если определено выражение для расчета composition - выполним его
						if expr, ok := object.Options.FilterPostRefExprs[composition]; ok {

							// Только выражения, которые применимы к типу запроса
							if !expr.CheckAction(exprAction) {
								continue
							}

							// Фильтрация Composition
							if compositionRowsOut, err := s.processCompositionFiltering(requestID, expr, rowOut, compositionRows, composition); err != nil {
								errors.Append(requestID, err)
							} else {
								if compositionRowsOut != nil {
									// Переопределим текущую Composition после фильтрации
									rowOut.CompositionMap[composition] = compositionRowsOut
								} else {
									// Пустой набор после фильтрации - удаляем Composition
									delete(rowOut.CompositionMap, composition)
								}
							}
						}
					}
				}
			}

			// Вычисляемые поля
			if calculate {
				if !rowOut.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, rowOut, exprAction); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			// Валидация данных
			if validate {
				if rowOut.Options.Global.Validate && s.validator != nil {
					if err = s.validator.ValidateObject(requestID, rowOut); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			// Проверка ключей PK, UK - дубль по ключу
			if checkExists || checkNotExists {
				for _, key := range rowOut.Entity.Keys() {
					if key != nil && (key.Type == _meta.KEY_TYPE_PK || key.Type == _meta.KEY_TYPE_UK) {
						// TODO - добавить локальный cache на время выполнения запроса -целостность чтения на уровне транзакции
						exists, _, _, err, localErrors := s.getByKeyUnsafe(ctx, requestID, rowOut.Entity, rowOut.Options, rowOut, key)
						innerErrors.AppendErrors(localErrors)
						if err != nil {
							errors.Append(requestID, err)
						} else {
							if exists {
								// Найден дубль по ключу
								if checkExists {
									errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_DUPLICATE_VALUE_ON_KEY, requestID, rowOut.Entity.Name, rowOut.KeyValueString(key)))
								}
							} else {
								// Не найден по ключу
								if checkNotExists {
									errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_NO_DATA_FOUND_ON_KEY, requestID, rowOut.Entity.Name, rowOut.KeyValueString(key)))
								}
							}
						}
					}
				}
			}

			_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, rowOut, errors, rowOut.Options.Global.EmbedError, "Post process")
			return err, errors
		}
	} else {
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && object != nil && object.Options != nil {}", []interface{}{s, object}).PrintfError(), errors
	}
}
