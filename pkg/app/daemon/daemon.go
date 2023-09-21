package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_httplog "github.com/romapres2010/meta_api/pkg/common/httplog"
	_httpserver "github.com/romapres2010/meta_api/pkg/common/httpserver"
	_http "github.com/romapres2010/meta_api/pkg/common/httpservice"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_wpservice "github.com/romapres2010/meta_api/pkg/common/workerpoolservice"

	apiservice "github.com/romapres2010/meta_api/pkg/meta_api/apiservice"
	apicacheservice "github.com/romapres2010/meta_api/pkg/meta_api/cacheristretto"
	apidbservice "github.com/romapres2010/meta_api/pkg/meta_api/storagedb"

	_cfg "github.com/romapres2010/mtoir/pkg/app/config"
	httphandler "github.com/romapres2010/mtoir/pkg/app/httphandler"
)

// Daemon represent top level daemon
type Daemon struct {
	ctx    context.Context    // корневой контекст
	cancel context.CancelFunc // функция закрытия корневого контекста
	cfg    *_cfg.Config       // конфигурация демона

	// Сервисы демона
	httpServer      *_httpserver.Server // HTTP сервер
	httpServerErrCh chan error          // канал ошибок для HTTP сервера

	httpLogger      *_httplog.Logger // сервис логирования HTTP трафика
	httpLoggerErrCh chan error       // канал ошибок для HTTP логгера

	httpService      *_http.Service // сервис HTTP запросов
	httpServiceErrCh chan error     // канал ошибок для HTTP

	httpHandler      *httphandler.Service // сервис обработки HTTP запросов
	httpHandlerErrCh chan error           // канал ошибок для HTTP

	wpService      *_wpservice.Service // сервис worker pool
	wpServiceErrCh chan error          // канал ошибок для сервиса worker pool

	apiDbStorage      *apidbservice.Service // реализация DB API сервиса
	apiDbStorageErrCh chan error            // канал ошибок для DB API сервиса

	apiService      *apiservice.Service // реализация API сервиса
	apiServiceErrCh chan error          // канал ошибок для API сервиса

	apiCacheService      *apicacheservice.Service // реализация API cache сервиса
	apiCacheServiceErrCh chan error               // канал ошибок для API cache сервиса
}

// New create Daemon
func New(ctx context.Context, cfg *_cfg.Config) (*Daemon, error) {
	var err error

	_log.Info("Create new daemon")

	{ // входные проверки
		if cfg == nil {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cfg == nil {}").PrintfError()
		}
	} // входные проверки

	// Создаем новый демон
	daemon := &Daemon{
		cfg:                  cfg,
		httpServerErrCh:      make(chan error, 1),
		httpServiceErrCh:     make(chan error, 1),
		httpHandlerErrCh:     make(chan error, 1),
		httpLoggerErrCh:      make(chan error, 1),
		wpServiceErrCh:       make(chan error, 1),
		apiDbStorageErrCh:    make(chan error, 1),
		apiServiceErrCh:      make(chan error, 1),
		apiCacheServiceErrCh: make(chan error, 1),
	}

	// создаем корневой контекст с отменой
	if ctx == nil {
		daemon.ctx, daemon.cancel = context.WithCancel(context.Background())
	} else {
		daemon.ctx, daemon.cancel = context.WithCancel(ctx)
	}

	// создаем сервис обработчиков
	if daemon.wpService, err = _wpservice.New(daemon.ctx, "WorkerPool - background", daemon.wpServiceErrCh, &daemon.cfg.WorkerPoolServiceCfg); err != nil {
		return nil, err
	}

	// создаем обработчик для логирования HTTP
	if daemon.httpLogger, err = _httplog.New(daemon.ctx, &daemon.cfg.HttpLoggerCfg); err != nil {
		return nil, err
	}

	// создаем сервис API PostgreSQL
	if daemon.cfg.ApiDbStorageCfg != nil {

		// создаем сервис API DB storage
		if daemon.apiDbStorage, err = apidbservice.New(daemon.ctx, daemon.apiDbStorageErrCh, daemon.cfg.ApiDbStorageCfg); err != nil {
			return nil, err
		}

		// создаем сервис API Cache
		if daemon.apiCacheService, err = apicacheservice.New(daemon.ctx, daemon.apiCacheServiceErrCh, daemon.cfg.ApiCacheServiceCfg); err != nil {
			return nil, err
		}

		// создаем сервис API
		if daemon.apiService, err = apiservice.New(daemon.ctx, daemon.apiServiceErrCh, daemon.cfg.ApiServiceCfg, nil, daemon.apiCacheService, nil); err != nil {
			return nil, err
		} else {

			// Регистрируем DB storage
			if err = daemon.apiService.RegisterStorages(daemon.apiDbStorage.StorageServiceMap()); err != nil {
				return nil, err
			}

		}

	} else {
		_log.Info("API service config is null, does not create it")
	}

	// HTTP сервис и HTTP logger
	if daemon.httpService, daemon.httpLogger, err = _http.New(daemon.ctx, &daemon.cfg.HttpServiceCfg, daemon.httpLogger); err != nil {
		return nil, err
	}

	// создаем обработчиков HTTP
	if daemon.httpHandler, err = httphandler.New(daemon.ctx, &daemon.cfg.HttpHandlerCfg, daemon.wpService, daemon.apiService, daemon.httpService); err != nil {
		return nil, err
	}

	// Установим HTTP обработчики
	if err = daemon.httpService.SetHttpHandler(daemon.ctx, daemon.httpHandler); err != nil {
		return nil, err
	}

	// Создаем HTTP server
	if daemon.httpServer, err = _httpserver.New(daemon.ctx, daemon.httpServerErrCh, &daemon.cfg.HttpServerCfg, daemon.httpService); err != nil {
		return nil, err
	}

	_log.Info("New daemon was created")

	return daemon, nil
}

