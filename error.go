package errors

import (
	"encoding/json"
	"regexp"
	"runtime"
)

type Err struct {
	Cause   error    `json:"cause"`
	Message string   `json:"message"`
	Trace   ErrTrace `json:"trace"`
}

type ErrTrace struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

func (e *Err) Error() string {
	encoded, _ := json.Marshal(e)
	return string(encoded)
}

func Cause (err error) Err {
	return Err{Cause: err}
}

func (e *Err) GetCause() error {
	return e.Cause
}

func (e *Err) GetMessage() string {
	return e.Message
}

func (e *Err) GetTrace() ErrTrace {
	return e.Trace
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
