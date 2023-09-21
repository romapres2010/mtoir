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

// CreateEntityMeta создать метаданные entity
func (s *Service) CreateEntityMeta(ctx context.Context, inBuf []byte, format string) (result bool, outBuf []byte, err error) {

	requestID := _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
	tic := time.Now()
	entityIn := &_meta.Entity{}
	var entityOut *_meta.Entity

	_log.Debug("START: requestID", requestID)

	if s.storageMap == nil {
		return false, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty API Service", []interface{}{s.storageMap}).PrintfError()
	}

	// Тестовый парсинг - определить ошибки во входных данных
	err = s.unmarshal(requestID, inBuf, entityIn, "CreateEntityMeta", entityIn.Name, format)
	if err != nil {
		return false, nil, err
	}

	if entityExist := s.GetEntity(entityIn.Name); entityExist == nil {
		_log.Info("Does not Found Entity with name 'entity':", entityIn.Name)

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

	} else {
		return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Entity with name '%s' already exists", entityIn.Name)).PrintfError()
	}

	if entityOut == nil {
		return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Error create Entity Meta")).PrintfError()
	}

	// сформируем ответ
	outBuf, err = s.marshal(requestID, entityOut, "CreateEntityMeta", entityOut.Name, format)
	if err != nil {
		return false, nil, err
	}

	_log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityOut.Name, time.Now().Sub(tic))

	return true, outBuf, nil
}
