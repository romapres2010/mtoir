package apiservice

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"gopkg.in/yaml.v3"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
	_recover "github.com/romapres2010/meta_api/pkg/common/recover"
	_validator "github.com/romapres2010/meta_api/pkg/common/validator"

	cache "github.com/romapres2010/meta_api/pkg/meta_api/cacheservice"
	_storage "github.com/romapres2010/meta_api/pkg/meta_api/storageservice"
)

// Config конфигурационные настройки
type Config struct {
	EntityConfigFile     string            `yaml:"entity_config_file" json:"entity_config_file"`
	PopulateCacheOnStart bool              `yaml:"populate_cache_on_start" json:"populate_cache_on_start"`
	QueryOption          QueryOptionConfig `yaml:"query_option" json:"query_option"`
	EntityConfig         EntityConfig      `yaml:"entity_config" json:"entity_config"`
	Meta                 *_meta.Meta       `yaml:"meta" json:"meta"`
}

// EntityConfig конфигурационные настройки сущностей
type EntityConfig struct {
	EntityDirName string             `yaml:"source_dir" json:"source_dir"`
	Definition    []_meta.Definition `yaml:"definition,omitempty" json:"definition,omitempty" xml:"definition,omitempty"` // Определение сущностей
}

type QueryOptionConfig struct {
	DelimiterStart   string `yaml:"delimiter_start" json:"delimiter_start"`       // [ разделить начало для определения спец параметров
	DelimiterEnd     string `yaml:"delimiter_end" json:"delimiter_end"`           // ] разделить конец для определения спец параметров
	FromEntity       string `yaml:"from_entity" json:"from_entity"`               // [from_entity] имя входной сущности
	Fields           string `yaml:"fields" json:"fields"`                         // [fields] фильтрация список полей в ответе
	SkipCache        string `yaml:"skip_cache" json:"skip_cache"`                 // [skip_cache] принудительно считать из внешнего источника
	SkipCalculation  string `yaml:"skip_calculation" json:"skip_calculation"`     // [skip_calculation] принудительно отключить все вычисления
	UseCache         string `yaml:"use_cache" json:"use_cache"`                   // [use_cache] принудительно использовать кеширование - имеет приоритет над skip_cache
	EmbedError       string `yaml:"embed_error" json:"embed_error"`               // [embed_error] встраивать отдельные типы некритичных ошибок в текст ответа
	CascadeUp        string `yaml:"cascade_up" json:"cascade_up"`                 // [cascade_up] сколько уровней вверх по FK
	CascadeDown      string `yaml:"cascade_down" json:"cascade_down"`             // [cascade_down] сколько уровней вниз по FK
	TxExternal       string `yaml:"tx" json:"tx"`                                 // [tx] идентификатор внешней транзакции
	IgnoreExtraField string `yaml:"ignore_extra_field" json:"ignore_extra_field"` // [ignore_extra_field] игнорировать лишние поля в параметрах запроса
	NameFormat       string `yaml:"name_format" json:"name_format"`               // [name_format] формат именования полей в параметрах запроса 'json', 'yaml', 'xml', 'xsl', 'name'
	OutFormat        string `yaml:"out_format" json:"out_format"`                 // [out_format] формат вывода результата 'json', 'yaml', 'xml', 'xsl'
	OutTrace         string `yaml:"out_trace" json:"out_trace"`                   // [out_trace] вывод трассировки
	Validate         string `yaml:"validate" json:"validate"`                     // [validate] проверка данных
	MultiRow         string `yaml:"multi_row" json:"multi_row"`                   // [multi_row] признак многострочной обработки
	Filter           string `yaml:"filter" json:"filter"`                         // [filter] признак  фильтрации
	StaticFiltering  string `yaml:"static_filtering" json:"static_filtering"`     // [static_filtering] признак статической фильтрации
	Persist          string `yaml:"persist" json:"persist"`                       // [persist] признак, что отправлять данные в хранилище
	DbOrder          string `yaml:"db_order" json:"db_order"`                     // [db_order] последовательность сортировки строк в ответе
	DbWhere          string `yaml:"db_where" json:"db_where"`                     // [db_where] фраза where для встраивания в запрос
	DbLimit          string `yaml:"db_limit" json:"db_limit"`                     // [db_limit] ограничение на выборку данных в запросе
	DbOffset         string `yaml:"db_offset" json:"db_offset"`                   // [db_offset] сдвиг строки, с которой начать выводить данные в запросе

	DelimiterStartFilter string `yaml:"-" json:"-"`
	DelimiterStartFull   string `yaml:"-" json:"-"`
	DelimiterEndFull     string `yaml:"-" json:"-"`
	FromEntityFull       string `yaml:"-" json:"-"`
	FieldsFull           string `yaml:"-" json:"-"`
	SkipCacheFull        string `yaml:"-" json:"-"`
	SkipCalculationFull  string `yaml:"-" json:"-"`
	UseCacheFull         string `yaml:"-" json:"-"`
	EmbedErrorFull       string `yaml:"-" json:"-"`
	CascadeUpFull        string `yaml:"-" json:"-"`
	CascadeDownFull      string `yaml:"-" json:"-"`
	TxExternalFull       string `yaml:"-" json:"-"`
	IgnoreExtraFieldFull string `yaml:"-" json:"-"`
	NameFormatFull       string `yaml:"-" json:"-"`
	OutFormatFull        string `yaml:"-" json:"-"`
	OutTraceFull         string `yaml:"-" json:"-"`
	ValidateFull         string `yaml:"-" json:"-"`
	MultiRowFull         string `yaml:"-" json:"-"`
	FilterFull           string `yaml:"-" json:"-"`
	StaticFilteringFull  string `yaml:"-" json:"-"`
	PersistFull          string `yaml:"-" json:"-"`
	DbOrderFull          string `yaml:"-" json:"-"`
	DbWhereFull          string `yaml:"-" json:"-"`
	DbLimitFull          string `yaml:"-" json:"-"`
	DbOffsetFull         string `yaml:"-" json:"-"`
}