// Run daemon and wait for system signal or error in error channel
func (d *Daemon) Run() error {
	_log.Info("Starting daemon")

	// запускаем сервис обработчиков - паники должны быть обработаны внутри
	go func() { d.wpServiceErrCh <- d.wpService.Run() }()

	// запускаем в фоне HTTP сервер, возврат в канал ошибок - паники должны быть обработаны внутри
	go func() { d.httpServerErrCh <- d.httpServer.Run() }()

	_log.Info("Daemon was running. For exit <CTRL-c>")

	// подписываемся на системные прикрывания
	signalCh := make(chan os.Signal, 1) // канал системных прибываний
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// ожидаем прерывания или возврат в канал ошибок
	for {
		var err error
		select {
		case s := <-signalCh: //  ожидаем системное призывание
			_log.Info("Exiting, got signal", s)
			d.Shutdown(false, d.cfg.ShutdownTimeout) // останавливаем daemon
			return nil
		case err = <-d.httpServerErrCh: // возврат от HTTP сервера в канал ошибок
			_log.Info("Got error from HTTP")
		case err = <-d.wpServiceErrCh: // возврат от обработчиков в канал ошибок
			_log.Info("Got error from worker pool")
		}

		// от сервиса пришла пустая ошибка - игнорируем
		if err != nil {
			_log.Error(err.Error())                 // логируем ошибку
			d.Shutdown(true, d.cfg.ShutdownTimeout) // останавливаем daemon
			return err
		} else {
			_log.Info("Got empty error - ignore it")
		}
	}
}

// Shutdown daemon
func (d *Daemon) Shutdown(hardShutdown bool, shutdownTimeout time.Duration) {
	_log.Info("Shutting down daemon")

	// Закрываем корневой контекст
	defer d.cancel()

	//Останавливаем обработчик worker pool - прерываем обработку текущего задания
	if err := d.wpService.Shutdown(hardShutdown, shutdownTimeout); err != nil {
		_log.ErrorAsInfo(err) // дополнительно логируем результат остановки
	}

	// Останавливаем служебные сервисы
	if err := d.httpService.Shutdown(); err != nil {
		_log.ErrorAsInfo(err) // дополнительно логируем результат остановки
	}

	// Останавливаем HTTP сервер, ожидаем завершения активных подключений
	if err := d.httpServer.Shutdown(); err != nil {
		_log.ErrorAsInfo(err) // дополнительно логируем результат остановки
	}

	// Останавливаем API PostgreSQL сервис
	if d.apiDbStorage != nil {
		if myerr := d.apiDbStorage.Shutdown(); myerr != nil {
			_log.ErrorAsInfo(myerr) // дополнительно логируем результат
		}
	}

	// Останавливаем API CacheService сервис
	if d.apiCacheService != nil {
		if myerr := d.apiCacheService.Shutdown(); myerr != nil {
			_log.ErrorAsInfo(myerr) // дополнительно логируем результат
		}
	}

	// Останавливаем API service сервис
	if d.apiService != nil {
		if myerr := d.apiService.Shutdown(); myerr != nil {
			_log.ErrorAsInfo(myerr) // дополнительно логируем результат
		}
	}

	_log.Info("Daemon was shutdown")

	// Закрываем logger для корректного закрытия лог файла
	if err := d.httpLogger.Shutdown(); err != nil {
		_log.ErrorAsInfo(err) // дополнительно логируем результат остановки
	}
}
