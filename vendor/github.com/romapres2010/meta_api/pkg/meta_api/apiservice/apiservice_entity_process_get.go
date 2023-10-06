package apiservice

import (
	"context"
	"fmt"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
)

// processGet - обработка одной строки при считывании данных, включая валидацию
func (s *Service) processGet(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (err error, errors _err.Errors) {
	if s != nil && rowOut != nil && rowOut.Entity != nil {

		//tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов

		//_log.Debug("START: requestID, EntityName", requestID, rowOut.Entity.Name)

		// Консолидируем все ошибки
		defer func() {
			if innerErrors.HasError() {
				errors.AppendErrors(innerErrors)
			}
		}()

		// Обрабатываем только структуры
		if rowOut.IsSlice {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if rowOut.IsSlice {}", []interface{}{s, rowOut}).PrintfError(), errors
		}

		// Обработать FK M:1 - при отрицательном cascadeUp - обрабатываем все уровни без ограничений
		if rowOut.Entity.HasAssociations() && cascadeUp != 0 {
			//if !rowOut.Entity.Embed { // Для встроенных сущностей не выводим ассоциации
			//  уровень уменьшаем на 1, 0 - означает только себя
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processAssociations(ctx, requestID, rowIn, rowOut, action, exprAction, cascadeUp-1, 0, opt, s.getAssociation); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
			//}
		}

		// Обработать FK 1:M - при отрицательном cascadeDown - обрабатываем все уровни без ограничений
		if rowOut.Entity.HasCompositions() && cascadeDown != 0 {
			//  уровень уменьшаем на 1, 0 - означает только себя
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if err, localErrors = s.processCompositions(ctx, requestID, rowIn, rowOut, action, exprAction, 0, cascadeDown-1, opt, s.getComposition); err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		// Вычисляемые поля
		if opt.calculate {
			if !rowOut.Options.Global.SkipCalculation {
				if err = s.processExprs(ctx, requestID, rowIn, rowOut, exprAction); err != nil {
					errors.Append(requestID, err)
				}
			}
		}

		// Валидация данных
		if opt.validate {
			if rowOut.Options.Global.Validate && s.validator != nil {
				if err = s.validator.ValidateObject(requestID, rowOut); err != nil {
					errors.Append(requestID, err)
				}
			}
		}

		//_log.Debug("END: requestID, rowOut.EntityName, duration, errors", requestID, rowOut.Entity.Name, time.Now().Sub(tic), len(errors))
		err = s.processErrors(requestID, rowOut, errors, rowOut.Options.Global.EmbedError, "Get")
		return err, errors
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && rowOut.Entity != nil {}", []interface{}{s, rowOut}).PrintfError(), errors
}

// getAssociation извлечь данные в struct - FK 1:M
func (s *Service) getAssociation(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, associationField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (associationRow *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && associationField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := associationField.Reference()
		toEntity := reference.ToEntity()
		toKey := reference.ToKey()

		//_log.Debug("START: requestID, rowOut.EntityName, reference.Name, toEntity.Name, toEntity.Name, toKey.Name, toKey.fields", requestID, rowOut.Entity.Name, reference.Name, toEntity.Name, toKey.Name, toKey.FieldsString())

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
				err = _recover.GetRecoverError(r, requestID, "getAssociation", rowOut.Entity.Name)
			}
		}()

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		keyArgs, err = rowOut.ReferenceFieldsValue(reference)
		if err != nil {
			return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Association '%s' - error get fields value", rowOut.Entity.Name, reference.Name)), errors
		}

		// По пустому ключу не ищем
		if _meta.ArgsAllEmpty(keyArgs) {
			return nil, nil, nil, errors
		}

		optionsRef, err := s.ParseQueryOptions(ctx, requestID, rowOut.Options.Key+".association."+toEntity.Name, toEntity, associationField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
		if err != nil {
			return nil, nil, err, errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			_, associationRow, err, localErrors = s.GetSingleUnsafe(ctx, requestID, optionsRef.Entity, optionsRef, cascadeUp, cascadeDown, toKey, keyArgs...)
			if err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		} // Найти в globalCache или считать из внешнего сервиса

		if errors.HasError() {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
			return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Association '%s' - error 'Get'", rowOut.Entity.Name, rowOut.KeysValueString(), associationField.Name)), errors
		} else {
			//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
			return associationRow, keyArgs, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && associationField != nil {}", []interface{}{s, rowOut, associationField}).PrintfError(), errors
}

// getComposition извлечь данные в struct - FK 1:M
func (s *Service) getComposition(ctx context.Context, requestID uint64, rowIn *_meta.Object, rowOut *_meta.Object, compositionField *_meta.Field, action Action, exprAction _meta.ExprAction, cascadeUp int, cascadeDown int, opt *processOptions) (compositionRows *_meta.Object, keyArgs []interface{}, err error, errors _err.Errors) {
	if s != nil && rowOut != nil && compositionField != nil {

		tic := time.Now()
		innerErrors := _err.Errors{} // Ошибки вложенных методов
		reference := compositionField.Reference()
		toEntity := reference.ToEntity()
		toKey := reference.ToKey()

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
				err = _recover.GetRecoverError(r, requestID, "getComposition", rowOut.Entity.Name)
			}
		}()

		// Сформируем список критериев для поиска по ключу - порядок полей должен совпадать
		keyArgs, err = rowOut.ReferenceFieldsValue(reference)
		if err != nil {
			return nil, nil, _err.WithCauseTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("Entity '%s', Composition '%s' - error get fields value", rowOut.Entity.Name, reference.Name)), errors
		}

		// По пустому ключу не ищем
		if _meta.ArgsAllEmpty(keyArgs) {
			return nil, nil, nil, errors
		}

		optionsRef, err := s.ParseQueryOptions(ctx, requestID, rowOut.Options.Key+".composition."+toEntity.Name, toEntity, compositionField, rowOut.Options.QueryOptionsDown, rowOut.Options.Global)
		if err != nil {
			return nil, nil, err, errors
		}

		{ // Найти в globalCache или считать из внешнего сервиса
			localErrors := _err.Errors{} // локальные ошибки вложенного метода
			if reference.Cardinality == _meta.REFERENCE_CARDINALITY_M {
				_, compositionRows, err, localErrors = s.selectUnsafe(ctx, requestID, optionsRef.Entity, optionsRef, cascadeUp, cascadeDown, toKey, keyArgs...)
			} else {
				_, compositionRows, err, localErrors = s.GetSingleUnsafe(ctx, requestID, optionsRef.Entity, optionsRef, cascadeUp, cascadeDown, toKey, keyArgs...)
			}
			if err != nil {
				errors.Append(requestID, err)
			}
			innerErrors.AppendErrors(localErrors)
		}

		if errors.HasError() {
			_log.Debug("ERROR: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
			return nil, nil, errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s], Composition '%s' - error 'Select'", rowOut.Entity.Name, rowOut.KeysValueString(), compositionField.Name)), errors
		} else {
			//_log.Debug("SUCCESS: requestID, entityName, duration", requestID, rowOut.Entity.Name, time.Now().Sub(tic))
			return compositionRows, keyArgs, nil, errors
		}
	}
	return nil, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if s != nil && rowOut != nil && compositionField != nil {}", []interface{}{s, rowOut, compositionField}).PrintfError(), errors
}