func (cfg *QueryOptionConfig) init() {
	if cfg != nil {
		delimiterStart := cfg.DelimiterStart
		delimiterEnd := cfg.DelimiterEnd

		cfg.DelimiterStartFilter = delimiterStart + cfg.Filter
		cfg.DelimiterStartFull = delimiterStart + cfg.DelimiterStart + delimiterEnd
		cfg.DelimiterEndFull = delimiterStart + cfg.DelimiterEnd + delimiterEnd
		cfg.FromEntityFull = delimiterStart + cfg.FromEntity + delimiterEnd
		cfg.FieldsFull = delimiterStart + cfg.Fields + delimiterEnd
		cfg.SkipCacheFull = delimiterStart + cfg.SkipCache + delimiterEnd
		cfg.SkipCalculationFull = delimiterStart + cfg.SkipCalculation + delimiterEnd
		cfg.UseCacheFull = delimiterStart + cfg.UseCache + delimiterEnd
		cfg.EmbedErrorFull = delimiterStart + cfg.EmbedError + delimiterEnd
		cfg.CascadeUpFull = delimiterStart + cfg.CascadeUp + delimiterEnd
		cfg.CascadeDownFull = delimiterStart + cfg.CascadeDown + delimiterEnd
		cfg.TxExternalFull = delimiterStart + cfg.TxExternal + delimiterEnd
		cfg.IgnoreExtraFieldFull = delimiterStart + cfg.IgnoreExtraField + delimiterEnd
		cfg.NameFormatFull = delimiterStart + cfg.NameFormat + delimiterEnd
		cfg.OutFormatFull = delimiterStart + cfg.OutFormat + delimiterEnd
		cfg.OutTraceFull = delimiterStart + cfg.OutTrace + delimiterEnd
		cfg.ValidateFull = delimiterStart + cfg.Validate + delimiterEnd
		cfg.MultiRowFull = delimiterStart + cfg.MultiRow + delimiterEnd
		cfg.FilterFull = delimiterStart + cfg.Filter + delimiterEnd
		cfg.StaticFilteringFull = delimiterStart + cfg.StaticFiltering + delimiterEnd
		cfg.PersistFull = delimiterStart + cfg.Persist + delimiterEnd
		cfg.DbOrderFull = delimiterStart + cfg.DbOrder + delimiterEnd
		cfg.DbWhereFull = delimiterStart + cfg.DbWhere + delimiterEnd
		cfg.DbOffsetFull = delimiterStart + cfg.DbOffset + delimiterEnd
		cfg.DbLimitFull = delimiterStart + cfg.DbLimit + delimiterEnd
	}
}

