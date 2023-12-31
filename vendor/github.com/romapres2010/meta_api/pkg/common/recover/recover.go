package recover

import (
    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// GetRecoverError - формирует и логирует ошибку
func GetRecoverError(r any, externalId uint64, args ...interface{}) (err error) {
    if r != nil {
        caller := _log.GetCallerShort(4)
        _log.Info("Recover from panic: [caller='" + caller + "']")
        switch t := r.(type) {
        case error:
            err = _err.WithCauseTyped(_err.ERR_PANIC_RECOVER_ERROR, externalId, t, caller, t, args)
            _log.Log(_log.LEVEL_ERROR, 2, err.Error())
        default:
            err = _err.NewTyped(_err.ERR_PANIC_RECOVER_ERROR, externalId, caller, t, args)
            _log.Log(_log.LEVEL_ERROR, 2, err.Error())
        }
        return err
    } else {
        return nil
	}
}
