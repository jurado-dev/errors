package errors

import (
	"database/sql"
	"errors"
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

func TestWithTrace(t *testing.T) {
	err := NewInternal(Msg("test"), WithTrace())

	trace := GetTrace(err)
	assert.NotEmpty(t, trace.File)
	assert.NotEmpty(t, trace.Function)
	assert.Greater(t, trace.Line, 0)
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
