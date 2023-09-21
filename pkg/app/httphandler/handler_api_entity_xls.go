package httphandler

import (
	"context"
	"strconv"

	"github.com/gorilla/mux"
	"mime/multipart"
	"net/http"

	_ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_http "github.com/romapres2010/meta_api/pkg/common/httpservice"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// ApiXlsHandler Сервис отвечает за извлечение из БД и формирование xls
func (s *Service) ApiXlsHandler(w http.ResponseWriter, r *http.Request) {
	_log.Debug("START   ==================================================================================")

	file, _, err := r.FormFile("file")
	if err != nil {
		err = _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err).PrintfError()
	}
	if file != nil {
		defer func(file multipart.File) {
			err = file.Close()
			if err != nil {
				_ = _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, "httpservice.ApiXlsHandler -> multipart.File.Close()").PrintfError()
			}
		}(file)
	}

	// Запускаем типовой Process, возврат ошибки игнорируем
	_ = s.httpService.Process(false, "POST", w, r, func(ctx context.Context, requestBuf []byte, buf []byte) ([]byte, _http.Header, int, error) {
		var requestID = _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
		var queryOptions = _http.UrlValuesToMap(r.URL.Query())
		var vars = mux.Vars(r) // Считаем параметры из URL path
		var entityName = vars["entity"]

		queryOptions["[entity_name]"] = entityName

		_log.Debug("START: requestID", requestID)

		if s.apiService == nil {
			err := _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty Api service", []interface{}{s.apiService}).PrintfError()
			return nil, nil, http.StatusBadRequest, err
		}

		// вызываем сервис
		responseBuf, err := s.apiService.XlsMarshal(ctx, entityName, queryOptions, file)
		if err != nil {
			return requestBuf, nil, http.StatusBadRequest, err
		}

		// формируем заголовок ответа
		header := _http.Header{}
		header[_http.HEADER_CUSTOM_ERR_CODE] = _http.HEADER_CUSTOM_ERR_CODE_SUCCESS
		header[_http.HEADER_CUSTOM_REQUEST_ID] = strconv.FormatUint(requestID, 10)
		header[_http.HEADER_CONTENT_TYPE] = "application/octet-stream"
		header[_http.HEADER_CONTENT_DISPOSITION] = "attachment; filename=" + entityName + ".xlsx"
		header[_http.HEADER_CONTENT_TRANSFER_ENCODING] = _http.HEADER_CONTENT_TRANSFER_ENCODING_BINARY

		_log.Debug("SUCCESS: requestID", requestID)
		return responseBuf, header, http.StatusOK, nil
	})

	_log.Debug("SUCCESS ==================================================================================")
}
