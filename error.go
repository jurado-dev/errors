package errors

import (
	"fmt"
	"regexp"
	"runtime"
)

type Err struct {
	Cause   string   `json:"cause"`
	Message string   `json:"message"`
	Trace   ErrTrace `json:"trace"`
	Wrapped error    `json:"-"`
}

type ErrTrace struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

func (e *Err) Error() string {

	output := fmt.Sprintf("Info: %s", e.Message)
	if e.Trace.Line != 0 {
		output = fmt.Sprintf("Info: %s | Line: %d | Function: %s", e.Message, e.Trace.Line, e.Trace.Function)
	}

	return output
}

func ErrorF(err error) string {
	e, ok := err.(*Err)
	if !ok {
		return err.Error()
	}

	return fmt.Sprintf("Info: %s | Cause: %s | Line: %d | Function: %s | File: %s", e.Message, e.Cause, e.Trace.Line, e.Trace.Function, e.Trace.File)
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
	rgx, err = regexp.Compile(`(?i)(/[\w\d_\-]+/[\w\d_*().\-]+$)`)
	if err == nil {
		matches := rgx.FindStringSubmatch(funcName)
		if len(matches) > 0 {
			funcName = matches[1]
		}
	}
	return ErrTrace{Line: line, File: file, Function: funcName}
}
