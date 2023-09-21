package xlsx

import (
    _err "github.com/romapres2010/meta_api/pkg/common/error"
)

func (xls *Xlsx) WriteObject(val any, opt WriteOption) (err error) {
    if xls != nil {

    }
    return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "s != nil", []interface{}{xls}).PrintfError()
}
