package errors

import (
	stderrors "errors"
	"fmt"
	"sync"
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
	// Initialize mutex immediately to avoid race conditions
	err.mu = &sync.RWMutex{}
	for _, opt := range opts {
		opt(&err)
	}
	// Return by value is safe because mutex is already initialized
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
	return 401
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


type Timeout struct {
	Err
}

func NewTimeout(opts ...Option) error {
	return &Timeout{Err: applyOptions(opts)}
}

func (e *Timeout) GetErr() *Err {
	return &e.Err
}

func (e *Timeout) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 408
}

func (e *Timeout) Unwrap() error {
	return e.Err.Wrapped
}

func IsTimeout(err error) bool {
	var t *Timeout
	return stderrors.As(err, &t)
}

type Forbidden struct {
	Err
}

func NewForbidden(opts ...Option) error {
	return &Forbidden{Err: applyOptions(opts)}
}

func (e *Forbidden) GetErr() *Err {
	return &e.Err
}

func (e *Forbidden) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 403
}

func (e *Forbidden) Unwrap() error {
	return e.Err.Wrapped
}

func IsForbidden(err error) bool {
	var f *Forbidden
	return stderrors.As(err, &f)
}

type MethodNotAllowed struct {
	Err
}

func NewMethodNotAllowed(opts ...Option) error {
	return &MethodNotAllowed{Err: applyOptions(opts)}
}

func (e *MethodNotAllowed) GetErr() *Err {
	return &e.Err
}

func (e *MethodNotAllowed) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 405
}

func (e *MethodNotAllowed) Unwrap() error {
	return e.Err.Wrapped
}

func IsMethodNotAllowed(err error) bool {
	var m *MethodNotAllowed
	return stderrors.As(err, &m)
}

type Gone struct {
	Err
}

func NewGone(opts ...Option) error {
	return &Gone{Err: applyOptions(opts)}
}

func (e *Gone) GetErr() *Err {
	return &e.Err
}

func (e *Gone) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 410
}

func (e *Gone) Unwrap() error {
	return e.Err.Wrapped
}

func IsGone(err error) bool {
	var g *Gone
	return stderrors.As(err, &g)
}

type UnprocessableEntity struct {
	Err
}

func NewUnprocessableEntity(opts ...Option) error {
	return &UnprocessableEntity{Err: applyOptions(opts)}
}

func (e *UnprocessableEntity) GetErr() *Err {
	return &e.Err
}

func (e *UnprocessableEntity) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 422
}

func (e *UnprocessableEntity) Unwrap() error {
	return e.Err.Wrapped
}

func IsUnprocessableEntity(err error) bool {
	var u *UnprocessableEntity
	return stderrors.As(err, &u)
}

type TooManyRequests struct {
	Err
}

func NewTooManyRequests(opts ...Option) error {
	return &TooManyRequests{Err: applyOptions(opts)}
}

func (e *TooManyRequests) GetErr() *Err {
	return &e.Err
}

func (e *TooManyRequests) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 429
}

func (e *TooManyRequests) Unwrap() error {
	return e.Err.Wrapped
}

func IsTooManyRequests(err error) bool {
	var t *TooManyRequests
	return stderrors.As(err, &t)
}

type NotImplemented struct {
	Err
}

func NewNotImplemented(opts ...Option) error {
	return &NotImplemented{Err: applyOptions(opts)}
}

func (e *NotImplemented) GetErr() *Err {
	return &e.Err
}

func (e *NotImplemented) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 501
}

func (e *NotImplemented) Unwrap() error {
	return e.Err.Wrapped
}

func IsNotImplemented(err error) bool {
	var n *NotImplemented
	return stderrors.As(err, &n)
}

type BadGateway struct {
	Err
}

func NewBadGateway(opts ...Option) error {
	return &BadGateway{Err: applyOptions(opts)}
}

func (e *BadGateway) GetErr() *Err {
	return &e.Err
}

func (e *BadGateway) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 502
}

func (e *BadGateway) Unwrap() error {
	return e.Err.Wrapped
}

func IsBadGateway(err error) bool {
	var b *BadGateway
	return stderrors.As(err, &b)
}

type ServiceUnavailable struct {
	Err
}

func NewServiceUnavailable(opts ...Option) error {
	return &ServiceUnavailable{Err: applyOptions(opts)}
}

func (e *ServiceUnavailable) GetErr() *Err {
	return &e.Err
}

func (e *ServiceUnavailable) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 503
}

func (e *ServiceUnavailable) Unwrap() error {
	return e.Err.Wrapped
}

func IsServiceUnavailable(err error) bool {
	var s *ServiceUnavailable
	return stderrors.As(err, &s)
}

type GatewayTimeout struct {
	Err
}

func NewGatewayTimeout(opts ...Option) error {
	return &GatewayTimeout{Err: applyOptions(opts)}
}

func (e *GatewayTimeout) GetErr() *Err {
	return &e.Err
}

func (e *GatewayTimeout) GetCode() int {
	if e.Err.Code != 0 {
		return e.Err.Code
	}
	return 504
}

func (e *GatewayTimeout) Unwrap() error {
	return e.Err.Wrapped
}

func IsGatewayTimeout(err error) bool {
	var g *GatewayTimeout
	return stderrors.As(err, &g)
}