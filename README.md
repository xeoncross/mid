# Mid

A tiny, fast HTTP middleware for creating JSON APIs with automatic input validation using Go generics.

## Installation

```bash
go get github.com/xeoncross/mid
```

## Overview

`mid` provides a generic `Handler[T]` function that wraps your business logic into a standard `http.Handler`. It automatically handles JSON decoding, struct validation, and error responses, allowing you to focus on your application logic.

## Handler[T] Usage

The core function is `Handler[T]`, which takes four parameters and returns an `http.Handler`:

```go
func Handler[T any](
    handler HandlerFunc[T],   // Your business logic
    decode  Decoder[*T],      // JSON decoder
    validate Validator[T],    // Struct validator
    onErr   ErrorHandler,     // Error handler
) http.Handler
```

### HandlerFunc[T]

Your handler function must match the `HandlerFunc[T]` signature:

```go
type HandlerFunc[T any] func(input T) (error, any)
```

- **input**: The hydrated input struct (decoded from JSON and validated)
- **returns**: An error (if something went wrong) and any response value (encoded as JSON)

### Full Example

```go
package main

import (
    "net/http"
    "github.com/xeoncross/mid"
)

// Define your input struct with validation tags
type CreateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
}

// Define your response struct
type CreateUserResponse struct {
    ID    int    `json:"id"`
    Status string `json:"status"`
}

// Your handler logic
func createUser(input CreateUserInput) (error, any) {
    // ... business logic here ...
    return nil, CreateUserResponse{
        ID:     123,
        Status: "created",
    }
}

func main() {
    mux := http.NewServeMux()

    // Register the endpoint
    mux.Handle("/users", mid.Handler(
        createUser,                          // Your handler
        mid.JSONDecoder[CreateUserInput],    // Decode JSON body
        mid.StructValidator[CreateUserInput],// Validate struct
        mid.JSONErrorHandler,                // Handle errors
    ))

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    server.ListenAndServe()
}
```

## Components

### JSONDecoder[T]

Decodes the request body into your input struct. Handles common JSON errors and returns appropriate error messages.

```go
type Decoder[T any] func(w http.ResponseWriter, r *http.Request, input T) bool
```

Returns `false` if decoding fails (error response is automatically written to the client).

### StructValidator[T]

Validates your input struct using `go-playground/validator`.

```go
type Validator[T any] func(w http.ResponseWriter, r *http.Request, input T) bool
```

Returns `false` if validation fails (validation errors are automatically returned to the client as JSON).

### JSONErrorHandler

Default error handler that returns errors in a consistent JSON format.

```go
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
```

Default response format:

```json
{
    "error": "error message here"
}
```

## Validation with go-playground/validator

This package uses [go-playground/validator](https://github.com/go-playground/validator) for struct validation. Validation rules are defined using struct tags.

### Important Notes

1. **Global Validator Instance**: The package maintains a global `ValidatorInstance` (`*validator.Validate`) that is shared across all validation calls. This instance is lazily initialized on first use.

2. **Validation Tags**: Use the `validate` struct tag to define validation rules:

```go
type Input struct {
    Name     string  `validate:"required"`
    Email    string  `validate:"required,email"`
    Age      int     `validate:"gte=0,lte=150"`
    Role     string  `validate:"oneof=admin user guest"`
    Tags     []string `validate:"min=1"`
    Score    float64 `validate:"gte=0,lte=100"`
}
```

3. **Validation Errors**: When validation fails, the raw `validator.ValidationErrors` slice is returned as JSON, allowing clients to inspect which fields failed validation.

4. **Custom Validators**: You can register custom validation functions on `ValidatorInstance` before first use:

```go
mid.ValidatorInstance = validator.New()
mid.ValidatorInstance.RegisterValidation("custom_rule", myCustomValidator)
```

### Common Validation Tags

| Tag | Description |
|-----|-------------|
| `required` | Field must be set and non-empty |
| `email` | Field must be a valid email address |
| `min=N` | Field must be >= N |
| `max=N` | Field must be <= N |
| `gte=N` | Field must be >= N (for numbers) |
| `lte=N` | Field must be <= N (for numbers) |
| `oneof=x y z` | Field must be one of the specified values |
| `url` | Field must be a valid URL |
| `len=N` | Field must have exactly length N |

For a complete list of validation tags, see the [go-playground/validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10#readme-builtin-validators).

## Custom Implementations

All components are designed to be replaceable. You can provide your own decoder, validator, or error handler by implementing the corresponding type:

```go
// Custom decoder example
func myDecoder[T any](w http.ResponseWriter, r *http.Request, input *T) bool {
    // Custom decoding logic
    return true
}

// Custom validator example
func myValidator[T any](w http.ResponseWriter, r *http.Request, input T) bool {
    // Custom validation logic
    return true
}

// Usage
mid.Handler(myHandler, myDecoder, myValidator, mid.JSONErrorHandler)
```

## License

MIT