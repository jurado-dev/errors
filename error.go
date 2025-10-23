package errors

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"sync"
)

const (
	maxCauseLength = 200
)

var (
	fileRegex = regexp.MustCompile(`(?i)/([\w\d_+*()\[\]%=\-]+\.\w+)$`)
	funcRegex = regexp.MustCompile(`(?i)(/[\w\d_*().\-]+$)`)
)

type Err struct {
	Cause        string     `json:"cause"`
	Message      string     `json:"message"`
	StackMessage string     `json:"stack_message"`
	Trace        ErrTrace   `json:"trace"`
	Stack        []ErrTrace `json:"stack"`
	Wrapped      error      `json:"-"`
	Code         int        `json:"code"`
	mu           *sync.RWMutex
}

func (e *Err) lock() {
	e.mu.Lock()
}

func (e *Err) unlock() {
	e.mu.Unlock()
}

func (e *Err) rlock() {
	e.mu.RLock()
}

func (e *Err) runlock() {
	e.mu.RUnlock()
}

type ErrTrace struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// TypedError interface for all error types in this package
type TypedError interface {
	error
	GetErr() *Err
	GetCode() int
}

// Error implements the interface
func (e *Err) Error() string {
	// Simple format when no trace is available
	if e.Trace.Line == 0 {
		return e.Message
	}

	// Full format with trace information
	cause := e.Cause
	if len(cause) > maxCauseLength {
		cause = cause[:maxCauseLength] + "..."
	}

	// If no cause is set, omit it from output
	if cause == "" {
		return fmt.Sprintf("%s (at %s:%d)", e.Message, e.Trace.Function, e.Trace.Line)
	}

	return fmt.Sprintf("%s: %s (at %s:%d)", cause, e.Message, e.Trace.Function, e.Trace.Line)
}

func extractErr(err error) *Err {
	if te, ok := err.(TypedError); ok {
		return te.GetErr()
	}
	return nil
}

// Stack adds a trace to the stack slice (thread-safe)
func Stack(err error, trace ErrTrace) error {
	if err == nil {
		return err
	}

	if te, ok := err.(TypedError); ok {
		e := te.GetErr()
		e.lock()
		defer e.unlock()
		e.Stack = append(e.Stack, trace)
		return err
	}

	return err
}

func StackMsg(err error, msg string, trace ErrTrace) error {
	if err == nil {
		return err
	}

	if te, ok := err.(TypedError); ok {
		e := te.GetErr()
		e.lock()
		defer e.unlock()
		e.Stack = append(e.Stack, trace)
		e.StackMessage = msg
		return err
	}

	return err
}

// ErrorF returns the full error information with detailed stack trace
func ErrorF(err error) string {
	if err == nil {
		return ""
	}

	e := extractErr(err)
	if e == nil || (e.Cause == "" && e.Message == "") {
		return err.Error()
	}

	e.rlock()
	defer e.runlock()

	// Get the actual HTTP code (handles default codes for each type)
	code := GetCode(err)

	// Build the output
	output := fmt.Sprintf("\nError [Code: %d]", code)

	if e.Cause != "" {
		output += fmt.Sprintf("\n  Cause:   %s", e.Cause)
	}

	if e.Message != "" {
		output += fmt.Sprintf("\n  Message: %s", e.Message)
	}

	if e.StackMessage != "" {
		output += fmt.Sprintf("\n  Stack:   %s", e.StackMessage)
	}

	// Add stack trace if available
	if len(e.Stack) > 0 {
		output += "\n\nStack Trace:"
		for i, trace := range e.Stack {
			output += fmt.Sprintf("\n  %d. %s:%d in %s", i+1, trace.File, trace.Line, trace.Function)
		}
	}

	return output
}

// Unwrap returns the original error (supports standard errors.Unwrap)
func Unwrap(err error) error {

	if err == nil {
		return err
	}

	e := extractErr(err)
	if e == nil || (e.Cause == "" && e.Message == "") {
		return err
	}

	return e.Wrapped
}

func GetCause(err error) string {
	if err == nil {
		return ""
	}
	e := extractErr(err)
	if e == nil {
		return err.Error()
	}
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
	if e == nil || e.Message == "" {
		return err.Error()
	}
	return e.Message
}