type Action string

const (
	PERSIST_ACTION_NONE   Action = "None"   // Фиктивная пустая action
	PERSIST_ACTION_GET    Action = "Get"    // Запрос одного объекта
	PERSIST_ACTION_CREATE Action = "Create" // Создание объекта
	PERSIST_ACTION_UPDATE Action = "Update" // Обновление объекта
	PERSIST_ACTION_DELETE Action = "Delete" // Удаление объекта
	PERSIST_ACTION_MERGE  Action = "Merge"  // Слияние объекта
)

type ctxKeType int

const ctxCacheKey ctxKeType = 3
const ctxTxIdKey ctxKeType = 4

// Service represent API service
type Service struct {
	ctx    context.Context    // корневой контекст при инициации сервиса
	cancel context.CancelFunc // функция закрытия глобального контекста
	cfg    *Config            // конфигурационные параметры
	errCh  chan<- error       // канал ошибок
	stopCh chan struct{}      // канал подтверждения об успешном закрытии сервиса

	meta        *_meta.Meta                 // метаданные сервиса
	storageMap  map[string]_storage.Service // Сервисы внешнего хранения данных
	globalCache cache.CacheService          // Глобальный сервис кеширования данных
	validator   *_validator.Validator       // Валидатор модели данных, настроенный по правилам

	mx sync.RWMutex
}

// New returns a new Service
func New(ctx context.Context, errCh chan<- error, cfg *Config, storageMap map[string]_storage.Service, cache cache.CacheService, gMeta *_meta.Meta) (*Service, error) {
	var err error
	var requestID = _err.ERR_UNDEFINED_ID

	_log.Info("Creating new API service")

	{ // входные проверки
		if cfg == nil {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if cfg == nil {}").PrintfError()
		}
		if cache == nil {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if globalCache == nil {}").PrintfError()
		}
	} // входные проверки

	cfg.QueryOption.init()

	// Создаем новый сервис
	service := &Service{
		cfg:         cfg,
		errCh:       errCh,
		globalCache: cache,
		stopCh:      make(chan struct{}, 1), // канал подтверждения об успешном закрытии сервиса
		meta:        _meta.NewMeta(),
	}

	// storage могут регистрироваться позже
	if storageMap != nil {
		service.storageMap = storageMap
	} else {
		service.storageMap = make(map[string]_storage.Service)
	}

	service.meta.Status = _meta.STATUS_ENABLED
	service.meta.Name = "API Service meta"

	if cfg.EntityConfigFile != "" {
		if err = cfg.loadYamlConfig(); err != nil {
			return nil, err
		}

		// Считать все definition и создать Entity
		for _, definition := range cfg.EntityConfig.Definition {

			if err = definition.LoadFromFile(cfg.EntityConfig.EntityDirName); err != nil {
				return nil, err
			}

			if _, err = service.setEntityFromDefinitionUnsafe(&definition, false, false, false); err != nil {
				return nil, err
			}
		}
		if err = service.metaInitUnsafe(); err != nil {
			return nil, err
		}
	}

	if gMeta != nil {
		// copy of gMeta
		for _, entity := range gMeta.Entities {
			if entity != nil {
				if err = service.meta.Set(entity, false, false); err != nil {
					return nil, err
				}
			}
		}
	}

	// copy of cfg.Meta
	if cfg.Meta != nil {
		for _, entity := range cfg.Meta.Entities {
			if entity != nil {
				if err = service.meta.Set(entity, true, false); err != nil {
					return nil, err
				}
			}
		}
	}

	if err = service.meta.Init(); err != nil {
		return nil, err
	}

	// настроим валидатор
	if err = service.InitValidator(); err != nil {
		return nil, err
	}

	// создаем контекст с отменой
	if ctx == nil {
		service.ctx, service.cancel = context.WithCancel(context.Background())
	} else {
		service.ctx, service.cancel = context.WithCancel(ctx)
	}

	_log.Info("API service was created")
	return service, nil
}

