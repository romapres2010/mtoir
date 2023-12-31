package workerpool

import (
    "context"
    "sync"
    "sync/atomic"
    "time"

    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// TaskPool represent pooling of Task
type TaskPool struct {
    pool sync.Pool
}

// Represent a pool statistics for benchmarking
var (
    countGet uint64 // количество запросов кэша
    countPut uint64 // количество возвратов в кэша
    countNew uint64 // количество создания нового объекта
)

// newTaskPool create new TaskPool
func newTaskPool() *TaskPool {
    p := &TaskPool{
        pool: sync.Pool{
            New: func() interface{} {
                atomic.AddUint64(&countNew, 1)
                task := new(Task)
                task.stopCh = make(chan interface{}, 1)                          // канал закрывается только при получении команды на остановку task
                task.localDoneCh = make(chan interface{}, 1)                     // канал закрывается при timeout и при получении команды на остановку task
                task.timer = time.NewTimer(POOL_MAX_TIMEOUT)                     // новый таймер - начально максимальное время ожидания
                task.timer.Stop()                                                // остановим таймер, сбрасывать канал не требуется, так как он не сработал
                task.ctx, task.cancel = context.WithCancel(context.Background()) // создаем локальный контекст с отменой
                task.setStateUnsafe(TASK_STATE_NEW)                              // установим состояние task
                return task
            },
        },
    }
    return p
}

// getTask allocates a new Task
func (p *TaskPool) getTask() *Task {
    atomic.AddUint64(&countGet, 1)
    task := p.pool.Get().(*Task)
    if task.state != TASK_STATE_NEW {
        task.setStateUnsafe(TASK_STATE_POOL_GET) // установим состояние task
    }
    return task
}

// putTask return Task to pool
func (p *TaskPool) putTask(task *Task) {
    // Если task не был успешно завершен, то в нем могли быть закрыты каналы или сработал таймер - такие не подходят для повторного использования
    if task.state == TASK_STATE_NEW || task.state == TASK_STATE_DONE_SUCCESS || task.state == TASK_STATE_POOL_GET {
        atomic.AddUint64(&countPut, 1)
        task.requests = nil                // обнулить указатель, чтобы освободить для сбора мусора
        task.responses = nil               // обнулить указатель, чтобы освободить для сбора мусора
        task.setState(TASK_STATE_POOL_PUT) // установим состояние task с ожиданием разблокировки
        p.pool.Put(task)                   // отправить в pool
    }
}

// глобальный TaskPool
var gTaskPool = newTaskPool()

// PrintTaskPoolStats print statistics about task pool
func (p *Pool) PrintTaskPoolStats() {
    if p != nil {
        _log.Info("Usage task pool: countGet, countPut, countNew", countGet, countPut, countNew)
    } else {
        _ = _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "p != nil").PrintfError()
	}
}
