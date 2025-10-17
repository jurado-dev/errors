# errors

A Go error handling library with semantic HTTP-aligned error types, automatic stack tracing, and error wrapping capabilities.

## Installation

```bash
go get github.com/jurado-dev/errors
```

## Features

- **Semantic Error Types**: HTTP-aligned error types (`BadRequest`, `NotFound`, `Unauthorized`, `Internal`, `Conflict`, `Fatal`, `NoContent`)
- **Automatic Stack Tracing**: Capture file, function, and line information
- **Error Wrapping**: Wrap underlying errors while adding context
- **Flexible Construction**: Pass error details in any order
- **Type Checking**: Simple `Is*` functions for error type validation
- **HTTP Code Mapping**: Automatic HTTP status code resolution

## Quick Start

### Creating Errors

```go
import "github.com/jurado-dev/errors"

// Simple error with message
err := errors.NewBadRequest("invalid user input")

// Error with wrapped error and trace
err := errors.NewInternal(dbErr, "database query failed", errors.Trace())

// Error with custom HTTP code
err := errors.NewBadRequest("invalid format", 422)

// All parameters are optional and order-independent
err := errors.NewNotFound(errors.Trace(), originalErr, "user not found")
```

### ⚠️ Important: Error Wrapping Rules

**Use `New*()` constructors ONLY for errors from external packages:**

```go
// ✅ CORRECT: Wrapping standard library or third-party errors
dbErr := sql.ErrNoRows  // External package error
return errors.NewNotFound(dbErr, "user not found", errors.Trace())
```

**Use `Stack()` or `StackMsg()` for errors already using this package:**

```go
// ✅ CORRECT: Adding trace to errors from your own services
err := userService.GetUser(id)  // Already returns errors.* type
if err != nil {
    return errors.Stack(err, errors.Trace())
}

// ❌ WRONG: Don't wrap errors that are already wrapped
err := userService.GetUser(id)
if err != nil {
    return errors.NewInternal(err, "failed", errors.Trace())  // DON'T DO THIS!
}
```

**Why this matters:** Double-wrapping loses the original error type and breaks type checking with `Is*()` functions.

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

### Type Checking

```go
if errors.IsBadRequest(err) {
    // Handle bad request
}

if errors.IsNotFound(err) {
    // Handle not found
}
```

### Adding Stack Traces

**Always use `Stack()` or `StackMsg()` when propagating errors from methods that already use this package:**

```go
// Propagating error from your own service/repository
func processUser(id string) error {
    user, err := userService.GetUser(id)  // Returns errors.* type
    if err != nil {
        // ✅ Add current location to existing error's stack
        return errors.Stack(err, errors.Trace())
    }
    return nil
}

// With additional context message
func processOrder(id string) error {
    order, err := orderService.GetOrder(id)  // Returns errors.* type
    if err != nil {
        // ✅ Add trace with contextual message
        return errors.StackMsg(err, "failed during order processing", errors.Trace())
    }
    return nil
}

// Wrapping external library errors
func getUserFromDB(id string) error {
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    if err == sql.ErrNoRows {
        // ✅ External error: use New* constructor
        return errors.NewNotFound(err, "user not found", errors.Trace())
    }
    if err != nil {
        // ✅ External error: use New* constructor
        return errors.NewInternal(err, "database query failed", errors.Trace())
    }
    return nil
}
```
```

### Extracting Error Information

```go
// Get the original cause message
cause := errors.GetCause(err)

// Get the user-friendly message
msg := errors.GetMessage(err)

// Get HTTP status code
code := errors.GetCode(err)

// Get the initial trace
trace := errors.GetTrace(err)

// Get full stack trace
stack := errors.GetStack(err)

// Get stack trace as JSON string
jsonStack := errors.GetStackJson(err)

// Get the wrapped underlying error
original := errors.GetWrapped(err)
```

### Full Error Details

```go
// Print complete error information including stack trace
fmt.Println(errors.ErrorF(err))

// Output:
// Full error information:
// - Cause: sql: no rows in result set
// - Info: user not found
// - Stack msg: failed during user lookup
// - Error code: 404
// - Stack trace:
// > Line=45   | Function=/services.GetUserByID              | File=user_service.go
// > Line=102  | Function=/handlers.UserHandler              | File=handlers.go
```

### Error Wrapping & Unwrapping

```go
// Wrap an error with context
dbErr := sql.ErrNoRows
err := errors.NewNotFound(dbErr, "user not found", errors.Trace())

