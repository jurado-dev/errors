package errors

func parseFields(fields []interface{}) Err {

	var err Err
	for _, field := range fields {
		if _, ok := field.(error); ok {
			err.Cause = field.(error)
			continue
		}

		if _, ok := field.(string); ok {
			err.Message = field.(string)
			continue
		}

		if _, ok := field.(ErrTrace); ok {
			err.Trace = field.(ErrTrace)
			continue
		}
	}
	return err
}

type BadRequest struct {
	Err
}
func NewBadRequest(fields ...interface{}) *BadRequest {
	return &BadRequest{Err:parseFields(fields)}
}
func IsBadRequest(err error) bool {
	_, ok := err.(*BadRequest)
	return ok
}

type Internal struct{
	Err
}
func NewInternal(fields ...interface{}) *Internal {
	return &Internal{Err:parseFields(fields)}
}
func IsInternal(err error) bool {
	_, ok := err.(*Internal)
	return ok
}

type NotFound struct{
	Err
}
func NewNotFound(fields ...interface{}) *NotFound {
	return &NotFound{Err:parseFields(fields)}
}
func IsNotFound(err error) bool {
	_, ok := err.(*NotFound)
	return ok
}

type Conflict struct{
	Err
}
func NewConflict7(fields ...interface{}) *Conflict {
	return &Conflict{Err:parseFields(fields)}
}
func IsConflict(err error) bool {
	_, ok := err.(*Conflict)
	return ok
}

type Unauthorized struct{
	Err
}
func NewUnauthorized(fields ...interface{}) *Unauthorized {
	return &Unauthorized{Err:parseFields(fields)}
}
func IsUnauthorized(err error) bool {
	_, ok := err.(*Unauthorized)
	return ok
}