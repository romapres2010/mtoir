package apiservice

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// TODO - перейти на общие опции
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

		//_log.Debug("START: requestID, entityName, object.IsSlice", requestID, entity.Name, object.IsSlice)

		if object.IsSlice {

			rows := object // обработка массива - slice

			for _, rowOut := range rows.Objects {
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPost(ctx, requestID, rowOut, action, exprAction, validate, calculate, filter, checkExists, checkNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}

			// Обработать ошибку - встраивать в slice некуда
			if errors.HasError() {
				_log.Debug("ERROR: requestID, entityName, duration, errors", requestID, entity.Name, time.Now().Sub(tic), len(errors))
				if rows.Reference == nil {
					return errors.Error(requestID, fmt.Sprintf("Entity '%s' Slice - error 'Post process'", entity.Name)), errors
				} else {
					return errors.Error(requestID, fmt.Sprintf("Entity '%s', Referense '%s', Referense Entity '%s' - error 'Post process'", entity.Name, rows.Reference.Name, rows.Reference.ToEntity().Name)), errors
				}
			} else {
				//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
				return nil, errors
			}

		} else {

			row := object // обработка одного объекта - struct

			// Обработать Association
			for _, associationRow := range row.AssociationMap {

				localCheckExists := false
				localCheckNotExists := false

				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				// Для ассоциаций не проверяем дубли по ключу
				if err, localErrors = s.processPost(ctx, requestID, associationRow, action, exprAction, validate, calculate, filter, localCheckExists, localCheckNotExists); err != nil {
					errors.Append(requestID, err)
				}
				innerErrors.AppendErrors(localErrors)
			}

			// Обработать Composition
			for composition, compositionRows := range row.CompositionMap {

				localCheckExists := checkExists
				localCheckNotExists := checkNotExists

				if composition.Embed {
					localCheckExists = false
				}

				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPost(ctx, requestID, compositionRows, action, exprAction, validate, calculate, filter, localCheckExists, localCheckNotExists); err != nil {
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
							if compositionRowsOut, err := s.processCompositionFiltering(requestID, expr, row, compositionRows, composition); err != nil {
								errors.Append(requestID, err)
							} else {
								if compositionRowsOut != nil {
									// Переопределим текущую Composition после фильтрации
									row.CompositionMap[composition] = compositionRowsOut
								} else {
									// Пустой набор после фильтрации - удаляем Composition
									delete(row.CompositionMap, composition)
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
							if compositionRowsOut, err := s.processCompositionFiltering(requestID, expr, row, compositionRows, composition); err != nil {
								errors.Append(requestID, err)
							} else {
								if compositionRowsOut != nil {
									// Переопределим текущую Composition после фильтрации
									row.CompositionMap[composition] = compositionRowsOut
								} else {
									// Пустой набор после фильтрации - удаляем Composition
									delete(row.CompositionMap, composition)
								}
							}
						}
					}
				}
			}

			// Вычисляемые поля
			if calculate {
				if !row.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, row, exprAction); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			{ // Проверка ключей PK, UK - дубль по ключу
				//if false {
				if checkExists || checkNotExists {
					for _, key := range row.Entity.KeysUK() {
						exists, _, _, err, localErrors := s.GetSingleByKeyUnsafe(ctx, requestID, row.Entity, row.Options, row, key)
						innerErrors.AppendErrors(localErrors)
						if err != nil {
							errors.Append(requestID, err)
						} else {
							if exists {
								if checkNotExists {
									// Если нашли по одному ключу, дальше не проверяем
									// В списке ключей PK идет первым, если он будет найден, то до проверки UK дело не дойдет
									break
								}
								if checkExists && !row.Entity.Embed {
									// Найден дубль по ключу, кроме встраиваемых сущностей
									errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_DUPLICATE_VALUE_ON_KEY, requestID, row.Entity.Name, row.KeyValueString(key)))
								}
							} else {
								// В списке ключей PK идет первым, если он будет найден, то до проверки UK дело не дойдет
								if checkNotExists {
									errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_NO_DATA_FOUND_ON_KEY, requestID, row.Entity.Name, row.KeyValueString(key)))
								}
							}
						}
					}
				}
			} // Проверка ключей PK, UK - дубль по ключу

			// Валидация данных
			if validate {
				if row.Options.Global.Validate && s.validator != nil {
					if err = s.validator.ValidateObject(requestID, row); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			//_log.Debug("END: requestID, row.EntityName, duration, errors", requestID, row.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, row, errors, row.Options.Global.EmbedError, "Post process")
			return err, errors
		}
	} else {
		return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && object != nil && object.Options != nil {}", []interface{}{s, object}).PrintfError(), errors
	}
}
