package apiservice

import (
	"context"
	"fmt"
	"time"

	_ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// UpdateEntityMeta обновить метаданные entity
func (s *Service) UpdateEntityMeta(ctx context.Context, inBuf []byte, entityName, format string) (result bool, outBuf []byte, err error) {

	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
	tic := time.Now()
	entityIn := &_meta.Entity{}
	var entityOut *_meta.Entity

	_log.Debug("START: requestID, entityName", requestID, entityName)

	if s.storageMap == nil {
		return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty API Service", []interface{}{s.storageMap}).PrintfError()
	}

	// Тестовый парсинг - определить ошибки во входных данных
	err = s.UnmarshalEntity(requestID, nil, inBuf, entityIn, "UpdateEntityMeta", entityIn.Name, format, false)
	if err != nil {
		return false, nil, err
	}

	if entityIn.Name != entityName {
		return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, fmt.Sprintf("Entity Meta name from params '%s' does not corespond Entity Meta from request body '%s'", entityName, entityIn.Name)).PrintfError()
	}

	if entityExist := s.GetEntity(entityIn.Name); entityExist == nil {
		return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' does not exists", entityName)).PrintfError()
	} else {
		_log.Debug("Found Entity with name 'entity':", entityIn.Name)

		s.metaLock()
		defer s.metaUnLock()
		_log.Info("Lock meta", entityIn.Name)

		// Заполним определение сущности
		definition := &_meta.Definition{
			Name:           entityIn.Name,
			Source:         "REST",
			Format:         format,
			Buf:            inBuf,
			SourceFileName: "",
		}

		entityOut, err = s.setEntityFromDefinitionUnsafe(definition, true, true, true)
		if err != nil {
			return false, nil, err
		}

		// Пересоздадим валидатор, иначе не подхватываются новые правила валидации - bug?
		if err = s.initValidatorUnsafe(); err != nil {
			return false, nil, err
		}

	}

	if entityOut == nil {
		return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Error update Entity Meta")).PrintfError()
	}

	// сформируем ответ
	outBuf, err = s.MarshalEntity(requestID, entityOut, "UpdateEntityMeta", entityOut.Name, format)
	if err != nil {
		return false, nil, err
	}

	_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityOut.Name, time.Now().Sub(tic))

	return true, outBuf, nil
}
