# errors

A modern Go error handling library with semantic HTTP-aligned error types, functional options API, automatic stack tracing, and full Go 1.13+ error chain support.

## Installation

```bash
go get github.com/jurado-dev/errors
```

## Features

- **Semantic Error Types**: HTTP-aligned error types (`BadRequest`, `NotFound`, `Unauthorized`, `Internal`, `Conflict`, `Fatal`, `NoContent`)
- **Functional Options API**: Clean, composable error construction
- **Automatic Stack Tracing**: Capture file, function, and line information
- **Error Chain Support**: Full compatibility with Go 1.13+ `errors.Is()` and `errors.As()`
- **Thread-Safe**: All operations protected with mutexes
- **Type Checking**: Simple `Is*` functions for error type validation
- **HTTP Code Mapping**: Automatic HTTP status code resolution

## Quick Start

### Creating Errors

```go
import "github.com/jurado-dev/errors"

// Simple error with message
err := errors.NewBadRequest(errors.Msg("invalid user input"))

// Error with cause and trace
err := errors.NewNotFound(
    errors.Cause(dbErr),
    errors.Msg("user not found"),
    errors.WithTrace(),
)

// Error with formatted message
err := errors.NewInternal(
    errors.Cause(dbErr),
    errors.Messagef("failed to query user %s", userID),
    errors.WithTrace(),
)

// Error with custom HTTP code
err := errors.NewBadRequest(
    errors.Msg("invalid format"),
    errors.Code(422),
)

// All options are chainable and optional
err := errors.NewConflict(
    errors.Cause(originalErr),
    errors.Msg("email already exists"),
    errors.Code(409),
    errors.WithTrace(),
)
```

### Error Types

| Type | Default HTTP Code | Use Case |
|------|-------------------|----------|
| `BadRequest` | 400 | Invalid client input |
| `Unauthorized` | 403 | Authentication/authorization failures |
| `NotFound` | 404 | Resource not found |
| `Conflict` | 409 | Resource conflicts (e.g., duplicate key) |
| `Internal` | 500 | Internal server errors |
| `Fatal` | 500 | Critical failures |
| `NoContent` | 204 | Successful operation with no content |

### Functional Options

| Option | Description | Example |
|--------|-------------|---------|
| `Cause(err)` | Wrap an underlying error | `errors.Cause(dbErr)` |
| `Msg(string)` | Set user-friendly message | `errors.Msg("user not found")` |
| `Messagef(fmt, args...)` | Set formatted message | `errors.Messagef("user %s not found", id)` |
| `WithTrace()` | Add stack trace | `errors.WithTrace()` |
| `Code(int)` | Override HTTP status code | `errors.Code(422)` |

### Type Checking

```go
// Using package-level type checkers (recommended)
if errors.IsBadRequest(err) {
    // Handle bad request
}

if errors.IsNotFound(err) {
    // Handle not found
}

// Using Go 1.13+ errors.As (also works)
var notFoundErr *errors.NotFound
if errors.As(err, &notFoundErr) {
    // Handle not found
}
```

### Adding Stack Traces

```go
// Propagating error with stack trace
func processUser(id string) error {
    user, err := userService.GetUser(id)
    if err != nil {
        return errors.Stack(err, errors.Trace())
    }
    return nil
}

// With additional context message
func processOrder(id string) error {
    order, err := orderService.GetOrder(id)
    if err != nil {
        return errors.StackMsg(err, "failed during order processing", errors.Trace())
    }
    return nil
}
```

### Extracting Error Information

```go
// Get the wrapped error cause
cause := errors.GetCause(err)

// Get the user-friendly message
msg := errors.GetMessage(err)

// Get HTTP status code
code := errors.GetCode(err)  // Returns 404, 500, etc.

// Get the trace information
trace := errors.GetTrace(err)

// Get full stack trace
stack := errors.GetStack(err)

// Get stack trace as JSON string
jsonStack := errors.GetStackJson(err)

// Get the wrapped underlying error
wrapped := errors.GetWrapped(err)
```

### Full Error Details

```go
// Print complete error information including stack trace
fmt.Println(errors.ErrorF(err))

// Example Output:
// Error [Code: 404]
//   Cause:   sql: no rows in result set
//   Message: user not found
//   Stack:   failed during user lookup
//
// Stack Trace:
//   1. user_service.go:45 in /services.GetUserByID
//   2. handlers.go:102 in /handlers.UserHandler
```

### Error Output Format

The `Error()` method provides clean, readable output:

```go
err := errors.NewNotFound(
    errors.Cause(sql.ErrNoRows),
    errors.Msg("user not found"),
    errors.WithTrace(),
)

fmt.Println(err)
// Output: sql: no rows in result set: user not found (at /services.GetUserByID:45)

// Without cause:
err := errors.NewBadRequest(errors.Msg("invalid input"), errors.WithTrace())
fmt.Println(err)
// Output: invalid input (at /api.CreateUser:28)

// Without trace:
err := errors.NewBadRequest(errors.Msg("invalid input"))
fmt.Println(err)
// Output: invalid input
```

## Common Patterns

