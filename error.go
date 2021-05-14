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

func (e *Err) Error() string {

	output := fmt.Sprintf("Info: %s", e.Message)
	if e.Trace.Line != 0 {

		cause := e.Cause
		if len(cause) > 100 {
			cause = cause[:100] + "..."
		}

		output = fmt.Sprintf("Cause: %s | Info: %s | Line: %d | Function: %s", cause, e.Message, e.Trace.Line, e.Trace.Function)
	}

	return output
}

func Stack(err error, trace ErrTrace) error {
	e, ok := err.(*Err)
	if !ok {
		return err
	}

	e.Stack = append(e.Stack, trace)

	return e
}

func ErrorF(err error) string {
	e, ok := err.(*Err)
	if !ok {
		return err.Error()
	}

	var stackTrace string

	traceFormat := "Line: %d | Function: %s | File: %s"

	stackTrace += fmt.Sprintf("\n"+traceFormat, e.Trace.Line, e.Trace.Function, e.Trace.File)

	for _, stack := range e.Stack {
		stackTrace += fmt.Sprintf("\n"+traceFormat, stack.Line, stack.Function, stack.File)
	}

	return fmt.Sprintf("Cause: %s\nInfo: %s\nStack trace: %s",  e.Cause, e.Message, stackTrace)
}

func Unwrap(err error) error {

	e, ok := err.(*Err)
	if !ok {
		return err
	}

	return e.Wrapped
}

func (e *Err) GetCause() string {
	return e.Cause
}

func (e *Err) GetMessage() string {
	return e.Message
}

func (e *Err) GetTrace() ErrTrace {
	return e.Trace
}

func (e *Err) GetStack() []ErrTrace {
	return e.Stack
}

func (e *Err) GetWrapped() error {
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
