package apiservice

import (
	"context"
	"fmt"
	"reflect"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// processPrePersist - в ходе разбора сообщения marshaling
func (s *Service) processPrePersist(ctx context.Context, requestID uint64, row *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (err error, errors _err.Errors) {
	if s != nil && row != nil {

		//tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, row.EntityName", requestID, row.Entity.Name)

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		switch action {
		case PERSIST_ACTION_CREATE, PERSIST_ACTION_UPDATE:
		default:
			return _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - 'processPrePersist' incorrect action '%s'", row.Entity.Name, action)), errors
		}

		// Обрабатываем только структуры
		if row.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if row.IsSlice {}", []interface{}{s, row}).PrintfError(), errors
		} else {

			// Вычисляемые поля
			if opt.calculate {
				if !opt.isAssociation { // для association не выполняем операций create, update - вычисления не требуются
					if !row.Options.Global.SkipCalculation {
						if action == PERSIST_ACTION_CREATE {
							err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_PRE_CREATE)
						} else if action == PERSIST_ACTION_UPDATE {
							err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_PRE_UPDATE)
						}
						if err != nil {
							errors.Append(requestID, err)
						}
					}
				}
			}

			// Обработать FK M:1 - при отрицательном cascadeUp - обрабатываем все уровни без ограничений
			if row.Entity.HasAssociations() {
				if cascadeUp != 0 {
					//  уровень уменьшаем на 1, 0 - означает только себя
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					// У Association не должно быть своих Composition
					if err, localErrors = s.processAssociations(ctx, requestID, nil, row, action, exprAction, cascadeUp-1, 0, opt, s.prePersistAssociation); err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)
				}
			}

			// Вычисляемые поля, которые зависят от ранее вычисленных Association
			if opt.calculate {
				if !row.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_INSIDE_MARSHAL); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			{ // В режиме UPDATE, если row существует по UK - взять его PK
				if opt.persistUseUK || opt.persistUpdateAllFields {
					if !opt.isAssociation { // для association не выполняем операций create, update
						if !(errors.HasError() || innerErrors.HasError()) {
							localErrors := _err.Errors{} // локальные ошибки вложенного метода
							if err, localErrors = s.setFromKey(ctx, requestID, row, opt); err != nil {
								errors.Append(requestID, err)
							}
							innerErrors.AppendErrors(localErrors)
						}
					}
				}
			} // В режиме UPDATE, если row существует по UK - взять его PK

			// Обработать FK 1:M - при отрицательном cascadeDown - обрабатываем все уровни без ограничений
			if row.Entity.HasCompositions() {
				if cascadeDown != 0 {
					//  уровень уменьшаем на 1, 0 - означает только себя
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if err, localErrors = s.processCompositions(ctx, requestID, nil, row, action, exprAction, cascadeUp, cascadeDown-1, opt, s.prePersistComposition); err != nil {
						errors.Append(requestID, err)
					}
					innerErrors.AppendErrors(localErrors)
				}
			}

			// Вычисляемые поля
			if opt.calculate {
				if !row.Options.Global.SkipCalculation {
					if err = s.processExprs(ctx, requestID, nil, row, _meta.EXPR_ACTION_POST_MARSHAL); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			{ // В режиме UPDATE, если row существует по UK - взять его PK
				if opt.persistUseUK || opt.persistUpdateAllFields {
					if !opt.isAssociation { // для association не выполняем операций create, update
						if !(errors.HasError() || innerErrors.HasError()) {
							localErrors := _err.Errors{} // локальные ошибки вложенного метода
							if err, localErrors = s.setFromKey(ctx, requestID, row, opt); err != nil {
								errors.Append(requestID, err)
							}
							innerErrors.AppendErrors(localErrors)
						}
					}
				}
			} // В режиме UPDATE, если row существует по UK - взять его PK

			// Валидация данных
			if opt.validate {
				if row.Options.Global.Validate && s.validator != nil {
					if err = s.validator.ValidateObject(requestID, row); err != nil {
						errors.Append(requestID, err)
					}
				}
			}

			//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, row.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, row, errors, row.Options.Global.EmbedError, "Post Marshal")
			return err, errors
		}
	}

	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError(), errors
}

