package errors

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest(Msg("invalid input"))
	assert.NotNil(t, err)
	assert.Equal(t, "invalid input", GetMessage(err))
	assert.Equal(t, 400, GetCode(err))
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound(Msg("not found"))
	assert.NotNil(t, err)
	assert.Equal(t, "not found", GetMessage(err))
	assert.Equal(t, 404, GetCode(err))
}

func TestNewInternal(t *testing.T) {
	err := NewInternal(Msg("internal error"))
	assert.NotNil(t, err)
	assert.Equal(t, "internal error", GetMessage(err))
	assert.Equal(t, 500, GetCode(err))
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized(Msg("unauthorized"))
	assert.NotNil(t, err)
	assert.Equal(t, "unauthorized", GetMessage(err))
	assert.Equal(t, 403, GetCode(err))
}

func TestNewConflict(t *testing.T) {
	err := NewConflict(Msg("conflict"))
	assert.NotNil(t, err)
	assert.Equal(t, "conflict", GetMessage(err))
	assert.Equal(t, 409, GetCode(err))
}

func TestNewFatal(t *testing.T) {
	err := NewFatal(Msg("fatal error"))
	assert.NotNil(t, err)
	assert.Equal(t, "fatal error", GetMessage(err))
	assert.Equal(t, 500, GetCode(err))
}

func TestNewNoContent(t *testing.T) {
	err := NewNoContent(Msg("no content"))
	assert.NotNil(t, err)
	assert.Equal(t, 204, GetCode(err))
}

func TestWithCause(t *testing.T) {
	original := sql.ErrNoRows
	err := NewNotFound(Cause(original), Msg("user not found"))

	assert.Equal(t, "user not found", GetMessage(err))
	assert.Equal(t, original.Error(), GetCause(err))
	assert.Equal(t, original, Unwrap(err))

}

func TestWithMessagef(t *testing.T) {
	userID := 123
	err := NewNotFound(MsgF("user %d not found", userID))

	assert.Equal(t, "user 123 not found", GetMessage(err))
}

func nestedTrace() error {
	return NewInternal(Msg("nested error"), WithTrace())
}

func TestWithTrace(t *testing.T) {
	err := NewInternal(Msg("test"), WithTrace())

	trace := GetTrace(err)
	assert.NotEmpty(t, trace.File)
	assert.NotEmpty(t, trace.Function)
	assert.Greater(t, trace.Line, 0)

	// Testing trace
	errTxt := err.Error()
	if !assert.Contains(t, errTxt, "TestWithTrace") {
		t.Fatalf("trace does not contain function name | got: %s | errorF: %s", errTxt, ErrorF(err))
	}

	// Testing nested trace
	err = nestedTrace()
	nested := Stack(err, Trace())
	nestedTrace := GetStack(nested)
	assert.Len(t, nestedTrace, 2)
	t.Log(ErrorF(nested))
}

func TestWithCode(t *testing.T) {
	err := NewBadRequest(Msg("test"), Code(422))

	assert.Equal(t, 422, GetCode(err))
}

func TestStack(t *testing.T) {
	err := NewInternal(Msg("original"), WithTrace())
	err = Stack(err, Trace())

	stack := GetStack(err)
	assert.Len(t, stack, 2)
}

func TestStackMsg(t *testing.T) {
	err := NewInternal(Msg("original"), WithTrace())
	err = StackMsg(err, "context message", Trace())

	stack := GetStack(err)
	assert.Len(t, stack, 2)
}

func TestIsBadRequest(t *testing.T) {
	err := NewBadRequest(Msg("test"))
	assert.True(t, IsBadRequest(err))
	assert.False(t, IsNotFound(err))
}

func TestIsNotFound(t *testing.T) {
	err := NewNotFound(Msg("test"))
	assert.True(t, IsNotFound(err))
	assert.False(t, IsBadRequest(err))
}

func TestIsInternal(t *testing.T) {
	err := NewInternal(Msg("test"))
	assert.True(t, IsInternal(err))
	assert.False(t, IsNotFound(err))
}

func TestIsUnauthorized(t *testing.T) {
	err := NewUnauthorized(Msg("test"))
	assert.True(t, IsUnauthorized(err))
	assert.False(t, IsNotFound(err))
}

func TestIsConflict(t *testing.T) {
	err := NewConflict(Msg("test"))
	assert.True(t, IsConflict(err))
	assert.False(t, IsNotFound(err))
}

func TestIsFatal(t *testing.T) {
	err := NewFatal(Msg("test"))
	assert.True(t, IsFatal(err))
	assert.False(t, IsNotFound(err))
}

func TestIsNoContent(t *testing.T) {
	err := NewNoContent(Msg("test"))
	assert.True(t, IsNoContent(err))
	assert.False(t, IsNotFound(err))
}

func TestErrorsAs(t *testing.T) {
	err := NewNotFound(Msg("test"))

	var nf *NotFound
	assert.True(t, errors.As(err, &nf))
}

func TestGetStackJson(t *testing.T) {
	err := NewInternal(Msg("test"), WithTrace())

	json := GetStackJson(err)
	assert.NotEmpty(t, json)
	assert.NotEqual(t, "{}", json)
}

