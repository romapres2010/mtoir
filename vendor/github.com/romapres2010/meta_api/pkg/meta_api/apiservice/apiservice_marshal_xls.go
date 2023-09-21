package apiservice

import (
	"bytes"
	"time"

	"github.com/xuri/excelize/v2"
	"mime/multipart"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
	_xlsx "github.com/romapres2010/meta_api/pkg/common/xlsx"
)

// MarshalXls трансформировать произвольную структуру в 'xls'
func (s *Service) MarshalXls(requestID uint64, val any, name string, inFile multipart.File) (buf []byte, myerr error) {
	_log.Debug("START: requestID, name", requestID, name)

	var err error
	var tic = time.Now()
	var xlsxFile *excelize.File
	var xlsx *_xlsx.Xlsx

	if val != nil {

		if inFile != nil {
			// Работаем на основе шаблона xls
			xlsxFile, err = excelize.OpenReader(inFile)
			if err != nil {
				return nil, _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, requestID, err, "MarshalXls").PrintfError()
			}
			if xlsxFile != nil {
				defer func(xlsxFile *excelize.File) {
					err = xlsxFile.Close()
					if err != nil {
						myerr = _err.WithCauseTyped(_err.ERR_COMMON_ERROR, requestID, err, "apiservice.MarshalXls -> excelize.File.Close()").PrintfError()
					}
				}(xlsxFile)
			}
			xlsx = _xlsx.NewXlsx(requestID, xlsxFile)
		} else {
			// Работаем без шаблона шаблона xls
			xlsx = _xlsx.NewXlsx(requestID, nil)
		}

		err = xlsx.WriteStruct(val, "", _xlsx.WriteOption{SetTitles: true, NewRow: true, GroupTitles: true, FloatPrecision: 6, CascadeStruct: true, Transpose: false})
		if err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, requestID, err, "xlsx.Write").PrintfError()
			return nil, myerr
		}

		var buffer *bytes.Buffer
		buffer, err = xlsx.WriteToBuffer()
		if err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, requestID, err, "xlsx.WriteToBuffer").PrintfError()
			return nil, myerr
		}

		err = xlsx.Close()
		if err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_XLSX_COMMON_ERROR, requestID, err, "xlsx.Close").PrintfError()
			return nil, myerr
		}
		buf = buffer.Bytes()
	}
	_metrics.IncMarshalingDurationVec("xls", "MarshalXls", name, time.Now().Sub(tic))

	_log.Debug("SUCCESS: requestID, name, duration", requestID, name, time.Now().Sub(tic))
	return buf, nil
}
