package cacheristretto

import (
	"context"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
	_meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// Config конфигурационные настройки
type Config struct {
	//Default        EntityConfig `json:"default,omitempty" yaml:"default"`                 // конфиг по умолчанию для

	NumCounters        int64 `json:"num_counters,omitempty" yaml:"num_counters"`                 // number of keys to track frequency of (10M)
	MaxCost            int64 `json:"max_cost,omitempty" yaml:"max_cost"`                         // maximum cost of cache (1GB)
	BufferItems        int64 `json:"buffer_items,omitempty" yaml:"buffer_items"`                 // number of keys per Get buffer
	Metrics            bool  `json:"metrics,omitempty" yaml:"metrics"`                           // whether cache statistics are kept
	IgnoreInternalCost bool  `json:"ignore_internal_cost,omitempty" yaml:"ignore_internal_cost"` // cost of internally storing the value should be ignored
}

//// EntityConfig конфигурационные настройки
//type EntityConfig struct {
//	NumCounters        int64 `json:"num_counters,omitempty" yaml:"num_counters"`                 // number of keys to track frequency of (10M)
//	MaxCost            int64 `json:"max_cost,omitempty" yaml:"max_cost"`                         // maximum cost of cache (1GB)
//	BufferItems        int64 `json:"buffer_items,omitempty" yaml:"buffer_items"`                 // number of keys per Get buffer
//	Metrics            bool  `json:"metrics,omitempty" yaml:"metrics"`                           // whether cache statistics are kept
//	IgnoreInternalCost bool  `json:"ignore_internal_cost,omitempty" yaml:"ignore_internal_cost"` // cost of internally storing the value should be ignored
//}

// Service represent a Ristretto Cache service
type Service struct {
	ctx    context.Context    // корневой контекст при инициации сервиса
	cancel context.CancelFunc // функция закрытия глобального контекста
	cfg    *Config            // конфигурационные параметры
	errCh  chan<- error       // канал ошибок
	stopCh chan struct{}      // канал подтверждения об успешном закрытии сервиса

	// TODO - Переделать на индивидуальный кэш для каждой сущности чтобы не блокировать все сразу при добавлении и чтении
	cache *ristretto.Cache // Сервис кеширования данных

	mx sync.RWMutex
}

//type EntityCache *ristretto.Cache
//type EntitiesCache map[*_meta.Entity]EntityCache

// New returns a new Service
func New(ctx context.Context, errCh chan<- error, cfg *Config) (*Service, error) {
	var err error
	var requestID = _err.ERR_UNDEFINED_ID

	_log.Info("Creating new Ristretto Cache service")

	{ // входные проверки
		if cfg == nil {
			return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "if cfg == nil {}").PrintfError()
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

	ristrettoConfig := ristretto.Config{
		NumCounters:        cfg.NumCounters,        // number of keys to track frequency of (10M)
		MaxCost:            cfg.MaxCost,            // maximum cost of cache (1GB)
		BufferItems:        cfg.BufferItems,        // number of keys per Get buffer
		Metrics:            cfg.Metrics,            // metrics determines whether cache statistics are kept
		IgnoreInternalCost: cfg.IgnoreInternalCost, // number of keys per Get buffer
	}

	service.cache, err = ristretto.NewCache(&ristrettoConfig)
	if err != nil {
		return nil, err
	}

	_log.Info("Ristretto Cache service was created")
	return service, nil
}

// Shutdown shutting down service
func (s *Service) Shutdown() (err error) {
	_log.Info("Shutdown Ristretto Cache service")

	s.mx.Lock()
	defer s.mx.Unlock()

	defer s.cancel() // закрываем контекст

	if s.cache != nil {
		s.cache.Close()
	}

	_log.Info("Ristretto Cache service shutdown successfully")
	return
}

func formatCacheKeyStr(entity *_meta.Entity, key *_meta.Key, fieldsCacheKey string, keyArgs ...interface{}) string {
	return entity.Name + ":" + key.Name + ":" + fieldsCacheKey + ":" + strings.Join(_meta.ArgsToStrings(keyArgs...), "-")
}

// CacheEntry - элемент кеширования
type CacheEntry struct {
	cacheKey string
	entity   *_meta.Entity
	key      *_meta.Key
	object   *_meta.Object
	mx       sync.RWMutex
}

func newCacheEntry(key *_meta.Key, cacheKey string, object *_meta.Object) *CacheEntry {
	if object != nil {
		e := &CacheEntry{
			cacheKey: cacheKey,
			entity:   object.Entity,
			key:      key,
			object:   object,
		}
		return e
	}
	return nil
}

func (e *CacheEntry) clear() {
	if e != nil {
		e.object = nil // освободить указатель для сбора мусора
	}
}

//// CacheEntryPtr - элемент кеширования
//type CacheEntryPtr struct {
//	entity   *_meta.Entity
//	key      *_meta.Key
//	cacheKey string
//	ptr      interface{}
//	mx       sync.RWMutex
//}
//
//func newCacheEntryPtr(entity *_meta.Entity, key *_meta.Key, cacheKey string, val interface{}) *CacheEntryPtr {
//	if val != nil {
//		e := &CacheEntryPtr{
//			cacheKey: cacheKey,
//			entity:   entity,
//			key:      key,
//			ptr:      val,
//		}
//		return e
//	}
//	return nil
//}

//func (e *CacheEntryPtr) clear() {
//	if e != nil {
//		e.ptr = nil // освободить указатель для сбора мусора
//	}
//}

func (s *Service) CloseAll() {
	if s.cache != nil {
		_log.Info("Close full cache")
		s.mx.Lock()
		defer s.mx.Unlock()
		_log.Debug("Close full cache - after Lock")
		// TODO добавить цикл по сущностям
		s.cache.Close()
		_log.Debug("Close full cache - after Close")
	}
}

func (s *Service) ClearAll() {
	if s.cache != nil {
		_log.Info("Clear full cache")
		s.mx.Lock()
		defer s.mx.Unlock()
		_log.Debug("Clear full cache - after Lock")
		// TODO добавить цикл по сущностям
		s.cache.Clear()
		_log.Debug("Clear full cache - after Clear")
	}
}

func (s *Service) Clear(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		_log.Debug("START: entityName", entity.Name)
		// TODO переделать блокировку отдельно по каждой сущности
		s.mx.Lock()
		defer s.mx.Unlock()
		_log.Debug("START - after Lock: entityName", entity.Name)
		s.clearUnsafe(entity)
		_log.Debug("START - after Clear: entityName", entity.Name)
	}
}

func (s *Service) clearUnsafe(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		//_log.Info("START: entityName", entity.Name)
		s.cache.Clear()
	}
}

func (s *Service) Lock(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		// TODO переделать блокировку отдельно по каждой сущности
		s.mx.Lock()
	}
}

func (s *Service) RLock(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		// TODO переделать блокировку отдельно по каждой сущности
		s.mx.RLock()
	}
}

func (s *Service) Unlock(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		// TODO переделать блокировку отдельно по каждой сущности
		s.mx.Unlock()
	}
}

func (s *Service) RUnlock(entity *_meta.Entity) {
	if s.cache != nil && entity != nil {
		// TODO переделать блокировку отдельно по каждой сущности
		s.mx.RUnlock()
	}
}
