package errors

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
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

func Trace() ErrTrace {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	if n < 2 {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}

	// Get the immediate caller of Trace()
	function := runtime.FuncForPC(pc[0])
	if function == nil {
		return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
	}

	// Check if the caller is WithTrace() - if so, skip it and check the next caller
	var file string
	var line int
	var funcName string
	if strings.Contains(function.Name(), "WithTrace") && n >= 3 {
		// Get the caller of WithTrace()
		function2 := runtime.FuncForPC(pc[1])
		if function2 != nil && strings.Contains(function2.Name(), "applyOptions") && n >= 4 {
			// Skip applyOptions() as well, use the caller of applyOptions()
			function = runtime.FuncForPC(pc[2])
			if function == nil {
				return ErrTrace{Line: 0, File: "unknown", Function: "unknown"}
			}
			file, line = function.FileLine(pc[2])
			funcName = function.Name()
		} else {
			// Normal case: use the caller of WithTrace()
			file, line = function2.FileLine(pc[1])
			funcName = function2.Name()
		}
	} else {
		// Normal case: use the immediate caller
		file, line = function.FileLine(pc[0])
		funcName = function.Name()
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
