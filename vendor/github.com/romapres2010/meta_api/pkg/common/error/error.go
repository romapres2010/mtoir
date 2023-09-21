package errors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	pkgerr "github.com/pkg/errors"

	_log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// ErrorMessage Сообщения
//
//easyjson:json
type ErrorMessage struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

const ERR_UNDEFINED_ID = uint64(0)

var errorID uint64 // уникальный номер ошибки

// getNextErrorID - запросить номер следующей ошибки
func getNextErrorID() uint64 {
	return atomic.AddUint64(&errorID, 1)
}

// SetTypedErrorMessages - установить набор сообщений об ошибках
func SetTypedErrorMessages(errorMessage map[string]ErrorMessage) {
	globalTypedErrorMessages = errorMessage
}

// AppendTypedErrorMessages - добавить сообщения в набор сообщений об ошибках
func AppendTypedErrorMessages(errorMessage map[string]ErrorMessage) {
	for s, message := range errorMessage {
		globalTypedErrorMessages[s] = message
	}
}

type Errors []*Error

func (errors *Errors) HasError() bool {
	if errors != nil {
		return len(*errors) > 0
	}
	return false
}

func (errors *Errors) Clear() {
	if errors != nil {
		errors = new(Errors)
	}
}

func (errors *Errors) Append(externalId uint64, err error) *Errors {
	if errors != nil {
		if err != nil {
			if myErr, ok := err.(*Error); ok {
				*errors = append(*errors, myErr)
			} else {
				myErr = WithCauseTyped(ERR_ERROR, externalId, err, err.Error())
				*errors = append(*errors, myErr)
			}
		}
		return errors
	}
	return nil
}

func (errors *Errors) AppendErrors(addErrors Errors) Errors {
	if errors != nil {
		if len(addErrors) > 0 {
			*errors = append(*errors, addErrors...)
		}
		return *errors
	}
	return nil
}

func (errors *Errors) Error(externalId uint64, messageText string) error {
	if errors != nil {
		if len(*errors) > 0 {
			err := NewTypedTraceEmpty(ERR_ERROR, externalId, messageText)
			err.Trace = ""
			err.Args = ""
			for _, e := range *errors {
				e.Trace = ""
				err.Errors = append(err.Errors, e)
			}
			return err
		}
	}
	return nil
}

// Error represent custom error
type Error struct {
	ID           uint64          `json:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`                                  // уникальный номер ошибки
	ExternalId   uint64          `json:"external_id,omitempty" xml:"external_id,omitempty" yaml:"external_id,omitempty"`       // внешний ID, который был передан при создании ошибки
	Code         string          `json:"code,omitempty" xml:"code,omitempty" yaml:"code,omitempty"`                            // код ошибки
	Message      string          `json:"message,omitempty" xml:"message,omitempty" yaml:"message,omitempty"`                   // текст ошибки
	Details      []string        `json:"details,omitempty" xml:"details,omitempty" yaml:"details,omitempty"`                   // текст ошибки подробный
	Caller       string          `json:"caller,omitempty" xml:"caller,omitempty" yaml:"caller,omitempty"`                      // файл, строка и наименование метода в котором произошла ошибка
	Args         string          `json:"args,omitempty" xml:"args,omitempty" yaml:"args,omitempty"`                            // строка аргументов
	CauseMessage string          `json:"cause_message,omitempty" xml:"cause_message,omitempty" yaml:"cause_message,omitempty"` // текст ошибки - причины
	CauseErr     error           `json:"cause_err,omitempty" xml:"cause_err,omitempty" yaml:"cause_err,omitempty"`             // ошибка - причина
	Trace        string          `json:"trace,omitempty" xml:"trace,omitempty" yaml:"trace,omitempty"`                         // стек вызова
	Errors       Errors          `json:"errors,omitempty" xml:"errors,omitempty" yaml:"errors,omitempty"`                      // связанные ошибки
	Extra        interface{}     `json:"extra_info,omitempty" xml:"extra_info,omitempty" yaml:"extra_info,omitempty"`          // дополнительная информация
	MessageJson  json.RawMessage `json:"extra_info_json,omitempty"`                                                            // текст ошибки в formate Json
}

// Format output
//
// ,omitempty"%s    print the error code, message, arguments, and cause message.
//
//	%v    in addition to %s, print caller
//	%+v   extended format. Each Frame of the error's StackTrace will be printed in detail.
func (e *Error) Format(s fmt.State, verb rune) {
	if e != nil {
		mes := e.Error()
		switch verb {
		case 'v':
			mes = strings.Join([]string{mes, ", caller=[", e.Caller, "]"}, "")
			//_, _ = fmt.Fprint(s, e.Error())
			//_, _ = fmt.Fprintf(s, ", caller=[%s]", e.Caller)
			if s.Flag('+') {
				mes = strings.Join([]string{mes, ", trace=[", e.Trace, "]"}, "")
				//_, _ = fmt.Fprintf(s, ", trace=%s", e.Trace)
			}
			_, _ = fmt.Fprint(s, mes)
		case 's':
			_, _ = fmt.Fprint(s, mes)
			//_, _ = fmt.Fprint(s, e.Error())
		case 'q':
			_, _ = fmt.Fprint(s, mes)
			//_, _ = fmt.Fprint(s, e.Error())
		}
	}
}