func GetTrace(err error) ErrTrace {
	if err == nil {
		return ErrTrace{}
	}
	e := extractErr(err)
	if e == nil || e.Trace.File == "" {
		return ErrTrace{}
	}
	return e.Trace
}

func GetStack(err error) []ErrTrace {
	if err == nil {
		return []ErrTrace{}
	}
	e := extractErr(err)
	if e == nil {
		return []ErrTrace{}
	}

	e.rlock()
	defer e.runlock()

	if len(e.Stack) == 0 {
		return []ErrTrace{}
	}

	// Return a copy to prevent external modification
	stackCopy := make([]ErrTrace, len(e.Stack))
	copy(stackCopy, e.Stack)
	return stackCopy
}

func GetStackJson(err error) string {
	if err == nil {
		return ""
	}
	e := extractErr(err)
	if e == nil {
		return "{}"
	}

	e.rlock()
	defer e.runlock()

	if len(e.Stack) == 0 {
		return "{}"
	}

	encoded, jsonErr := json.Marshal(e.Stack)
	if jsonErr != nil {
		return "{}"
	}

	return string(encoded)
}

func GetWrapped(err error) error {
	if err == nil {
		return nil
	}
	e := extractErr(err)
	if e == nil || e.Wrapped == nil {
		return err
	}
	return e.Wrapped
}

func GetCode(err error) int {
	if err == nil {
		return 0
	}

	if te, ok := err.(TypedError); ok {
		return te.GetCode()
	}

	return 0
}

var applyOptionsPattern = regexp.MustCompile(`(?i)errors\.applyOptions`)
var traceFunPattern = regexp.MustCompile(`(?i)errors\.Trace`)

func Trace() ErrTrace {
	pc := make([]uintptr, 100)
	n := runtime.Callers(0, pc)
	if n < 2 {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}
	var file string
	var line int
	var funcName string

	// Identify the index of the last call to trace and applyOptions
	/*
		Example stack trace when using WithTrace:
		github.com/jurado-dev/errors.Trace
		github.com/jurado-dev/errors.Trace
		github.com/jurado-dev/errors.nestedTrace.WithTrace.func2
		github.com/jurado-dev/errors.applyOptions
		github.com/jurado-dev/errors.NewInternal
		github.com/jurado-dev/errors.nestedTrace // <- first caller outside errors package (applyOptionsIdx + 2)
		github.com/jurado-dev/errors.TestWithTrace
		testing.tRunner
		runtime.goexit

		Example stack trace when using Trace directly:
		github.com/jurado-dev/errors.Trace
		github.com/jurado-dev/errors.Trace
		github.com/jurado-dev/errors.TestWithTrace // <- first caller outside errors package (traceIdx + 1)
		testing.tRunner
		runtime.goexit
	*/
	var applyOptionsIdx int = -1
	var traceIdx int
	pc = pc[:n]
	for i, pcItem := range pc {
		function := runtime.FuncForPC(pcItem)
		if function == nil {
			continue
		}
		if applyOptionsPattern.MatchString(function.Name()) {
			applyOptionsIdx = i
		}
		if traceFunPattern.MatchString(function.Name()) {
			traceIdx = i
		}
	}
	if len(pc) == 0 {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}

	var firstCall uintptr
	if traceIdx > 0 && traceIdx < len(pc) && applyOptionsIdx == -1 {
		callIdx := traceIdx + 1 // Trace -> caller
		if callIdx < len(pc) {
			firstCall = pc[callIdx]
		}
	} else if applyOptionsIdx > 0 && applyOptionsIdx < len(pc) {
		callIdx := applyOptionsIdx + 2 // applyOptions -> NewXxx -> caller
		if callIdx < len(pc) {
			firstCall = pc[callIdx]
		}
	}
	if firstCall == 0 {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}

	function := runtime.FuncForPC(firstCall)
	file, line = function.FileLine(firstCall)
	funcName = function.Name()

	if file == "" {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}

	// Apply regex to shorten file and function names as before
	matches := fileRegex.FindStringSubmatch(file)
	if len(matches) > 0 {
		file = matches[1]
	}

	matches = funcRegex.FindStringSubmatch(funcName)
	if len(matches) > 0 {
		funcName = matches[1]
	}

	return ErrTrace{Line: line, File: file, Function: funcName}
}
