package storagedb

import (
	"context"
	"fmt"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_sql "github.com/romapres2010/meta_api/pkg/common/sqlxx"

	_storage "github.com/romapres2010/meta_api/pkg/meta_api/storageservice"
)

type Action string

const (
	PERSIST_ACTION_NONE   Action = "None"   // Фиктивная пустая action
	PERSIST_ACTION_GET    Action = "Get"    // Запрос одного объекта
	PERSIST_ACTION_CREATE Action = "Create" // Создание объекта
	PERSIST_ACTION_UPDATE Action = "Update" // Обновление объекта
	PERSIST_ACTION_DELETE Action = "Delete" // Удаление объекта
	PERSIST_ACTION_MERGE  Action = "Merge"  // Слияние объекта
)

// Config конфигурационные настройки
type Config struct {
	DbMapCfg map[string]*_sql.Config `yaml:"storages" json:"storage"`
}

// Service represent DB API services
type Service struct {
	ctx    context.Context    // корневой контекст при инициации сервиса
	cancel context.CancelFunc // функция закрытия глобального контекста
	cfg    *Config            // конфигурационные параметры
	errCh  chan<- error       // канал ошибок
	stopCh chan struct{}      // канал подтверждения об успешном закрытии сервиса

	storageServiceMap map[string]_storage.Service // Набор сервисов хранения для обработки
	storageMap        map[string]*Storage         // Сервисы внешнего хранения данных
}

// Storage represent DB API service
type Storage struct {
	db      *_sql.DB     // БД для обработки
	dbCfg   *_sql.Config // Конфиг БД для обработки
	service *Service     // родительский сервис
	txCache *txCache     // cache транзакций

}

// New create API service
func New(ctx context.Context, errCh chan<- error, cfg *Config) (*Service, error) {
	_log.Info("Creating new API service")

	{ // входные проверки
		if cfg == nil {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cfg == nil {}").PrintfError()
		}
	} // входные проверки

	// Создаем новый сервис
	service := &Service{
		cfg:    cfg,
		errCh:  errCh,
		stopCh: make(chan struct{}, 1), // канал подтверждения об успешном закрытии сервиса
	}

	// создаем контекст с отменой
	if ctx == nil {
		service.ctx, service.cancel = context.WithCancel(context.Background())
	} else {
		service.ctx, service.cancel = context.WithCancel(ctx)
	}

	if cfg.DbMapCfg != nil {
		service.storageMap = make(map[string]*Storage, len(cfg.DbMapCfg))
		service.storageServiceMap = make(map[string]_storage.Service, len(cfg.DbMapCfg))

		for dbName, dbConfig := range cfg.DbMapCfg {
			if db, err := _sql.New(dbConfig, nil); err != nil {
				return nil, err
			} else {
				storage := &Storage{
					db:      db,
					dbCfg:   dbConfig,
					service: service,
					txCache: newTxCache(),
				}
				service.storageMap[dbName] = storage
				service.storageServiceMap[dbName] = storage
			}
		}
	}

	_log.Info("DB API service was created")
	return service, nil
}

func (s *Service) StorageServiceMap() map[string]_storage.Service {
	if s != nil {
		return s.storageServiceMap
	}
	return nil
}

// Shutdown shutting down service
func (s *Service) Shutdown() (err error) {
	if s != nil {
		_log.Info("Shutdown API service")

		defer s.cancel() // закрываем контекст

		// Откатить все транзакции и закроем все БД
		for dbName, storage := range s.storageMap {

			// Откатить все транзакции, ошибки игнорируем
			_log.Info("Rollback DB transactions: dbName, txCount", dbName, len(storage.txCache.c))
			if storage.txCache != nil {
				for _, cacheEntry := range storage.txCache.c {
					err = storage.db.Rollback(_err.ERR_UNDEFINED_ID, cacheEntry.tx)
					if err != nil {
						_err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err, fmt.Sprintf("DB '%s' transaction '%v' error rollback", dbName, cacheEntry.txId)).PrintfError()
					}
				}
			}

			_log.Info("Close DB: dbName", dbName)
			err = storage.db.Close()
			if err != nil {
				_err.WithCauseTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, err, "Error close db:"+dbName).PrintfError()
			}
		}

		_log.Info("API service shutdown successfully")
		return nil
	}
	_err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil {}", []interface{}{s}).PrintfError()
	return nil
}
