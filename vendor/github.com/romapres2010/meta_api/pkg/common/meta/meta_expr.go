package meta

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
)

type ExprType string

const (
	EXPR_DB_CALCULATE ExprType = "DB Calculate"
	EXPR_CALCULATE    ExprType = "Calculate"
	EXPR_VALIDATE     ExprType = "Validate"
	EXPR_FILTER       ExprType = "Filter"
	EXPR_COPY         ExprType = "CopyField"
	EXPR_CONVERT      ExprType = "Convert"
)

type ExprAction string

const (
	EXPR_ACTION_NULL         ExprAction = "NULL"       // Фиктивная константа - ни чего не делать
	EXPR_ACTION_ALL          ExprAction = "ALL"        // Выполнять все вычисления
	EXPR_ACTION_GET          ExprAction = "GET"        // Срабатывает POST_GET и INSIDE_GET
	EXPR_ACTION_POST_FETCH   ExprAction = "POST_FETCH" // Срабатывает сразу после извлечения одной строки из источника хранения
	EXPR_ACTION_INSIDE_GET   ExprAction = "INSIDE_GET" // Срабатывает в самом конце извлечение полного дерева сущностей
	EXPR_ACTION_POST_GET     ExprAction = "POST_GET"   // Срабатывает после извлечения данных из кэша или внешнего источника по данной сущности
	EXPR_ACTION_PUT          ExprAction = "PUT"
	EXPR_ACTION_PRE_PUT      ExprAction = "PRE_PUT"      // Срабатывает перед началом вставки на данных исходной записи
	EXPR_ACTION_INSIDE_PUT   ExprAction = "INSIDE_PUT"   // Срабатывает перед началом вставки на данных исходной записи
	EXPR_ACTION_POST_PUT     ExprAction = "POST_PUT"     // Срабатывает после вставки на данных вставленной записи
	EXPR_ACTION_MARSHAL      ExprAction = "MARSHAL"      // Срабатывает после разбора одного объекта
	EXPR_ACTION_PRE_MARSHAL  ExprAction = "PRE_MARSHAL"  // Срабатывает после начала разбора одного объекта
	EXPR_ACTION_POST_MARSHAL ExprAction = "POST_MARSHAL" // Срабатывает после разбора всех объектов
)

type Exprs []*Expr

type ExprsByAction map[ExprAction]*Exprs

func (exs *Exprs) copyFrom(exsFrom Exprs, overwrite bool) {
	if exsFrom != nil {

		if overwrite {
			// Удалить все существующие и записать сверху
			exsTo := make(Exprs, 0, len(exsFrom))

			for _, exFrom := range exsFrom {
				if exFrom != nil {
					exTo := &Expr{}
					exTo.copyFrom(exFrom, true)
					exsTo = append(exsTo, exTo)
				}
			}

			// TODO - дублирование при повторной инициации
			*(exs) = exsTo
		} else {
			// Дописать новые
			for _, exFrom := range exsFrom {
				if exFrom != nil {
					exTo := &Expr{}
					exTo.copyFrom(exFrom, true)
					*(exs) = append(*(exs), exTo)
				}
			}
		}
	}
}

// Expr - вычисляемое выражение
type Expr struct {
	Status         string     `yaml:"status" json:"status" xml:"status"`                                              // Статус поля ENABLED, DEPRECATED, ...
	Name           string     `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`                      // Имя выражения
	Type           ExprType   `yaml:"type,omitempty" json:"type,omitempty" xml:"type,omitempty"`                      // Тип выражения
	Action         ExprAction `yaml:"action,omitempty" json:"action,omitempty" xml:"action,omitempty"`                // В каких условиях срабатывает выражение
	Code           string     `yaml:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`                      // Текст выражения
	FieldsArgsName []string   `yaml:"fields,omitempty" json:"fields,omitempty" xml:"fields,omitempty"`                // Поля сущности, участвующие в выражении
	FieldsDestName []string   `yaml:"fields_dest,omitempty" json:"fields_dest,omitempty" xml:"fields_dest,omitempty"` // Поля сущности, которым присвоить результат

	entity           *Entity     // Сущность, к которой относится выражение
	field            *Field      // Поле, к которой относится выражение
	argsFields       Fields      // Поля, участвующие в выражении, после разбора
	argsFieldsMap    FieldsMap   // Поля, участвующие в выражении, после разбора для быстрого поиска
	argsFieldsString string      // Форматированный список полей выражения для вывода в логи и ошибки
	destFields       Fields      // Поля, которым присвоить результат, после разбора
	destFieldsMap    FieldsMap   // Поля, которым присвоить результат, после разбора для быстрого поиска
	destFieldsString string      // Форматированный список полей выражения для вывода в логи и ошибки
	isInit           bool        // признак, что инициация успешная
	program          *vm.Program // Скомпилированное выражение
}