func TestErrorF(t *testing.T) {
	err := NewInternal(Cause(sql.ErrNoRows), Msg("test"), WithTrace())

	fullErr := ErrorF(err)
	assert.Contains(t, fullErr, "test")
	assert.Contains(t, fullErr, "sql:")
}

// Concurrent tests
func TestConcurrentStack(t *testing.T) {
	err := NewInternal(Msg("test"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Stack(err, Trace())
		}()
	}
	wg.Wait()

	stack := GetStack(err)
	assert.Len(t, stack, 100)
}

func TestConcurrentStackMsg(t *testing.T) {
	err := NewInternal(Msg("test"))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if index%2 == 0 {
				Stack(err, Trace())
			} else {
				StackMsg(err, "concurrent message", Trace())
			}
		}(i)
	}
	wg.Wait()

	stack := GetStack(err)
	assert.Len(t, stack, 50)
}

func TestConcurrentGetStack(t *testing.T) {
	err := NewInternal(Msg("test"), WithTrace())
	Stack(err, Trace())
	Stack(err, Trace())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stack := GetStack(err)
			assert.Len(t, stack, 3)
		}()
	}
	wg.Wait()
}

// Edge case tests
func TestNilError(t *testing.T) {
	var err error = nil

	assert.Equal(t, "", GetCause(err))
	assert.Equal(t, "", GetMessage(err))
	assert.Equal(t, 0, GetCode(err))
	assert.Empty(t, GetStack(err))
	assert.Equal(t, "", GetStackJson(err))
	assert.Nil(t, GetWrapped(err))
	assert.Equal(t, "", ErrorF(err))
	assert.Equal(t, ErrTrace{}, GetTrace(err))
}

func TestNonTypedError(t *testing.T) {
	err := errors.New("standard error")

	assert.Equal(t, "standard error", GetCause(err))
	assert.Equal(t, "standard error", GetMessage(err))
	assert.Equal(t, 0, GetCode(err))
	assert.Empty(t, GetStack(err))
	assert.Equal(t, "{}", GetStackJson(err))
	assert.Equal(t, err, GetWrapped(err))
	assert.Equal(t, "standard error", ErrorF(err))
}

func TestStackOnNilError(t *testing.T) {
	var err error = nil
	result := Stack(err, Trace())
	assert.Nil(t, result)

	result = StackMsg(err, "message", Trace())
	assert.Nil(t, result)
}

func TestStackOnNonTypedError(t *testing.T) {
	err := errors.New("standard error")
	result := Stack(err, Trace())
	assert.Equal(t, err, result)

	result = StackMsg(err, "message", Trace())
	assert.Equal(t, err, result)
}

func TestEmptyMessage(t *testing.T) {
	err := NewBadRequest()
	assert.Equal(t, "", GetMessage(err))
	assert.Equal(t, 400, GetCode(err))
}

func TestMultipleOptions(t *testing.T) {
	original := sql.ErrNoRows
	err := NewNotFound(
		Cause(original),
		Msg("user not found"),
		Code(410),
		WithTrace(),
	)

	assert.Equal(t, "user not found", GetMessage(err))
	assert.Equal(t, original.Error(), GetCause(err))
	assert.Equal(t, 410, GetCode(err))

	trace := GetTrace(err)
	assert.NotEmpty(t, trace.File)
	assert.Greater(t, trace.Line, 0)
}

func TestGetStackReturnsImmutableCopy(t *testing.T) {
	err := NewInternal(Msg("test"), WithTrace())

	stack1 := GetStack(err)
	stack1[0].Line = 9999 // Modify the copy

	stack2 := GetStack(err)
	assert.NotEqual(t, 9999, stack2[0].Line)
}

func TestErrorOutputWithLongCause(t *testing.T) {
	longCause := errors.New(string(make([]byte, 300)))
	err := NewInternal(Cause(longCause), Msg("test"), WithTrace())

	errStr := err.Error()
	// Error string should contain truncated cause (200 chars + "...") plus message and trace info
	assert.Contains(t, errStr, "...")
	assert.Contains(t, errStr, "test")
}

func TestUnwrapChain(t *testing.T) {
	err1 := errors.New("level 1")
	err2 := NewInternal(Cause(err1), Msg("level 2"))

	unwrapped := Unwrap(err2)
	assert.Equal(t, err1, unwrapped)

	// Test with errors.Is
	assert.True(t, errors.Is(err2, err1))
}

func TestErrorFWithNoStack(t *testing.T) {
	err := NewBadRequest(Msg("test"))
	output := ErrorF(err)

	assert.Contains(t, output, "Error [Code: 400]")
	assert.Contains(t, output, "Message: test")
	assert.NotContains(t, output, "Stack Trace:")
}

func TestErrorFWithStackMessage(t *testing.T) {
	err := NewInternal(Msg("original"))
	err = StackMsg(err, "stack context", Trace())

	output := ErrorF(err)
	assert.Contains(t, output, "Stack:   stack context")
}