// loadYamlConfig load configuration file
func (cfg *Config) loadYamlConfig() error {
	if cfg != nil && cfg.EntityConfigFile != "" {
		var err error
		var fileInfo os.FileInfo
		var file *os.File
		var fileName = cfg.EntityConfigFile

		if fileName == "" {
			return _err.NewTyped(_err.ERR_EMPTY_CONFIG_FILE, _err.ERR_UNDEFINED_ID).PrintfError()
		}
		_log.Info("Loading meta config from file: FileName", fileName)

		// Считать информацию о файле
		fileInfo, err = os.Stat(fileName)
		if os.IsNotExist(err) {
			return _err.NewTyped(_err.ERR_CONFIG_FILE_NOT_EXISTS, _err.ERR_UNDEFINED_ID, fileName).PrintfError()
		}

		_log.Debug("Meta config file exist: FileName, FileInfo", fileName, fileInfo)

		{ // Считать конфигурацию из файла
			file, err = os.Open(fileName)
			if err != nil {
				return err
			}
			defer func() {
				if file != nil {
					err = file.Close()
					if err != nil {
						_ = _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, "config.loadYamlConfig -> os.File.Close()").PrintfError()
					}
				}
			}()

			decoder := yaml.NewDecoder(file)
			err = decoder.Decode(&cfg.EntityConfig)

			if err != nil {
				return _err.WithCauseTyped(_err.ERR_CONFIG_FILE_LOAD_ERROR, _err.ERR_UNDEFINED_ID, err, fileName).PrintfError()
			}
		} // Считать конфигурацию из файла

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if cfg != nil {}").PrintfError()
}

