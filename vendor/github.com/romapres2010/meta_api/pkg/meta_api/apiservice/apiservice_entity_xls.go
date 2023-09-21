package apiservice

import (
    "context"
    "fmt"
    "reflect"
    "time"

    "mime/multipart"

    _ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
    _meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// XlsMarshal извлечь данные и преобразовать в XLS
func (s *Service) XlsMarshal(ctx context.Context, entityName string, queryOptions _meta.QueryOptions, inFile multipart.File) (outBuf []byte, err error) {
    var requestID = _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context

    if s != nil && s.storageMap != nil && entityName != "" {

        var localCtx = contextWithOptionsCache(ctx) // Создадим новый контекст и встроим в него OptionsCache
        var tic = time.Now()
        var entity *_meta.Entity
        var options *_meta.Options
        var rowsOutPtr interface{}

        _log.Debug("START: requestID, entityName", requestID, entityName)

        // Meta может меняться в online - повесим запрет на чтение
        s.metaRLock()
        defer s.metaRUnLock()

        if entity = s.getEntityUnsafe(entityName); entity == nil {
            return nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)).PrintfError()
        }

        if options, err = s.parseQueryOptions(localCtx, requestID, entity.Name, entity, nil, queryOptions, nil); err != nil {
            return nil, err
        }

        if _, rowsOutPtr, err, _ = s.Select(localCtx, requestID, options.Entity, options, options.CascadeUp, options.CascadeDown, nil); err != nil {
            return nil, err
        }

        // сформируем ответ
        outBuf, err = s.MarshalXls(requestID, reflect.Indirect(reflect.ValueOf(rowsOutPtr)).Interface(), "Select:"+entityName, inFile)
        if err != nil {
            return nil, err
        }

        _log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))
        return outBuf, nil
    }
    return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.storageMap != nil && entityName != '' {}", []interface{}{s, entityName}).PrintfError()
}