// Unwrap to get the original error
original := errors.Unwrap(err)
fmt.Println(original) // sql: no rows in result set

// Or use GetWrapped
wrapped := errors.GetWrapped(err)
```

## Common Patterns

### HTTP Handler

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := userService.GetUser(r.Context(), id)
    if err != nil {
        // Error already typed from service layer
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
    // Calling repository (uses this errors package)
    existing, err := s.repo.FindByEmail(user.Email)
    if err != nil && !errors.IsNotFound(err) {
        // ✅ Repository returns errors.* type: use Stack()
        return errors.Stack(err, errors.Trace())
    }
    
    if existing != nil {
        // ✅ New error originating here: use New*()
        return errors.NewConflict("user with this email already exists", errors.Trace())
    }
    
    // Calling repository (uses this errors package)
    if err := s.repo.Create(user); err != nil {
        // ✅ Repository returns errors.* type: use Stack()
        return errors.Stack(err, errors.Trace())
    }
    
    return nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Calling repository (uses this errors package)
    user, err := s.repo.FindByID(id)
    if err != nil {
        // ✅ Add context to existing error with StackMsg()
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
    
    // ✅ External package errors: wrap with New*()
    if err == sql.ErrNoRows {
        return nil, errors.NewNotFound(err, "user not found", errors.Trace())
    }
    if err != nil {
        return nil, errors.NewInternal(err, "database query failed", errors.Trace())
    }
    
    return &user, nil
}

func (r *UserRepository) Create(user User) error {
    _, err := r.db.Exec("INSERT INTO users (...) VALUES (...)", user.Fields...)
    
    // ✅ External package errors: wrap with New*()
    if err != nil {
        if isDuplicateKeyError(err) {
            return errors.NewConflict(err, "user already exists", errors.Trace())
        }
        return errors.NewInternal(err, "failed to insert user", errors.Trace())
    }
    
    return nil
}
```

### Cross-Service Communication

```go
func (s *OrderService) ProcessOrder(ctx context.Context, orderID string) error {
    // Calling another service that uses this errors package
    user, err := s.userService.GetUser(ctx, userID)
    if err != nil {
        // ✅ Error from internal service: use Stack() or StackMsg()
        return errors.StackMsg(err, "failed to fetch user for order", errors.Trace())
    }
    
    // Calling external API
    payment, err := s.paymentClient.Charge(amount)
    if err != nil {
        // ✅ External service error: use New*()
        return errors.NewInternal(err, "payment processing failed", errors.Trace())
    }
    
    return nil
}
```
```

## Constructor Parameters

The `New*` constructors accept variadic parameters of the following types in any order:

- `error`: Wrapped error (becomes the cause)
- `string`: User-friendly message
- `ErrTrace`: Stack trace information (use `errors.Trace()`)
- `int`: Custom HTTP status code

```go
// All valid, parameters can be in any order:
errors.NewBadRequest("message", errors.Trace())
errors.NewBadRequest(errors.Trace(), "message")
errors.NewBadRequest(originalErr, "message", errors.Trace(), 422)
errors.NewBadRequest(422, errors.Trace(), "message", originalErr)
```

### Decision Tree: When to Use What

```
Is the error from a package/service using github.com/jurado-dev/errors?
│
├─ YES → Use Stack() or StackMsg()
│         • Preserves original error type
│         • Maintains type checking with Is*()
│         • Builds complete stack trace
│
└─ NO → Use New*() constructor
          • Wraps external errors (stdlib, third-party, etc.)
          • Creates new typed error
          • Starts new stack trace
```

## API Reference

### Error Constructors
- `NewBadRequest(...interface{}) *BadRequest`
- `NewUnauthorized(...interface{}) *Unauthorized`
- `NewNotFound(...interface{}) *NotFound`
- `NewConflict(...interface{}) *Conflict`
- `NewInternal(...interface{}) *Internal`
- `NewFatal(...interface{}) *Fatal`
- `NewNoContent(...interface{}) *NoContent`

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
- `GetCause(error) string` - Get original error message
- `GetMessage(error) string` - Get user-friendly message
- `GetCode(error) int` - Get HTTP status code
- `GetTrace(error) ErrTrace` - Get initial trace
- `GetStack(error) []ErrTrace` - Get full stack trace
- `GetStackJson(error) string` - Get stack as JSON
- `GetWrapped(error) error` - Get wrapped error
- `ErrorF(error) string` - Get formatted full error details
- `Unwrap(error) error` - Unwrap to original error

## License

MIT
