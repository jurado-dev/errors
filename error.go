package errors

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
)

type Err struct {
	Cause        string     `json:"cause"`
	Message      string     `json:"message"`
	StackMessage string     `json:"stack_message"`
	Trace        ErrTrace   `json:"trace"`
	Stack        []ErrTrace `json:"stack"`
	Wrapped      error      `json:"-"`
	Code         int        `json:"code"`
}

type ErrTrace struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

//	Error implements the interface
func (e *Err) Error() string {

	output := fmt.Sprintf("Info: %s", e.Message)
	if e.Trace.Line != 0 {

		cause := e.Cause
		if len(cause) > 200 {
			cause = cause[:200] + "..."
		}

		output = fmt.Sprintf("Cause: %s | Info: %s | Line: %d | Function: %s", cause, e.Message, e.Trace.Line, e.Trace.Function)
	}

	return output
}

func extractErr(err error) Err {

	switch e := err.(type) {
	case *Internal:
		return e.Err
	case *NotFound:
		return e.Err
	case *Conflict:
		return e.Err
	case *BadRequest:
		return e.Err
	case *Unauthorized:
		return e.Err
	case *Fatal:
		return e.Err
	case *NoContent:
		return e.Err
	case *Timeout:
		return e.Err
	}

	return Err{}
}

//	Stack adds a trace to the stack slice
func Stack(err error, trace ErrTrace) error {

	if err == nil {
		return err
	}

	switch e := err.(type) {
	case *Internal:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *NotFound:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *Conflict:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *BadRequest:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *Unauthorized:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *Fatal:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *NoContent:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	case *Timeout:
		e.Err.Stack = append(e.Err.Stack, trace)
		return e
	}

	return err
}

func StackMsg(err error, msg string, trace ErrTrace) error {
	if err == nil {
		return err
	}

	switch e := err.(type) {
	case *Internal:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *NotFound:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *Conflict:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *BadRequest:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *Unauthorized:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *Fatal:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *NoContent:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	case *Timeout:
		e.Err.Stack = append(e.Err.Stack, trace)
		e.StackMessage = msg
		return e
	}
	return err
}

//	ErrorF returns the full error information
func ErrorF(err error) string {

	if err == nil {
		return ""
	}

	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err.Error()
	}

	var stackTrace string

	traceFormat := "> Line=%-4d | Function=%-40s | File=%-30s"

	for _, stack := range e.Stack {
		stackTrace += fmt.Sprintf("\n"+traceFormat, stack.Line, stack.Function, stack.File)
	}

	return fmt.Sprintf("\n Full error information:\n- Cause: %s\n- Info: %s\n- Stack msg: %s\n- Error code: %d\n- Stack trace: %s", e.Cause, e.Message, e.StackMessage, e.Code, stackTrace)
}

// Unwrap returns the original error
func Unwrap(err error) error {

	if err == nil {
		return err
	}

	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err
	}

	return e.Wrapped
}

func GetCause(err error) string {
	if err == nil {
		return ""
	}
	e := extractErr(err)
	if e.Cause == "" && e.Message != "" {
		return e.Message
	} else if e.Cause == "" && e.Message == "" {
		return err.Error()
	}
	return e.Cause
}

func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	e := extractErr(err)
	if e.Message == "" {
		return err.Error()
	}
	return e.Message
}

func GetTrace(err error) ErrTrace {
	if err == nil {
		return ErrTrace{}
	}
	e := extractErr(err)
	if e.Trace.File == "" {
		return ErrTrace{}
	}
	return e.Trace
}

func GetStack(err error) []ErrTrace {
	if err == nil {
		return nil
	}
	e := extractErr(err)
	if len(e.Stack) == 0 {
		return []ErrTrace{}
	}
	return e.Stack
}

func GetStackJson(err error) string {
	if err == nil {
		return ""
	}
	e := extractErr(err)
	if len(e.Stack) == 0 {
		return "{}"
	}

	encoded, _ := json.Marshal(e.Stack)

	return string(encoded)
}

func GetWrapped(err error) error {
	if err == nil {
		return nil
	}
	e := extractErr(err)
	if e.Wrapped == nil {
		return err
	}
	return e.Wrapped
}

func GetCode(err error) int {
	if err == nil {
		return 0
	}
	e := extractErr(err)
	if e.Code == 0 {
		if IsNotFound(err) {
			return 404
		}
		if IsUnauthorized(err) {
			return 403
		}
		if IsBadRequest(err) {
			return 400
		}
		if IsInternal(err) || IsFatal(err) {
			return 500
		}
		if IsConflict(err) {
			return 409
		}
		if IsNoContent(err) {
			return 204
		}
	}
	return e.Code
}

func Trace() ErrTrace {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	function := runtime.FuncForPC(pc[0])
	file, line := function.FileLine(pc[0])

	// matching only the file name
	rgx, err := regexp.Compile(`(?i)/([\w\d_+*()\[\]%=\-]+\.\w+)$`)
	if err == nil {
		matches := rgx.FindStringSubmatch(file)
		if len(matches) > 0 {
			file = matches[1]
		}
	}

	funcName := function.Name()
	rgx, err = regexp.Compile(`(?i)(/[\w\d_*().\-]+$)`)
	if err == nil {
		matches := rgx.FindStringSubmatch(funcName)
		if len(matches) > 0 {
			funcName = matches[1]
		}
	}
	return ErrTrace{Line: line, File: file, Function: funcName}
}