func (s *Service) getStorageByEntity(entity *_meta.Entity) (_storage.Service, error) {
	if s != nil && s.storageMap != nil && entity != nil {
		if entity.StorageName != "" {
			return s.getStorageByName(entity.StorageName)
		} else {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - empty 'storage_name'", entity.Name))
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.storageMap != nil && entity != nil {}", []interface{}{s, entity}).PrintfError()
}

func (s *Service) getStorageByName(name string) (_storage.Service, error) {
	if s != nil && s.storageMap != nil {
		if name != "" {
			if db, ok := s.storageMap[name]; ok {
				return db, nil
			} else {
				return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Storage with name '%s' does not registered", name)).PrintfError()
			}
		} else {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Empty storage name '%s'", name)).PrintfError()
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.storageMap != nil && name != \"\" {}", []interface{}{s, name}).PrintfError()
}

func (s *Service) RegisterStorages(storageMap map[string]_storage.Service) (err error) {
	if s != nil && s.storageMap != nil && storageMap != nil {
		for name, storage := range storageMap {
			if err = s.RegisterStorage(name, storage); err != nil {
				return err
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.storageMap != nil && storageMap != nil {}", []interface{}{s, storageMap}).PrintfError()
}

func (s *Service) RegisterStorage(name string, storage _storage.Service) (err error) {
	if s != nil && s.storageMap != nil && storage != nil {
		if _, ok := s.storageMap[name]; ok {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Storage with name '%s' already exists", name))
		} else {
			s.storageMap[name] = storage
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.storageMap != nil && storage != nil {}", []interface{}{s, storage}).PrintfError()
}

func (s *Service) InitValidator() (err error) {
	if s != nil {
		_log.Info("START: Configure validator")

		s.mx.RLock()
		defer s.mx.RUnlock()

		return s.initValidatorUnsafe()
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil {}", []interface{}{s}).PrintfError()
}

func (s *Service) initValidatorUnsafe() (err error) {
	if s != nil {

		if s.validator, err = _validator.NewValidator(s.meta, true); err != nil {
			return err
		} else {
			_log.Info("START: Configure validator")

			// Настроим правила валидации для каждой сущности - настраивать нужно на динамически созданные структуры
			for _, entity := range s.meta.Entities {

				_, row, err := entity.NewStruct(nil, entity.DefTypeCacheKey())
				if err != nil {
					return nil
				}
				rules := entity.ValidationRules()
				s.validator.RegisterValidationRules(entity.ValidationRules(), row.Value)

				_log.Debug("RegisterValidationRules: externalId, entity.Name, entity.Type, rules", _err.ERR_UNDEFINED_ID, entity.Name, reflect.TypeOf(row), rules)
			}
		}
		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil {}", []interface{}{s}).PrintfError()
}

func (s *Service) metaLock() {
	if s != nil && s.meta != nil {
		s.mx.Lock()
		//_log.Info("After Service lock")
		defer s.mx.Unlock()

		s.meta.Lock()
		//_log.Info("After Meta lock")
	}
}

func (s *Service) metaUnLock() {
	if s != nil && s.meta != nil {
		s.meta.Unlock()
		//_log.Info("After Meta UnLock")
	}
}

func (s *Service) metaRLock() {
	if s != nil && s.meta != nil {
		s.mx.RLock()
		defer s.mx.RUnlock()

		s.meta.RLock()
	}
}

func (s *Service) metaRUnLock() {
	if s != nil && s.meta != nil {
		s.meta.RUnlock()
	}
}

func (s *Service) GetEntity(entityName string) *_meta.Entity {
	if s != nil && s.meta != nil {
		s.mx.RLock()
		defer s.mx.RUnlock()

		return s.meta.GetEntity(entityName)
		//return s.getEntityUnsafe(entityName)
	}
	_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.meta != nil {}", []interface{}{s}).PrintfError()
	return nil
}

func (s *Service) getEntityUnsafe(entityName string) *_meta.Entity {
	if s != nil && s.meta != nil {
		return s.meta.GetEntityUnsafe(entityName)
	}
	_ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.meta != nil {}", []interface{}{s}).PrintfError()
	return nil
}

func (s *Service) metaInitUnsafe() (err error) {
	if s != nil && s.meta != nil {

		_log.Info("Init full meta")

		// Сохраним предыдущее состояние, если ошибка, то восстановим
		copyMeta := s.meta.Backup()

		// Функция восстановления после паники
		defer func() {
			r := recover()
			if r != nil {
				s.meta.Restore(copyMeta)
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "metaInitUnsafe", s.meta.Name)
			}
		}()

		// Полность инициируем всю meta
		if err = s.meta.InitUnsafe(); err != nil {
			s.meta.Restore(copyMeta)
			return err
		}

		if s.globalCache != nil {
			s.globalCache.ClearAll()
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.meta != nil {}", []interface{}{s}).PrintfError()
}

func (s *Service) setEntityFromDefinitionUnsafe(definition *_meta.Definition, doEntityReplace bool, doEntityInit bool, doMetaInit bool) (entity *_meta.Entity, err error) {
	if s != nil && s.meta != nil {

		_log.Info("Register entity from definition: definition.Name, definition.Format, definition.Source, definition.SourceFileName", definition.Name, definition.Format, definition.Source, definition.SourceFileName)

		// Сохраним предыдущее состояние, если ошибка, то восстановим
		copyMeta := s.meta.Backup()

		// Функция восстановления после паники
		defer func() {
			r := recover()
			if r != nil {
				s.meta.Restore(copyMeta)
				err = _recover.GetRecoverError(r, _err.ERR_UNDEFINED_ID, "setEntityFromDefinitionUnsafe", entity.Name)
			}
		}()

		entity, err = s.meta.SetEntityFromDefinitionUnsafe(definition, doEntityReplace, doEntityInit)
		if err != nil {
			s.meta.Restore(copyMeta)
			return nil, err
		}

		if s.globalCache != nil {
			s.globalCache.Clear(entity)
		}

		// Полность инициируем всю meta
		if doMetaInit {
			if err = s.meta.InitUnsafe(); err != nil {
				s.meta.Restore(copyMeta)
				return nil, err
			}
		}

		return entity, nil
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.meta != nil {}", []interface{}{s}).PrintfError()
}

//func (s *Service) setEntityUnsafe(entity *_meta.Entity) error {
//	if s != nil && s.meta != nil {
//
//		_log.Info("START: entityName", entity.Name)
//
//		if err := s.meta.setEntityUnsafe(entity, true, true); err != nil {
//			return err
//		}
//
//		if s.globalCache != nil {
//			s.globalCache.Clear(entity)
//		}
//
//		return nil
//	}
//	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if s != nil && s.meta != nil {}", []interface{}{s}).PrintfError()
//}

// Shutdown shutting down service
func (s *Service) Shutdown() (err error) {
	_log.Info("Shutdown API service")

	//s.mx.Lock()
	//defer s.mx.Unlock()

	defer s.cancel() // закрываем контекст

	_log.Info("Meta API service shutdown successfully")
	return
}