func (ex *Expr) clearInternal() {
	if ex != nil {
		ex.field = nil
		ex.argsFields = nil
		ex.argsFieldsMap = nil
		ex.argsFieldsString = ""
		ex.program = nil
		ex.isInit = false
	}
}

func (ex *Expr) copyFrom(from *Expr, overwrite bool) {
	if from != nil {
		if ex.Status == "" || overwrite {
			ex.Status = from.Status
		}

		if ex.Name == "" || overwrite {
			ex.Name = from.Name
		}

		if ex.Type == "" || overwrite {
			ex.Type = from.Type
		}

		if ex.Action == "" || overwrite {
			ex.Action = from.Action
		}

		if ex.Code == "" || overwrite {
			ex.Code = from.Code
		}

		if len(ex.FieldsArgsName) == 0 || overwrite {
			ex.FieldsArgsName = from.FieldsArgsName
		}
		if len(ex.FieldsDestName) == 0 || overwrite {
			ex.FieldsDestName = from.FieldsDestName
		}

		ex.clearInternal()
	}
}

func (ex *Expr) Entity() *Entity {
	if ex != nil {
		return ex.entity
	}
	return nil
}

func (ex *Expr) ArgsFields() Fields {
	if ex != nil {
		return ex.argsFields
	}
	return nil
}

func (ex *Expr) DestFields() Fields {
	if ex != nil {
		return ex.destFields
	}
	return nil
}

func (ex *Expr) Field() *Field {
	if ex != nil {
		return ex.field
	}
	return nil
}

func (ex *Expr) IsInit() bool {
	if ex != nil {
		return ex.isInit
	}
	return false
}

func (ex *Expr) CheckAction(action ExprAction) bool {
	if ex != nil && ex.entity != nil {

		if ex.Status != STATUS_ENABLED {
			return false
		}

		// Только выражения, которые применимы к типу запроса
		switch action {
		case EXPR_ACTION_GET:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_GET || ex.Action == EXPR_ACTION_INSIDE_GET || ex.Action == EXPR_ACTION_POST_GET) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_PUT:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_PUT || ex.Action == EXPR_ACTION_PRE_PUT || ex.Action == EXPR_ACTION_POST_PUT) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_MARSHAL:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_MARSHAL || ex.Action == EXPR_ACTION_PRE_MARSHAL || ex.Action == EXPR_ACTION_POST_MARSHAL) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_INSIDE_GET:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_INSIDE_GET) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_POST_GET:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_POST_GET) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_PRE_PUT:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_PRE_PUT) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_POST_PUT:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_POST_PUT) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_POST_FETCH:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_POST_FETCH) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_PRE_MARSHAL:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_PRE_MARSHAL) {
				return false
			} else {
				return true
			}
		case EXPR_ACTION_POST_MARSHAL:
			if !(ex.Action == EXPR_ACTION_ALL || ex.Action == EXPR_ACTION_POST_MARSHAL) {
				return false
			} else {
				return true
			}
		default:
			_err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - incorrect action type. Allowed only 'ALL', 'PUT', 'GET', 'MARSHAL', 'INSIDE_GET', 'POST_GET', 'PRE_PUT', 'POST_PUT', 'PRE_MARSHAL', 'POST_MARSHAL', 'POST_FETCH'", ex.entity.Name)).PrintfError()
			return false
		}
	}
	return false
}

func (ex *Expr) Run(externalId uint64, env interface{}, exprVm *vm.VM) (output interface{}, err error) {
	if ex != nil && ex.entity != nil && ex.program != nil {

		if ex.Status != STATUS_ENABLED {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - was not ENABLED", ex.entity.Name, ex.Name, ex.Code))
		}

		if !ex.isInit {
			return nil, _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - is not init", ex.entity.Name, ex.Name, ex.Code))
		}

		//tic := time.Now()
		if exprVm != nil {
			output, err = exprVm.Run(ex.program, env)
		} else {
			output, err = expr.Run(ex.program, env)
		}

		if err != nil {
			return nil, _err.WithCauseTyped(_err.ERR_ERROR, externalId, err, fmt.Sprintf("Entity '%s' Expression '%s' Code=['%s'] - error", ex.entity.Name, ex.Name, ex.Code)).PrintfError()
		} else {
			//_log.Info("SUCCESS: name, code, duration", ex.Name, ex.Code, time.Now().Sub(tic))
			return output, nil
		}
	}
	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, externalId, "if ex != nil && ex.Entity != nil && ex.program != nil {}", []interface{}{ex}).PrintfError()
}