func (s *Service) prePersistAssociation(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, associationField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (associationRow *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && associationField != nil {

		//tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		//reference := associationField.Reference()
		//toEntity := reference.ToEntity()
		//toKey := reference.ToKey()

		//_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				associationRow = nil
				keyArgs = nil
				err = _recover.GetRecoverError(r, requestID, "prePersistAssociation", rowOut.Entity.Name)
			}
		}()

		// Найдем объект, хранимый в поле
		associationRow = rowOut.GetAssociationUnsafe(associationField)
		if associationRow != nil {

			{ // Рекурсивная обработка каскадно
				localErrors := _err.Errors{} // локальные ошибки вложенного метода
				if err, localErrors = s.processPrePersist(ctx, requestID, associationRow, action, exprAction, cascadeUp, cascadeDown, opt); err != nil {
					errors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
				}
				innerErrors.AppendErrors(localErrors)
			} // Рекурсивная обработка каскадно

			{ // Если association существует по PK UK - взять ее, иначе ошибка
				existsOnAnyKey := false // Найдено хоть по одному ключу
				allKeysEmpty := true    // Все ключи были пустыми

				for _, key := range associationRow.Entity.KeysUK() {

					localOptions := associationRow.Options.Clone()
					localOptions.Fields = nil // Извлекать все поля - они нужны для формул
					LocalOptionsGlobal := associationRow.Options.Global.Clone()
					LocalOptionsGlobal.Validate = false // Валидацию данных не выполнять
					localOptions.Global = LocalOptionsGlobal

					exists, rowUK, ukKeyArgs, err, localErrors := s.GetSingleByKeyUnsafe(ctx, requestID, associationRow.Entity, localOptions, associationRow, key)
					innerErrors.AppendErrors(localErrors)
					if err != nil {
						errors.Append(requestID, err)
					} else {
						if exists {
							associationRow.Options = localOptions
							associationRow.SetFromRV(rowUK.RV) // возьмем найденную строку
							existsOnAnyKey = true              // Найдено хоть по одному ключу
							break                              // Не проверяем, что по разным PK UK, получены разные строки
						} else {
							// Данных не найдено при заполненном ключе
							if !_meta.ArgsAllEmpty(ukKeyArgs) {
								allKeysEmpty = false
								errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_NO_DATA_FOUND, requestID, associationRow.Entity.Name, key.Name, key.FieldsString(), _meta.ArgsToString("','", ukKeyArgs...)))
							}
						}
					}
				}

				// Не найдено association и все ключи пустые - удаляем
				if !existsOnAnyKey && allKeysEmpty {
					return nil, nil, nil, errors
				}

			} // Если association существует по PK UK - взять ее, иначе ошибка

			//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
			err = s.processErrors(requestID, associationRow, errors, rowOut.Options.Global.EmbedError, "Association Post Marshal")
			return associationRow, keyArgs, err, errors

		} else {
			// Данных нет, обрабатывать не чего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil  && associationField != nil {}", []interface{}{s, rowOut, associationField}).PrintfError(), errors
}

func (s *Service) prePersistComposition(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (compositionRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && compositionField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := compositionField.Reference()
		//toEntity := reference.ToEntity()
		//toKey := reference.ToKey()

		//_log.Debug("START: requestID, entityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				compositionRows = nil
				keyArgs = nil
				err = _recover.GetRecoverError(r, requestID, "prePersistComposition", rowOut.Entity.Name)
			}
		}()

		// Найдем объект, хранимый в поле
		compositionRows = rowOut.GetCompositionUnsafe(compositionField)
		if compositionRows != nil {
			for _, compositionRow := range compositionRows.Objects {

				if compositionRow == nil {
					continue
				}

				rowErrors := _err.Errors{}

				// Вычисляемые поля перед рекурсивной обработкой, чтобы они попали в подчиненные объекты
				if opt.calculate {
					if !compositionRow.Options.Global.SkipCalculation {
						if toReference := reference.ToReference(); toReference != nil {
							// Реверсивная reference должна быть типа Association
							if toReference.Type == _meta.REFERENCE_TYPE_ASSOCIATION {
								// toReference.field - поле зеркального данному Composition-Association
								if associationField := toReference.Field(); associationField != nil {
									// найдем значение поля, в которое поместить структуру
									associationFieldRV, err := compositionRow.FieldRV(associationField)
									if err != nil {
										// Целевого поля может не быть - это нормальная ситуация
									} else {
										// Поместим в себя реверсивную ссылку на родителя, для отработки всех вычислений, позже удалим ее
										associationFieldRV.Set(rowOut.RV) // Встроим ссылку на родителя

										if err = s.processExprs(ctx, requestID, nil, compositionRow, _meta.EXPR_ACTION_INSIDE_MARSHAL); err != nil {
											rowErrors.Append(requestID, err)
										}

										associationFieldRV.Set(reflect.Zero(associationFieldRV.Type())) // Очищаем поле
									}
								}
							}
						}
					}
				} // Вычисляемые поля перед рекурсивной обработкой, чтобы они попали в подчиненные объекты

				{ // Рекурсивная обработка каскадно
					localErrors := _err.Errors{} // локальные ошибки вложенного метода
					if err, localErrors = s.processPrePersist(ctx, requestID, compositionRow, action, exprAction, cascadeUp, cascadeDown, opt); err != nil {
						rowErrors.Append(requestID, err) // Не встраиваем ошибки - накапливаем
					}
					rowErrors.AppendErrors(localErrors)
				} // Рекурсивная обработка каскадно

				//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
				err = s.processErrors(requestID, compositionRow, rowErrors, rowOut.Options.Global.EmbedError, "Composition Post Marshal")
				errors.Append(requestID, err)
				innerErrors.AppendErrors(rowErrors)
			}

			if errors.HasError() {
				_log.Debug("ERROR: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - MarshalEntity error", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name)), errors
			} else {
				//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
				return compositionRows, keyArgs, nil, errors
			}

		} else {
			// Данных нет, обрабатывать нечего
			return nil, nil, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && compositionField != nil {}", []interface{}{s, rowOut, compositionField}).PrintfError(), errors
}

