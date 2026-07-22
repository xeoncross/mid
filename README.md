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
type HandlerFunc[T any] func(input T) (any, error)
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
| `WithDecoder` | `JSONDecoder` | `func WithDecoder[T any](d Decoder[T]) Option[T]` |
| `WithValidator` | `StructValidator` | `func WithValidator[T any](v Validator[T]) Option[T]` |
| `WithErrorHandler` | `JSONErrorHandler` | `func WithErrorHandler[T any](e ErrorHandler[T]) Option[T]` |

A decoder or validator only *reports* failure — it returns an `error` and never
touches the `http.ResponseWriter`. Every failure (query/body decode, validation,
your handler's own error, or a response-encoding error) is routed through the one
`ErrorHandler` configured via `WithErrorHandler`, so a single override changes how
*all* errors are rendered.

`WithDecoder` and `WithValidator` infer `T` from the function you pass in, so no type argument is needed. `ErrorHandler[T]` doesn't use `T` in its own signature, so when `WithErrorHandler` is the *only* option on a call, Go can't infer it from context and you need to spell it out:

```go
mux.Handle("/users", mid.Handler(createUser, mid.WithErrorHandler[CreateUserInput](myErrorHandler)))
```

## Components

### JSONDecoder[T]

Decodes the request body into your input struct. Handles common JSON errors and returns a descriptive error for the `ErrorHandler` to render.

```go
type Decoder[T any] func(r *http.Request, input *T) error
```

Returns a non-nil `error` if decoding fails; the error is routed to the configured `ErrorHandler`.

### StructValidator[T]

Validates your input struct using `go-playground/validator`.

```go
type Validator[T any] func(input T) error
```

Returns a `ValidationErrors` if constraints fail (rendered by `JSONErrorHandler` as a structured `{"errors":[...]}` body), or any other `error`, which is routed to the configured `ErrorHandler`.

### JSONErrorHandler

Default error handler that renders every failure in a consistent JSON format.

```go
type ErrorHandler[T any] func(w http.ResponseWriter, r *http.Request, input T, err error)
```

Default response format:

```json
{
    "error": "error message here"
}
```

A `ValidationErrors` is instead rendered as its structured form:

```json
{
    "errors": [
        {"field": "User.Email", "tag": "email", "message": "failed 'email' validation"}
    ]
}
```

## Validation with go-playground/validator

This package uses [go-playground/validator](https://github.com/go-playground/validator) for struct validation. Validation rules are defined using struct tags.

### Important Notes

1. **Validator behind an interface**: `mid` depends only on a small interface, so any implementation (including go-playground's `*validator.Validate`) can be used:

   ```go
   type Validate interface {
       Struct(s any) error
   }
   ```

   The default `StructValidator` is backed by the package-level `DefaultValidator` (a `validator.New()`). For dependency injection, build a validator explicitly and pass it per-handler with `NewStructValidator`:

   ```go
   v := validator.New()
   v.RegisterValidation("custom_rule", myCustomValidator)
   mux.Handle("/users", mid.Handler(createUser, mid.WithValidator(mid.NewStructValidator[CreateUserInput](v))))
   ```

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

4. **Custom Validators**: Configure a concrete `*validator.Validate` and assign it to `DefaultValidator` before first use, or inject it per-handler with `NewStructValidator` (see note 1):

```go
v := validator.New()
v.RegisterValidation("custom_rule", myCustomValidator)
mid.DefaultValidator = v // affects the default StructValidator globally
```

For a complete list of validation tags, see the [go-playground/validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10#readme-builtin-validators).

## Custom Implementations

All components are designed to be replaceable. You can provide your own decoder, validator, or error handler by implementing the corresponding type:

```go
// Custom decoder example — populate *input, return an error to reject.
func myDecoder[T any](r *http.Request, input *T) error {
    // Custom decoding logic
    return nil
}

// Custom validator example — return an error to reject.
func myValidator[T any](input T) error {
    // Custom validation logic
    return nil
}

// Usage
mid.Handler(myHandler, mid.WithDecoder(myDecoder), mid.WithValidator(myValidator))
```

## License

MIT