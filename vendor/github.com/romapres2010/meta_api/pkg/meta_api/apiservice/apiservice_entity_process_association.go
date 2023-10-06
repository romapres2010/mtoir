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

type processAssociationFn func(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, associationField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (associationRow *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors)

func (s *Service) processAssociations(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions, processFn processAssociationFn) (err error, errors _err.Errors) {
	if s != nil && rowOut != nil && processFn != nil {

		innerErrors := _err.Errors{} // Ошибки вложенных методов
		entity := rowOut.Entity

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

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
				err = _recover.GetRecoverError(r, requestID, "processAssociations", entity.Name)
			}
		}()

		tic := time.Now()

		// Обрабатываем только структуры
		if rowOut.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, rowOut}).PrintfError(), errors
		}

		// По всем виртуальным поля Association
		associationMap := entity.AssociationMap()
		for _, associationField := range associationMap {

			if err = associationField.CheckFieldReference(); err != nil {
				errors.Append(requestID, err)
				continue
			}

			reference := associationField.Reference()
			toEntity := reference.ToEntity()
			toKey := reference.ToKey()

			// проверим, что поле в есть в списке ограничений
			if rowOut.Fields != nil && len(rowOut.Fields) > 0 {
				if _, ok := rowOut.Fields[associationField.Name]; !ok {
					_log.Debug("Reference Association - Skip out by outFields restriction: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())
					continue
				} else {
					_log.Debug("Reference Association - PROCESS: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())
				}
			}

			// Найдем значение поля, которое содержит не разобранные данные, в него же поместим структуру
			associationFieldRV, err := rowOut.FieldRV(associationField)
			if err != nil {
				errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Association field='%s' value by in index", entity.Name, associationField.Name)))
				continue
			}

			if !(associationFieldRV.IsValid()) {
				continue
			}

			// Обработаем Association
			localOpt := opt.Clone()
			localOpt.isComposition = false
			localOpt.isAssociation = true
			associationRow, keyArgs, err, localErrors := processFn(ctx, requestID, rowIn, rowOut, associationField, action, exprAction, cascadeUp, cascadeDown, localOpt)
			innerErrors.AppendErrors(localErrors)
			if err != nil {
				errors.Append(requestID, err)
			} else {
				if associationRow != nil {

					// Если типы можно присваивать
					if reflect.TypeOf(associationRow.Value).AssignableTo(associationField.ReflectType()) {
						_log.Debug("Reference Association - Entity FOUND: requestID, rowOut.EntityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

						// Для корректного разбора XML нужно задать значение для поля XMLName https://pkg.go.dev/encoding/xml#Marshal
						if associationRow.Options.Global.OutFormat == "xml" {
							xmlName := associationField.GetXmlNameFromTag(true) // Считаем со ссылочного поля
							if err = associationRow.SetXmlNameValue(xmlName); err != nil {
								errors.Append(requestID, err) // накапливаем ошибки
							}
						}

						associationFieldRV.Set(associationRow.RV)

						// Сохранить объект как reference на поле
						if err = rowOut.SetAssociationUnsafe(associationField, associationRow); err != nil {
							errors.Append(requestID, err)
							continue
						}
					} else {
						errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s' incompatible Struct type AssociationField='%s' AssociationFieldType='%s', rowRefPtrType='%s'", entity.Name, associationField.Name, associationFieldRV.String(), associationRow.RV.String())).PrintfError())
					}
				} else {
					_log.Debug("Reference Association - Entity NOT FOUND: requestID, entityName, reference.Name, toEntity.Name, toKey.Name, toKey.fields, keyArgs", requestID, entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString(), keyArgs)

					associationFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_ASSOCIATION_RT)) // Очищаем associationFieldRV
					delete(rowOut.AssociationMap, reference)                              // Удаляем из структуры объектов

					// Ссылка обязательная - ошибка
					if reference.Required {
						errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Keys [%s] required Association '%s'['%s'] - empty or not found - reference entity '%s' key '%s'['%s']=['%s']", entity.Name, rowOut.KeysValueString(), associationField.Name, reference.FieldsString(), toEntity.Name, toKey.Name, toKey.FieldsString(), _meta.ArgsToString("','", keyArgs...))))
					} else {
						// Если все аргументы были пустыми, то нормальная ситуация для опционального ключа
						if !_meta.ArgsAllEmpty(keyArgs) {
							errors.Append(requestID, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity '%s', Keys [%s] Association '%s'['%s'] - empty or not found - reference entity '%s' key '%s'['%s']=['%s']", entity.Name, rowOut.KeysValueString(), associationField.Name, reference.FieldsString(), toEntity.Name, toKey.Name, toKey.FieldsString(), _meta.ArgsToString("','", keyArgs...))))
						}
					}
				}
			}
		}

		if errors.HasError() {
			_log.Debug("ERROR - Reference Association: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Associations - error", entity.Name, rowOut.KeysValueString())), errors
		} else {
			//_log.Debug("SUCCESS - Reference Association: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return nil, errors
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && processFn != nil {}", []interface{}{s, rowOut, processFn}).PrintfError(), errors
}

func (s *Service) clearAssociations(ctx context.Context, requestID uint64, row *_meta.Object) (err error) {
	if s != nil && row != nil {

		entity := row.Entity

		// Функция восстановления после паники в reflect
		defer func() {
			r := recover()
			if r != nil {
				err = _recover.GetRecoverError(r, requestID, "clearCompositions", entity.Name)
			}
		}()

		//_log.Debug("START: requestID, entityName", requestID, entity.Name)

		tic := time.Now()
		errors := _err.Errors{}

		// Обрабатываем только структуры
		if row.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, row}).PrintfError()
		}

		// По всем виртуальным поля Association
		associationMap := entity.AssociationMap()
		for _, associationField := range associationMap {

			if err = associationField.CheckFieldReference(); err != nil {
				errors.Append(requestID, err)
				continue
			}

			// проверим, что поле в есть в списке ограничений
			if row.Fields != nil && len(row.Fields) > 0 {
				if _, ok := row.Fields[associationField.Name]; !ok {
					continue
				}
			}

			// найдем значение поля
			associationFieldRV, err := row.FieldRV(associationField)
			if err != nil {
				errors.Append(requestID, _err.WithCauseTyped(_err.ERR_ERROR, requestID, err, fmt.Sprintf("Entity '%s' - ERROR get Association field='%s' value by in index", entity.Name, associationField.Name)))
				continue
			}

			if associationFieldRV.IsValid() {
				associationFieldRV.Set(reflect.Zero(_meta.FIELD_TYPE_ASSOCIATION_RT)) // Очищаем associationFieldRV
				delete(row.CompositionMap, associationField.Reference())              // Удаляем из структуры объектов
			}
		}

		if errors.HasError() {
			_log.Debug("ERROR - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Associations clear - error", entity.Name, row.KeysValueString()))
		} else {
			//_log.Debug("SUCCESS - Reference: requestID, entityName, duration", requestID, entity.Name, time.Now().Sub(tic))
			return nil
		}
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && row != nil {", []interface{}{s, row}).PrintfError()
}
