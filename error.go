package errors

import (
	"fmt"
	"regexp"
	"runtime"
)

type Err struct {
	Cause   string     `json:"cause"`
	Message string     `json:"message"`
	Trace   ErrTrace   `json:"trace"`
	Stack   []ErrTrace `json:"stack"`
	Wrapped error      `json:"-"`
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
	}

	return Err{}
}

//	Stack adds a trace to the stack slice
func Stack(err error, trace ErrTrace) error {

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
	}

	return err
}

//	ErrorF returns the full error information
func ErrorF(err error) string {

	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err.Error()
	}

	var stackTrace string

	traceFormat := "> Line=%-15d | Function=%-35s | File=%-30s"

	stackTrace += fmt.Sprintf("\n"+traceFormat, e.Trace.Line, e.Trace.Function, e.Trace.File)

	for _, stack := range e.Stack {
		stackTrace += fmt.Sprintf("\n"+traceFormat, stack.Line, stack.Function, stack.File)
	}

	return fmt.Sprintf("\n Full error information:\n- Cause: %s\n- Info: %s\n- Stack trace: %s",  e.Cause, e.Message, stackTrace)
}

// Unwrap returns the original error
func Unwrap(err error) error {

	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err
	}

	return e.Wrapped
}

func GetCause(err error) string {
	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err.Error()
	}
	return e.Cause
}

func GetMessage(err error) string {
	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err.Error()
	}
	return e.Message
}

func GetTrace(err error) ErrTrace {
	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return ErrTrace{}
	}
	return e.Trace
}

func GetStack(err error) []ErrTrace {
	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return []ErrTrace{}
	}
	return e.Stack
}

func GetWrapped(err error) error {
	e := extractErr(err)
	if e.Cause == "" && e.Message == "" {
		return err
	}
	return e.Wrapped
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
