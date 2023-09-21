package apiservice

import (
    "fmt"

    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
    _meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// processErrors встроить ошибку
func (s *Service) processErrors(requestID uint64, row *_meta.Object, errors _err.Errors, embedError bool, operationNme string) (err error) {
    if s != nil && row != nil {

        // Обрабатываем только структуры
        if row.IsSlice {
            return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if row.IsSlice {}", []interface{}{s, row}).PrintfError()
        }

        if len(errors) > 0 {
            _log.Debug("ERROR: requestID, entityName, errors", requestID, row.Entity.Name, len(errors))

            if !embedError {
                return errors.Error(requestID, fmt.Sprintf("Entity '%s', Keys [%s] - error '%s'", row.Entity.Name, row.KeysValueString(), operationNme)) // возвращаем обобщенную ошибку
            } else {
                // Встраиваем ошибки
                if err = row.SetErrorValue(errors); err != nil {
                    errors.Append(requestID, err)
                    return errors.Error(requestID, fmt.Sprintf("Entity '%s' - SetErrorValue error", row.Entity.Name))
                }
                return nil // возвращаем выходной объект со встроенной ошибкой
            }
        } else {
            _log.Debug("SUCCESS: requestID, entityName, duration", requestID, row.Entity.Name)
            return nil
        }
    }
    return errors.Error(requestID, fmt.Sprintf("Error '%s'", operationNme)) // возвращаем обобщенную ошибку
}
