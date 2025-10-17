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
	if e.mu == nil {
		e.mu = &sync.RWMutex{}
	}
	e.mu.Lock()
}

func (e *Err) unlock() {
	if e.mu != nil {
		e.mu.Unlock()
	}
}

func (e *Err) rlock() {
	if e.mu == nil {
		e.mu = &sync.RWMutex{}
	}
	e.mu.RLock()
}

func (e *Err) runlock() {
	if e.mu != nil {
		e.mu.RUnlock()
	}
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

	switch e := err.(type) {
	case *Internal:
		return &e.Err
	case *NotFound:
		return &e.Err
	case *Conflict:
		return &e.Err
	case *BadRequest:
		return &e.Err
	case *Unauthorized:
		return &e.Err
	case *Fatal:
		return &e.Err
	case *NoContent:
		return &e.Err
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
		e.Stack = append(e.Stack, trace)
		e.unlock()
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
		e.Stack = append(e.Stack, trace)
		e.StackMessage = msg
		e.unlock()
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

	// Build the output
	output := fmt.Sprintf("\nError [Code: %d]", e.Code)

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

func Trace() ErrTrace {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	function := runtime.FuncForPC(pc[0])
	file, line := function.FileLine(pc[0])

	// matching only the file name
	matches := fileRegex.FindStringSubmatch(file)
	if len(matches) > 0 {
		file = matches[1]
	}

	funcName := function.Name()
	matches = funcRegex.FindStringSubmatch(funcName)
	if len(matches) > 0 {
		funcName = matches[1]
	}

	return ErrTrace{Line: line, File: file, Function: funcName}
}