### HTTP Handler

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := userService.GetUser(r.Context(), id)
    if err != nil {
        statusCode := errors.GetCode(err)
        message := errors.GetMessage(err)
        
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(map[string]string{
            "error": message,
        })
        
        // Log full error details including stack trace
        log.Error(errors.ErrorF(err))
        return
    }
    
    json.NewEncoder(w).Encode(user)
}
```

### Service Layer

```go
func (s *UserService) CreateUser(ctx context.Context, user User) error {
    // Check if user exists
    existing, err := s.repo.FindByEmail(user.Email)
    if err != nil && !errors.IsNotFound(err) {
        return errors.Stack(err, errors.Trace())
    }
    
    if existing != nil {
        return errors.NewConflict(
            errors.Msg("user with this email already exists"),
            errors.WithTrace(),
        )
    }
    
    // Create user
    if err := s.repo.Create(user); err != nil {
        return errors.Stack(err, errors.Trace())
    }
    
    return nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.StackMsg(err, "failed to retrieve user", errors.Trace())
    }
    return user, nil
}
```

### Repository Layer

```go
func (r *UserRepository) FindByID(id string) (*User, error) {
    var user User
    err := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    
    if err == sql.ErrNoRows {
        return nil, errors.NewNotFound(
            errors.Cause(err),
            errors.Msg("user not found"),
            errors.WithTrace(),
        )
    }
    
    if err != nil {
        return nil, errors.NewInternal(
            errors.Cause(err),
            errors.Msg("database query failed"),
            errors.WithTrace(),
        )
    }
    
    return &user, nil
}

func (r *UserRepository) Create(user User) error {
    _, err := r.db.Exec("INSERT INTO users (...) VALUES (...)", user.Fields...)
    
    if err != nil {
        if isDuplicateKeyError(err) {
            return errors.NewConflict(
                errors.Cause(err),
                errors.Msg("user already exists"),
                errors.WithTrace(),
            )
        }
        return errors.NewInternal(
            errors.Cause(err),
            errors.Msg("failed to insert user"),
            errors.WithTrace(),
        )
    }
    
    return nil
}
```

### Input Validation

```go
func (s *UserService) ValidateUser(user User) error {
    if user.Email == "" {
        return errors.NewBadRequest(
            errors.Msg("email is required"),
        )
    }
    
    if !isValidEmail(user.Email) {
        return errors.NewBadRequest(
            errors.Messagef("invalid email format: %s", user.Email),
        )
    }
    
    if len(user.Password) < 8 {
        return errors.NewBadRequest(
            errors.Msg("password must be at least 8 characters"),
            errors.Code(422),
        )
    }
    
    return nil
}
```

### Cross-Service Communication

```go
func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
    // Get user from internal service
    user, err := s.userService.GetUser(ctx, userID)
    if err != nil {
        return errors.StackMsg(err, "failed to fetch user for order", errors.Trace())
    }
    
    // Call external payment API
    payment, err := s.paymentClient.Charge(amount)
    if err != nil {
        return errors.NewInternal(
            errors.Cause(err),
            errors.Msg("payment processing failed"),
            errors.WithTrace(),
        )
    }
    
    return nil
}
```

## Error Wrapping & Unwrapping

This package fully supports Go 1.13+ error unwrapping:

```go
// Create error with cause
dbErr := sql.ErrNoRows
err := errors.NewNotFound(
    errors.Cause(dbErr),
    errors.Msg("user not found"),
)

// Using standard library errors.Is
if errors.Is(err, sql.ErrNoRows) {
    // Handle sql.ErrNoRows
}

// Using standard library errors.As
var notFoundErr *errors.NotFound
if errors.As(err, &notFoundErr) {
    // Access NotFound-specific methods
    code := notFoundErr.GetCode()
}

// Using package functions
original := errors.Unwrap(err)
fmt.Println(original)  // sql: no rows in result set

wrapped := errors.GetWrapped(err)
```

## API Reference

### Error Constructors

All constructors accept functional options and return `error` interface:

- `NewBadRequest(...ErrOption) error`
- `NewUnauthorized(...ErrOption) error`
- `NewNotFound(...ErrOption) error`
- `NewConflict(...ErrOption) error`
- `NewInternal(...ErrOption) error`
- `NewFatal(...ErrOption) error`
- `NewNoContent(...ErrOption) error`

### Functional Options

- `Cause(error) ErrOption` - Wrap an underlying error
- `Msg(string) ErrOption` - Set user-friendly message
- `Messagef(string, ...interface{}) ErrOption` - Set formatted message
- `WithTrace() ErrOption` - Add stack trace
- `Code(int) ErrOption` - Override HTTP status code

### Type Checking

- `IsBadRequest(error) bool`
- `IsUnauthorized(error) bool`
- `IsNotFound(error) bool`
- `IsConflict(error) bool`
- `IsInternal(error) bool`
- `IsFatal(error) bool`
- `IsNoContent(error) bool`

### Stack Tracing

- `Trace() ErrTrace` - Capture current file/line/function
- `Stack(error, ErrTrace) error` - Add trace to stack
- `StackMsg(error, string, ErrTrace) error` - Add trace with message

### Information Extraction

- `GetCause(error) string` - Get wrapped error message
- `GetMessage(error) string` - Get user-friendly message
- `GetCode(error) int` - Get HTTP status code
- `GetTrace(error) ErrTrace` - Get initial trace
- `GetStack(error) []ErrTrace` - Get full stack trace (thread-safe copy)
- `GetStackJson(error) string` - Get stack as JSON
- `GetWrapped(error) error` - Get wrapped error
- `ErrorF(error) string` - Get formatted full error details
- `Unwrap(error) error` - Unwrap to original error (Go 1.13+ compatible)

## Thread Safety

All operations in this package are thread-safe:

- Error construction uses functional options (immutable)
- Stack operations use mutexes to protect concurrent access
- `GetStack()` returns a copy to prevent external modification

## License

MIT
