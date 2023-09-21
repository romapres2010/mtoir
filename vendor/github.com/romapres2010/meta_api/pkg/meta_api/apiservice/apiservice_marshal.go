package apiservice

import (
	"bytes"
	"time"

	"encoding/xml"
	"gopkg.in/yaml.v3"

	"github.com/bytedance/sonic"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_metrics "github.com/romapres2010/meta_api/pkg/common/metrics"
	_xlsx "github.com/romapres2010/meta_api/pkg/common/xlsx"
)

// marshal трансформировать произвольную структуру в 'json', 'yaml', 'xml', 'xls'
func (s *Service) marshal(requestID uint64, val any, operation, name string, format string) (buf []byte, myerr error) {
	_log.Debug("START: requestID, name", requestID, name)

	var err error
	var tic = time.Now()

	switch format {
	case "json":
		//bytesBuf := bytes.Buffer{}
		//enc := json.NewEncoder(&bytesBuf)
		//enc.SetEscapeHTML(false)
		//if err = enc.Encode(val); err != nil {
		//	myerr = _err.WithCauseTyped(_err.ERR_JSON_MARSHAL_ERROR, requestID, err).PrintfError()
		//	return nil, myerr
		//}
		//buf = bytesBuf.Bytes()

		//if buf, err = json.Marshal(val); err != nil {
		sonicConfig := sonic.Config{
			DisallowUnknownFields: true,
		}.Froze()
		if buf, err = sonicConfig.Marshal(val); err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_JSON_MARSHAL_ERROR, requestID, err, val).PrintfError()
			return nil, myerr
		}
	case "xml":
		if buf, err = xml.Marshal(val); err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_XML_MARSHAL_ERROR, requestID, err, val).PrintfError()
			return nil, myerr
		}
	case "yaml":
		if buf, err = yaml.Marshal(val); err != nil {
			myerr = _err.WithCauseTyped(_err.ERR_YAML_MARSHAL_ERROR, requestID, err, val).PrintfError()
			return nil, myerr
		}
	case "xls":
		if val != nil {
			xlsx := _xlsx.NewXlsx(requestID, nil)

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
	default:
		return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "Allowed only 'format'='json', 'yaml', 'xml', 'xls'", format).PrintfError()
	}
	_metrics.IncMarshalingDurationVec(format, operation, name, time.Now().Sub(tic))

	_log.Debug("SUCCESS: requestID, name, duration", requestID, name, time.Now().Sub(tic))
	return buf, nil
}
