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

type processCompositionFn func(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool) (compositionRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors)

// processCompositions - обработать все Composition обработчиком
func (s *Service) processCompositions(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, validate bool, calculate bool, processFn processCompositionFn) (err error, errors _err.Errors) {
	if s != nil && rowOut != nil && processFn != nil {

		innerErrors := _err.Errors{} // Ошибки вложенных методов
		entity := rowOut.Entity

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, requestID, "processCompositions", entity.Name)
			}
		}()

		tic := time.Now()

		// Обрабатываем только структуры
		if rowOut.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, rowOut}).PrintfError(), errors
		}

		// По всем виртуальным поля
		compositionMap := entity.CompositionMap()
		for _, compositionField := range compositionMap {

			if err = compositionField.CheckFieldReference(); err != nil {
				errors.Append(requestID, err)
				continue
			}

			reference := compositionField.Reference()
			toEntity := reference.ToEntity()
			toKey := reference.ToKey()

			// проверим, что поле в есть в списке ограничений
			if rowOut.Fields != nil && len(rowOut.Fields) > 0 {
				if _, ok := rowOut.Fields[compositionField.Name]; !ok {
					_log.Debug("Reference Composition - Skip out by outFields restriction: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())
					continue
				} else {
					_log.Debug("Reference Composition - PROCESS: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())
				}
			}

			// найдем значение поля, в которое поместить структуру
			compositionFieldRV, err := rowOut.FieldRV(compositionField)
			if err != nil {
				errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Composition field='%s' value by in index", entity.Name, compositionField.Name)))
				continue
			}

			if !(compositionFieldRV.IsValid()) {
				continue
			}

			// Обработаем Composition
			compositionRows, keyArgs, err, localErrors := processFn(ctx, requestID, rowIn, rowOut, compositionField, action, exprAction, cascadeUp, cascadeDown, validate, calculate)
			innerErrors.AppendErrors(localErrors)
			if err != nil {
				errors.Append(requestID, err)
			} else {

				if compositionRows != nil {

					// Если типы можно присваивать
					if reflect.TypeOf(compositionRows.Value).AssignableTo(compositionField.ReflectType()) {

						_log.Debug("Reference Composition - Entity FOUND: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

						// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
						if compositionRows.Options.Global.OutFormat == "xml" {
							xmlName := compositionField.GetXmlNameFromTag(true) // Считаем со ссылочного поля
							if err = compositionRows.SetXmlNameValueSlice(xmlName); err != nil {
								errors.Append(requestID, err) // накапливаем ошибки
							}
						}

						compositionFieldRV.Set(compositionRows.RV) // Встроим ссылку на slice в структуру

						// Сохранить объект как reference на поле
						if err = rowOut.SetCompositionUnsafe(compositionField, compositionRows); err != nil {
							errors.Append(requestID, err)
							continue
						}

						{ // Добавить себя в реверсивную ссылку Composition-Association
							if toReference := reference.ToReference(); toReference != nil {

								// Реверсивная reference должна быть типа Association
								if toReference.Type == _meta.REFERENCE_TYPE_ASSOCIATION {

									// toReference.field - поле зеркального данному Composition-Association
									if associationField := toReference.Field(); associationField != nil {

										var rowAssociation *_meta.Object

										// Создать копию объекта без тегов, чтобы исключить рекурсии при выводе в JSON / XML
										// Сформировать выходную структуру - все поля, но без tag
										if reference.Embed {
											// Если сущность встраивается, родителя можно не выводить полностью
											rowAssociation, err = s.newRowAllEmptyTag(requestID, entity, rowOut.Options)
										} else {
											// Если НЕ встраивается, то удаляем только ref
											rowAssociation, err = s.newRowAllEmptyRef(requestID, entity, rowOut.Options)
										}

										if err != nil {
											errors.Append(requestID, err)
										} else {
											_log.Debug("Deep CopyField Struct - rowAssociation: entityName", entity.Name)
											if err = entity.CopyObjectStruct(rowOut, rowAssociation, rowOut.Fields); err != nil {
												errors.Append(requestID, err)
											} else {
												// По всем строкам Composition проставим ссылку на Association
												for _, rowComposition := range compositionRows.Objects {

													// найдем значение поля, в которое поместить структуру
													associationFieldRV, err := rowComposition.FieldRV(associationField)
													if err != nil {
														errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Composition ='%s' - FromEntityName '%s', ToReference '%s' ERROR get Association field='%s' value by in index", entity.Name, reference.Name, toEntity.Name, toReference.Name, associationField.Name)))
														continue
													}

													associationFieldRV.Set(rowAssociation.RV) // Встроим ссылку на структуру в структуру

													// Сохранить копию текущей строки как Association на поле
													if err = rowComposition.SetAssociationUnsafe(associationField, rowAssociation); err != nil {
														errors.Append(requestID, err)
														continue
													}
												}
											}
										}
									} else {
										errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Composition ='%s' - FromEntityName '%s', ToReference '%s' empty field pointer", entity.Name, reference.Name, toEntity.Name, toReference.Name)))
									}
								} else {
									errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s', Composition ='%s' - FromEntityName '%s', ToReference '%s' incorrect reference type '%s'", entity.Name, reference.Name, toEntity.Name, toReference.Name, toReference.Type)))
								}
							}
						} // Добавить себя в реверсивную ссылку Composition-Association

					} else {
						errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' incompatible Struct type CompositionField='%s' CompositionFieldType='%s', compositionRowsType='%s'", entity.Name, compositionField.Name, compositionField.ReflectType().String(), reflect.TypeOf(compositionRows.Value).String())).PrintfError())
					}
				} else {
					_log.Debug("Reference Composition - Empty: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields, keyArgs", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString(), keyArgs)

					compositionFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_COMPOSITION_RT)) // Очищаем compositionFieldRV
					delete(rowOut.CompositionMap, reference)                              // Удаляем из структуры объектов

					// Ссылка обязательная - ошибка
					if reference.Required {
						errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Keys [%s] required Compostition '%s'['%s'] - empty or not found - reference entity '%s' key '%s'['%s']=['%s']", entity.Name, rowOut.KeysValueString(), compositionField.Name, reference.FieldsString(), toEntity.Name, toKey.Name, toKey.FieldsString(), _meta.ArgsToString("','", keyArgs...))))
					}
				}
			}
		}

		if errors.HasError() {
			_log.Debug("ERROR - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition - error", entity.Name, rowOut.KeysValueString())), errors
		} else {
			_log.Debug("SUCCESS - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return nil, errors
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && processFn != nil {", []interface{}{s, rowOut, processFn}).PrintfError(), errors
}

func (s *Service) clearCompositions(ctx context.Context, requestID uint64, row *_meta.Object) (err error) {
	if s != nil && row != nil {

		entity := row.Entity

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, requestID, "clearCompositions", entity.Name)
			}
		}()

		_log.Debug("START: requestID, entityName", requestID, entity.Name)

		tic := time.Now()
		errors := _err.Errors{}

		// Обрабатываем только структуры
		if row.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, row}).PrintfError()
		}

		// По всем виртуальным поля
		compositionMap := entity.CompositionMap()
		for _, compositionField := range compositionMap {

			if err = compositionField.CheckFieldReference(); err != nil {
				errors.Append(requestID, err)
				continue
			}

			// проверим, что поле в есть в списке ограничений
			if row.Fields != nil && len(row.Fields) > 0 {
				if _, ok := row.Fields[compositionField.Name]; !ok {
					continue
				}
			}

			// найдем значение поля
			compositionFieldRV, err := row.FieldRV(compositionField)
			if err != nil {
				errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Composition field='%s' value by in index", entity.Name, compositionField.Name)))
				continue
			}

			if compositionFieldRV.IsValid() {
				compositionFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_COMPOSITION_RT)) // Очищаем compositionFieldRV
				delete(row.CompositionMap, compositionField.Reference())              // Удаляем из структуры объектов
			}
		}

		if errors.HasError() {
			_log.Debug("ERROR - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition clear - error", entity.Name, row.KeysValueString()))
		} else {
			_log.Debug("SUCCESS - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return nil
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {}", []interface{}{s, row}).PrintfError()
}
