package httphandler

import (
    "context"
    "strconv"

    "net/http"

    "github.com/gorilla/mux"

    _ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _http "github.com/romapres2010/meta_api/pkg/common/httpservice"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// ApiUpdateEntityMetaHandler Сервис отвечает за обновление метаданные entity
func (s *Service) ApiUpdateEntityMetaHandler(w http.ResponseWriter, r *http.Request) {
    _log.Debug("START   ==================================================================================")

    // Запускаем типовой Process, возврат ошибки игнорируем
    _ = s.httpService.Process(false, "PUT", w, r, func(ctx context.Context, requestBuf []byte, buf []byte) ([]byte, _http.Header, int, error) {
        requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
        vars := mux.Vars(r)
        entityName := vars["entity"]
        format := vars["format"]

        _log.Debug("START: requestID", requestID)

        if s.apiService == nil {
            err := _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty Api service", []interface{}{s.apiService}).PrintfError()
            return nil, nil, http.StatusBadRequest, err
        }

        // вызываем сервис
        result, responseBuf, err := s.apiService.UpdateEntityMeta(ctx, requestBuf, entityName, format)
        if err != nil {
            if !result {
                return requestBuf, nil, http.StatusNotFound, err
            } else {
                return requestBuf, nil, http.StatusBadRequest, err
            }
        }

        // формируем ответ
        header := _http.Header{}
        header[_http.HEADER_CUSTOM_ERR_CODE] = _http.HEADER_CUSTOM_ERR_CODE_SUCCESS
        header[_http.HEADER_CUSTOM_REQUEST_ID] = strconv.FormatUint(requestID, 10)

        switch format {
        case "json":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_JSON_UTF8
        case "xml":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_XML_UTF8
        case "yaml":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_PLAIN_UTF8
        default:
            return requestBuf, nil, http.StatusBadRequest, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "Allowed only format='json', 'yaml', 'xml'", format).PrintfError()
        }

        _log.Debug("SUCCESS: requestID", requestID)
        return responseBuf, header, http.StatusOK, nil
    })

    _log.Debug("SUCCESS ==================================================================================")
}
