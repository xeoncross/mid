# Mid

A tiny, fast HTTP middleware for creating JSON APIs with automatic input validation.

```go
type Input struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type Output struct {
    ID int `json:"id"`
}

// Your handler logic
func createUser(input Input) (any, error) {
    // 'input' is auto-populated and passed validation
    // ... business logic here ...
    return Output{ ID: 123 }, nil
}
```

The most ergonomic way to quickly spin up fast HTTP endpoints for building JSON API's in Go. [Benchmarks can be found here](https://github.com/xeoncross/mid-benchmarks).

## Installation

```bash
go get github.com/xeoncross/mid
```


## Overview

`mid` provides a generic `Handler[T]` function that wraps your business logic into a standard `http.Handler`. It automatically handles JSON decoding, struct validation, and error responses, allowing you to focus on your application logic.

The core function is `Handler[T]`, which takes your handler plus any number of optional overrides and returns a `http.Handler`:

```go
func Handler[T any](
    handler HandlerFunc[T], // Your business logic
    opts    ...Option[T],   // Optional overrides (see Options below)
) http.Handler
```
For example, using [httprouter](https://github.com/julienschmidt/httprouter) looks like this:

```go
router := httprouter.New()
router.POST("/user/create", mid.Handler(createUser))
```

Your handler function must match the `HandlerFunc[T]` signature which differs from a regular [http.Handler](https://pkg.go.dev/net/http#Handler)

```go
type HandlerFunc[T any] func(input T) (error, any)
```

However, the result returned by `mid.Handler()` is a `net/http` compatible handler which makes using mid with other frameworks or existing systems easy.

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
    ID int `json:"id"`
}

// Your handler logic
func createUser(input CreateUserInput) (any, error) {
    // 'input' is auto-populated and passed validation
    // ... business logic here ...
    return CreateUserResponse{ ID: 123 }, nil
}

func main() {
    mux := http.NewServeMux()

    // Register the endpoint - decode, validate, and error handling all
    // default to the package helpers
    mux.Handle("/users", mid.Handler(createUser))

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    server.ListenAndServe()
}
```

### Options

Each default can be overridden individually by passing an `Option[T]` to `Handler`. Unlisted overrides keep their default.

```go
// Override just the validator
mux.Handle("/users", mid.Handler(createUser, mid.WithValidator(myValidator)))

// Override several at once
mux.Handle("/users", mid.Handler(createUser,
    mid.WithDecoder(myDecoder),
    mid.WithValidator(myValidator),
))
```

| Option | Overrides | Signature |
|---|---|---|
| `WithDecoder` | `JSONDecoder` | `func WithDecoder[T any](d Decoder[*T]) Option[T]` |
| `WithValidator` | `StructValidator` | `func WithValidator[T any](v Validator[T]) Option[T]` |
| `WithErrorHandler` | `JSONErrorHandler` | `func WithErrorHandler[T any](e ErrorHandler[T]) Option[T]` |

`WithDecoder` and `WithValidator` infer `T` from the function you pass in, so no type argument is needed. `ErrorHandler[T]` doesn't use `T` in its own signature, so when `WithErrorHandler` is the *only* option on a call, Go can't infer it from context and you need to spell it out:

```go
mux.Handle("/users", mid.Handler(createUser, mid.WithErrorHandler[CreateUserInput](myErrorHandler)))
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

1. **Global Validator Instance**: The package maintains a global `ValidatorInstance` (`*validator.Validate`) that is shared across all validation calls.

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

3. **Validation Errors**: When validation fails, a JSON object is returned, allowing clients to inspect which fields failed validation.

4. **Custom Validators**: You can register custom validation functions on `ValidatorInstance` before first use:

```go
mid.ValidatorInstance = validator.New()
mid.ValidatorInstance.RegisterValidation("custom_rule", myCustomValidator)
```

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
mid.Handler(myHandler, mid.WithDecoder(myDecoder), mid.WithValidator(myValidator))
```

## License

MIT