// GetTypedMess Добавить типизированное сообщение
func GetTypedMess(key string, args ...interface{}) (code string, text string) {
	if v, ok := globalTypedErrorMessages[key]; ok {
		code = v.Code
		text = fmt.Sprintf(v.Text, args...)
	} else {
		v := globalTypedErrorMessages[ERR_MESSAGE_NOT_FOUND]
		text = fmt.Sprintf(v.Text, key, GetArgsString(args...))
	}

	return code, text
}

// Error print custom error
func (e *Error) Error() string {
	if e != nil {
		mes := strings.Join([]string{"errID=[", strconv.FormatUint(e.ID, 10), "], extID=[", strconv.FormatUint(e.ExternalId, 10), "], code=[", e.Code, "], message=[", e.Message, "]"}, "")
		//mes := fmt.Sprintf("errID=[%v], extID=[%v], code=[%s], message=[%s]", e.ID, e.ExternalId, e.Code, e.Message)
		if e.Details != nil && len(e.Details) > 0 {
			mes = fmt.Sprintf("%s, details=[%+v]", mes, e.Details)
		}
		if e.CauseMessage != "" {
			mes = strings.Join([]string{mes, ", cause_message=[", e.CauseMessage, "]"}, "")
			//mes = fmt.Sprintf("%s, cause_message=[%s]", mes, e.CauseMessage)
		}
		return mes
	}
	return ""
}

// PrintfDebug print custom error
func (e *Error) PrintfDebug(depths ...int) *Error {
	if e != nil {
		depth := 1
		if len(depths) == 1 {
			depth = depth + depths[0]
		}
		_log.Log(_log.LEVEL_DEBUG, depth, e.Error())
		return e
	}
	return nil
}

// PrintfInfo print custom error
func (e *Error) PrintfInfo(depths ...int) *Error {
	if e != nil {
		depth := 1
		if len(depths) == 1 {
			depth = depth + depths[0]
		}
		_log.Log(_log.LEVEL_INFO, depth, e.Error())
		return e
	}
	return nil
}

// PrintfError print custom error
func (e *Error) PrintfError(depths ...int) *Error {
	if e != nil {
		depth := 1
		if len(depths) == 1 {
			depth = depth + depths[0]
		}
		_log.Log(_log.LEVEL_ERROR, depth, e.Error())
		return e
	}
	return nil
}

// GetTrace - напечатать trace
func GetTrace() string {
	return fmt.Sprintf("'%+v'", pkgerr.New(""))
}

// New - create new custom error
func New(code string, msg string, args ...interface{}) *Error {
	err := Error{
		ID:         getNextErrorID(),
		ExternalId: ERR_UNDEFINED_ID,
		Code:       code,
		Message:    msg,
		Caller:     _log.GetCaller(4),
		Args:       GetArgsString(args...),               // get formatted string with arguments
		Trace:      fmt.Sprintf("'%+v'", pkgerr.New("")), // create err and print it trace
		Errors:     make(Errors, 0),
	}

	return &err
}

// NewTraceEmpty - create new custom error
func NewTraceEmpty(code string, msg string, args ...interface{}) *Error {
	err := Error{
		ID:         getNextErrorID(),
		ExternalId: ERR_UNDEFINED_ID,
		Code:       code,
		Message:    msg,
		//Caller:     _log.GetCaller(4),
		Args: GetArgsString(args...), // get formatted string with arguments
		//Trace:      fmt.Sprintf("'%+v'", pkgerr.New("")), // create err and print it trace
		Errors: make(Errors, 0),
	}

	return &err
}

// NewTyped - create new typed custom error
func NewTyped(key string, externalId uint64, args ...interface{}) *Error {
	code, msg := GetTypedMess(key, args...)
	//err := New(code, msg, args...)
	err := New(code, msg)
	err.ExternalId = externalId
	return err
}

// NewTypedTraceEmpty - create new typed custom error
func NewTypedTraceEmpty(key string, externalId uint64, args ...interface{}) *Error {
	code, msg := GetTypedMess(key, args...)
	//err := New(code, msg, args...)
	err := NewTraceEmpty(code, msg)
	err.ExternalId = externalId
	return err
}

// WithCause - create new custom error with cause
func WithCause(code string, msg string, causeErr error, args ...interface{}) *Error {
	//err := New(code, msg, args...)
	err := NewTraceEmpty(code, msg, args...)
	if causeErr != nil {
		//err.CauseMessage = fmt.Sprintf("'%+v'", causeErr) // get formatted string from cause error
		err.CauseMessage = causeErr.Error()
		err.CauseErr = causeErr
	}
	return err
}

// WithCauseTyped - create new custom error with cause
func WithCauseTyped(key string, externalId uint64, causeErr error, args ...interface{}) *Error {
	code, msg := GetTypedMess(key, args...)
	//err := WithCause(code, msg, causeErr, args...)
	err := WithCause(code, msg, causeErr)
	err.ExternalId = externalId
	return err
}

// GetArgsString return formated string with arguments
func GetArgsString(args ...interface{}) (argsStr string) {
	for _, arg := range args {
		if arg != nil {
			argsStr = argsStr + fmt.Sprintf("'%v', ", arg)
		}
	}
	argsStr = strings.TrimRight(argsStr, ", ")
	return
}
