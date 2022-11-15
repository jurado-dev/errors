package errors

func parseFields(fields []interface{}) Err {

	var err Err
	for _, field := range fields {

		if e, ok := field.(error); ok {
			err.Cause = e.Error()
			err.Wrapped = e
		}

		if _, ok := field.(string); ok {
			err.Message = field.(string)
			continue
		}

		if _, ok := field.(ErrTrace); ok {
			err.Trace = field.(ErrTrace)
			err.Stack = append(err.Stack, field.(ErrTrace))
			continue
		}

		if _, ok := field.(int); ok {
			err.Code = field.(int)
			continue
		}
	}
	return err
}

type BadRequest struct {
	Err
}
func NewBadRequest(fields ...interface{}) *BadRequest {
	e := &BadRequest{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 400
	}
	return e
}
func IsBadRequest(err error) bool {
	_, ok := err.(*BadRequest)
	return ok
}

type Internal struct {
	Err
}
func NewInternal(fields ...interface{}) *Internal {
	e := &Internal{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 500
	}
	return e
}
func IsInternal(err error) bool {
	_, ok := err.(*Internal)
	return ok
}

type NotFound struct {
	Err
}
func NewNotFound(fields ...interface{}) *NotFound {
	e := &NotFound{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 404
	}
	return e
}
func IsNotFound(err error) bool {
	_, ok := err.(*NotFound)
	return ok
}

type Conflict struct {
	Err
}
func NewConflict(fields ...interface{}) *Conflict {
	e := &Conflict{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 409
	}
	return e
}
func IsConflict(err error) bool {
	_, ok := err.(*Conflict)
	return ok
}

type Unauthorized struct {
	Err
}
func NewUnauthorized(fields ...interface{}) *Unauthorized {
	e := &Unauthorized{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 403
	}
	return e
}
func IsUnauthorized(err error) bool {
	_, ok := err.(*Unauthorized)
	return ok
}

type Fatal struct {
	Err
}
func NewFatal(fields ...interface{}) *Fatal {
	e := &Fatal{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 500
	}
	return e
}
func IsFatal(err error) bool {
	_, ok := err.(*Fatal)
	return ok
}

type NoContent struct {
	Err
}
func NewNoContent(fields ...interface{}) *NoContent {
	e := &NoContent{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 204
	}
	return e
}
func IsNoContent(err error) bool {
	_, ok := err.(*NoContent)
	return ok
}

type Timeout struct {
	Err
}
func NewTimeout(fields ...interface{}) *Timeout {
	e := &Timeout{Err: parseFields(fields)}
	if e.Err.Code <= 0 {
		e.Err.Code = 408
	}
	return e
}
func IsTimeout(err error) bool {
	_, ok := err.(*Timeout)
	return ok
}