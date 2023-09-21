package httphandler

import (
    "context"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"

    _ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _http "github.com/romapres2010/meta_api/pkg/common/httpservice"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// ApiSelectHandler Сервис отвечает за извлечение из БД - несколько строк
func (s *Service) ApiSelectHandler(w http.ResponseWriter, r *http.Request) {
    _log.Debug("START   ==================================================================================")

    // Запускаем типовой Process, возврат ошибки игнорируем
    _ = s.httpService.Process(false, "GET", w, r, func(ctx context.Context, requestBuf []byte, buf []byte) ([]byte, _http.Header, int, error) {
        requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
        queryOptions := _http.UrlValuesToMap(r.URL.Query())
        vars := mux.Vars(r) // Считаем параметры из URL path
        entityName := vars["entity"]
        inFormat := vars["format"]

        _log.Debug("START: requestID", requestID)

        if s.apiService == nil {
            err := _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty Api service", []interface{}{s.apiService}).PrintfError()
            return nil, nil, http.StatusBadRequest, err
        }

        // вызываем сервис
        exists, responseBuf, outFormat, err, innerErrors := s.apiService.SelectMarshal(ctx, entityName, inFormat, queryOptions)

        // подготавливаем заголовок ответа
        // TODO - вынести общий код
        header := _http.Header{}
        header[_http.HEADER_CUSTOM_ERR_CODE] = _http.HEADER_CUSTOM_ERR_CODE_SUCCESS
        header[_http.HEADER_CUSTOM_REQUEST_ID] = strconv.FormatUint(requestID, 10)

        switch outFormat {
        case "json":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_JSON_UTF8
        case "xml":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_XML_UTF8
        case "yaml":
            header[_http.HEADER_CONTENT_TYPE] = _http.HEADER_CONTENT_TYPE_PLAIN_UTF8
        case "xls":
            header[_http.HEADER_CONTENT_TYPE] = "application/octet-stream"
            header[_http.HEADER_CONTENT_DISPOSITION] = "attachment; filename=" + entityName + ".xlsx"
            header[_http.HEADER_CONTENT_TRANSFER_ENCODING] = _http.HEADER_CONTENT_TRANSFER_ENCODING_BINARY
        default:
            return requestBuf, nil, http.StatusBadRequest, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "Allowed only outFormat='json', 'yaml', 'xml', 'xsl'", outFormat).PrintfError()
        }
        // TODO - вынести общий код

        //if err != nil {
        //	var httpStatus = http.StatusBadRequest
        //
        //	// Если в ответе есть и ошибка и данные, то выводим только данные, они уже должны включать в себя ошибки
        //	if responseBuf != nil {
        //		return responseBuf, header, httpStatus, nil
        //	} else {
        //		return responseBuf, nil, httpStatus, err
        //	}
        //}
        //
        //_log.Debug("SUCCESS: requestID", requestID)
        //return responseBuf, header, http.StatusOK, nil

        if err != nil || innerErrors.HasError() {
            if responseBuf != nil {
                return responseBuf, header, http.StatusNotFound, nil
            } else {
                return nil, nil, http.StatusBadRequest, err
            }
        } else {
            if exists {
                _log.Debug("SUCCESS: requestID", requestID)
                return responseBuf, header, http.StatusOK, nil
            } else {
                _log.Debug("SUCCESS - NOT FOUND requestID", requestID)
                return responseBuf, header, http.StatusNotFound, nil
            }
        }

    })

    _log.Debug("SUCCESS ==================================================================================")
}
