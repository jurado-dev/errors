package errors

import (
	stderrors "errors"
	"fmt"
)

// Option is a functional option for error construction
type Option func(*Err)

// Cause sets the wrapped error
func Cause(err error) Option {
	return func(e *Err) {
		if err != nil {
			e.Cause = err.Error()
			e.Wrapped = err
		}
	}
}

// Msg sets the user-friendly message
func Msg(msg string) Option {
	return func(e *Err) {
		e.Message = msg
	}
}

// MsgF sets a formatted user-friendly message
func MsgF(format string, args ...interface{}) Option {
	return func(e *Err) {
		e.Message = fmt.Sprintf(format, args...)
	}
}

// WithTrace sets the error trace
func WithTrace() Option {
	return func(e *Err) {
		trace := Trace()
		e.Trace = trace
		e.Stack = append(e.Stack, trace)
	}
}

// Code sets a custom HTTP status code
func Code(code int) Option {
	return func(e *Err) {
		e.Code = code
	}
}

func applyOptions(opts []Option) Err {
	var err Err
	for _, opt := range opts {
		opt(&err)
	}
	// Return by value is safe because mutex is zero-value (unlocked)
	return err
}

type BadRequest struct {
	Err
}

func NewBadRequest(opts ...Option) error {
	return &BadRequest{Err: applyOptions(opts)}
}

func (e *BadRequest) GetErr() *Err {
	return &e.Err
}

func (e *BadRequest) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 400
}

func (e *BadRequest) Unwrap() error {
	return e.Err.Wrapped
}

func IsBadRequest(err error) bool {
	var br *BadRequest
	return stderrors.As(err, &br)
}

type Internal struct {
	Err
}

func NewInternal(opts ...Option) error {
	return &Internal{Err: applyOptions(opts)}
}

func (e *Internal) GetErr() *Err {
	return &e.Err
}

func (e *Internal) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 500
}

func (e *Internal) Unwrap() error {
	return e.Err.Wrapped
}

func IsInternal(err error) bool {
	var ie *Internal
	return stderrors.As(err, &ie)
}

type NotFound struct {
	Err
}

func NewNotFound(opts ...Option) error {
	return &NotFound{Err: applyOptions(opts)}
}

func (e *NotFound) GetErr() *Err {
	return &e.Err
}

func (e *NotFound) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 404
}

func (e *NotFound) Unwrap() error {
	return e.Err.Wrapped
}

func IsNotFound(err error) bool {
	var nf *NotFound
	return stderrors.As(err, &nf)
}

type Conflict struct {
	Err
}

func NewConflict(opts ...Option) error {
	return &Conflict{Err: applyOptions(opts)}
}

func (e *Conflict) GetErr() *Err {
	return &e.Err
}

func (e *Conflict) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 409
}

func (e *Conflict) Unwrap() error {
	return e.Err.Wrapped
}

func IsConflict(err error) bool {
	var c *Conflict
	return stderrors.As(err, &c)
}

type Unauthorized struct {
	Err
}

func NewUnauthorized(opts ...Option) error {
	return &Unauthorized{Err: applyOptions(opts)}
}

func (e *Unauthorized) GetErr() *Err {
	return &e.Err
}

func (e *Unauthorized) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 403
}

func (e *Unauthorized) Unwrap() error {
	return e.Err.Wrapped
}

func IsUnauthorized(err error) bool {
	var ua *Unauthorized
	return stderrors.As(err, &ua)
}

type Fatal struct {
	Err
}

func NewFatal(opts ...Option) error {
	return &Fatal{Err: applyOptions(opts)}
}

func (e *Fatal) GetErr() *Err {
	return &e.Err
}

func (e *Fatal) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 500
}

func (e *Fatal) Unwrap() error {
	return e.Err.Wrapped
}

func IsFatal(err error) bool {
	var f *Fatal
	return stderrors.As(err, &f)
}

type NoContent struct {
	Err
}

func NewNoContent(opts ...Option) error {
	return &NoContent{Err: applyOptions(opts)}
}

func (e *NoContent) GetErr() *Err {
	return &e.Err
}

func (e *NoContent) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 204
}

func (e *NoContent) Unwrap() error {
	return e.Err.Wrapped
}

func IsNoContent(err error) bool {
	var nc *NoContent
	return stderrors.As(err, &nc)
}