// setFromKey - Если row существует то взять его поля для полного Update
func (s *Service) setFromKey(ctx context.Context, requestID uint64, row *_meta.Object, opt *processOptions) (err error, errors _err.Errors) {
	if s != nil && row != nil {

		innerErrors := _err.Errors{} // Ошибки вложенных методов

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		var keyCnt int
		var keys _meta.Keys

		localOptions := row.Options.Clone()
		LocalOptionsGlobal := row.Options.Global.Clone()
		LocalOptionsGlobal.Validate = false // Валидацию данных не выполнять
		localOptions.Global = LocalOptionsGlobal

		if opt.persistUpdateAllFields {
			localOptions.Fields = nil  // Нужны все поля
			localOptions.CascadeUp = 1 // Запросить со всеми ассоциациями
		}

		if !opt.persistUseUK {
			if pkKey := row.PKKey(); pkKey != nil {
				// Работаем только по PK
				keys = append(keys, pkKey)
			} else {
				errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' - has not have PK", row.Entity.Name)))
			}
		} else {
			// Работаем по всем PK, UK ключам
			keys = append(keys, row.Entity.KeysUK()...)
		}

		// Ищем по всем ключам PK - идет первым
		for _, key1 := range keys {
			keyCnt++
			exists, rowUK, _, err, localErrors := s.GetSingleByKeyUnsafe(ctx, requestID, row.Entity, localOptions, row, key1)
			innerErrors.AppendErrors(localErrors)
			if err != nil {
				errors.Append(requestID, err)
			} else {
				if exists {

					// Из найденной строки возьмем только PK объекта и все остальные, незаполненные UK, кроме себя
					for _, key2 := range row.Entity.KeysUK() {
						if key2 != key1 {
							// Проверим, что ключ не заполнен
							if key2Val, err := row.KeyFieldsValue(key2); err != nil {
								errors.Append(requestID, err)
							} else {
								if _meta.ArgsAllEmpty(key2Val) {
									// Пустой ключ переписываем
									if rowUKKeyRV, err := rowUK.KeyFieldsRV(key2); err != nil {
										errors.Append(requestID, err)
									} else {
										// Скопируем к себе
										if err = row.SetKeyFieldsRV(key2, rowUKKeyRV); err != nil {
											errors.Append(requestID, err)
										}
									}
								}
							}
						}
					}

					// TODO - в итоговый объект она не встроилась
					if opt.persistUpdateAllFields {
						// Из найденной строки возьмем все и поверх запишем свои данные - только те поля, что обновляем
						err = row.Entity.CopyObjectStruct(row, rowUK, row.Fields)
						if err != nil {
							errors.Append(requestID, err)
						}
						// TODO - Копировать ли Association?
						//row.SetFrom(rowUK, false, false) // возьмем найденную строку
						row.SetFrom(rowUK, true, false) // возьмем найденную строку
					}

					break // достаточно, если нашли объект по одному ключу
				}
			}
		}

		// TODO - не нужно проверять, ошибки, так как UK может появиться позже
		//if keyCnt > 0 && !exists {
		//	//if err = row.ZeroFieldsRV(pkKey.Fields()); err != nil {
		//	//	errors.Append(requestID, err)
		//	//}
		//	errors.Append(requestID, _err.NewTypedTraceEmpty(_err.ERR_API_NO_DATA_FOUND_ON_KEY, requestID, row.Entity.Name, row.KeysValueString()))
		//}

		err = s.processErrors(requestID, row, errors, row.Options.Global.EmbedError, "setFromKey")
		return err, errors
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError(), errors
}
