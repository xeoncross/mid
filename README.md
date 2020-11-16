# Mid [![Go Report Card](https://goreportcard.com/badge/github.com/xeoncross/mid)](https://goreportcard.com/report/github.com/xeoncross/mid) [![GoDoc](https://godoc.org/github.com/xeoncross/mid?status.svg)](https://godoc.org/github.com/xeoncross/mid)

A small collection of lightweight, `net/http` compatible middlewares for common web server needs. It's a library that acts as a smart framework reducing the amount of work you need to do. It's a competitor to [echo](https://github.com/labstack/echo/), [gin](https://github.com/gin-gonic/gin), [buffalo](https://gobuffalo.io/en/), and [other web frameworks](https://awesome-go.com/#web-frameworks).

Mid currently solves:

# Readme Table of Contents
<!-- TOC -->

- [Automatic Input Validation](#validation) (build a REST API fast!)
- [Sessions](#sessions)
- [Serving Single Page Apps](#single-page-apps) (React, Angular, Vue, etc..)
- [os.Signals](#os-signals) (like CTRL+C to terminate a server)
- [Request throttling](#request-throttling)
- [Limiting Request Body Size](#max-body-size)
- [Templating](#templating)
- [Performance](#performace)

<!-- /TOC -->


## Validation

Mid provides a `net/http` compatible middleware for protecting, validating, and automatically
hydrating handlers with user input from JSON or multipart form bodies, route
parameters, and query string values.

This makes code easier to read by removing the need to type input decoding and validation checks for every handler. Imagine a simpler, automatic [gRPC](https://grpc.io/) for REST API's.

Compatible with:

- [golang.org/pkg/net/http/](https://golang.org/pkg/net/http/) (plain `http.Handler`)
- [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
- [github.com/gorilla/mux](https://github.com/gorilla/mux)

### Usage

For all user input you must define a struct that contains the expected fields and rules. For example, imagine we are saving a blog comment. We might get the blog post id from the URL path and the comment fields from a JSON body. We can use [struct tags](https://github.com/golang/go/wiki/Well-known-struct-tags) to specify the rules and location of the data we are expecting.

(We use [asaskevich/govalidator](https://github.com/asaskevich/govalidator#validatestruct-2) for validation.)

```go
type NewComment struct {
	PostID  int    `valid:"required" param:"post_id"`
	Comment string `valid:"required,stringlength(100|1000)"`
	Email   string `valid:"required,email"`
}
```

> Note: the "param" struct tag specifies a route parameter named "post_id" holds the value

Next we write a http.HandlerFunc with _one extra field_: a reference to the *populated* `NewComment`:

```go
handler := func(w http.ResponseWriter, r *http.Request, comment NewComment) error {
	// we now have access to populated fields like "comment.Email"
	return nil
}
```

We then wire this up to our router and are ready to start accepting input:

```go
// julienschmidt/httprouter
router.Handler("POST", "/post/:post_id/comment", mid.Hydrate(handler))

// gorilla/mux
router.HandleFunc("/post/{post_id}/comment", mid.Hydrate(handler)).Methods("POST")

// net/http
http.Handle("/post/comment", mid.Hydrate(handler))
```

> Note: the last net/http mux example does not provide "route parameters" so it
> won't actually pass the validation due to the PostID struct tag "param" not existing.

At this point we can rest assured that our handler will never be called unless
input matching our exact validation rules is provided by the client. If the
client passes invalid data, then a JSON response object is returned specifying
the issues.

		{
			error: string
			fields: map[string]string
		}

We follow a simpler version of the Google style for JSON API responses:

> "A JSON response should contain either a data object or an error object,
>  but not both. If both data and error are present, the error object takes
> precedence." - https://google.github.io/styleguide/jsoncstyleguide.xml

### Responding to clients

What if you want to return data to the client? You can specify any type and it
will be returned to the client packaged in a standard JSON wrapper.

```go
handler := func(w http.ResponseWriter, r *http.Request, newComment NewComment) (Comment, error) {
	comment, err := commentService.Save(newComment)
	return comment, err
}
```

This will result in the following HTTP 200 response:

		{
			data: { ...Comment fields here... }
		}

### See the [examples](https://github.com/Xeoncross/mid/tree/master/examples) for more.

Supported Validations can be found at the validation project:

https://github.com/asaskevich/govalidator#list-of-functions

## Sessions

Moved to https://github.com/Xeoncross/session

## Single Page Apps

Javascript applications often have a single index.html + assets
that should be served at every (unused) URL path because the client bundle
handles the routing creating a "virtual" filesystem. (see window.history)

Furthermore, it is often a security threat to allow reading of "dot files"
or directory listings. This file contains two http.FileSystem wrappers to
solve these need: `SpaFileSystem()` and `DotFileHidingFileSystem()`

The `FileSystem()` wrapper provides both making it easy to serve React, 
Angular, or Vue apps in a safe way along with other HTTP endpoints.

```go
http.Handle("/", mid.FileSystem(http.FileServer(http.Dir("./app/public"))))
http.HandleFunc("/api/news", newsHandler)
```

## OS Signals

Easily respond to interrupt signals (SIGTERM, CTRL+C, etc..) using the InterruptContext() function which returns a context that deadlines when an OS signal is recived by the process. Helpful for restarting servers or writing terminal apps.

This can be combined with the ListenWithContext() function to handle gracefully stopping a HTTP server.

```go
server := &http.Server{
	Addr:         ":0",
	Handler:      myhandler,
	ReadTimeout:  1 * time.Second,
	WriteTimeout: 1 * time.Second,
	IdleTimeout:  1 * time.Second,
}

ctx := mid.InterruptContext()

err := mid.ListenWithContext(ctx, server, &healthy)
```


## Max Body Size

HTTP request bodies can be any size, it is recommended you limit them using the
`mid.MaxBodySize()` middleware to prevent attacks.

## Request Throttling

A large number of TCP requests can cause multiple issues including degraded
performance and your OS terminating your Go service because of high memory usage.
A [Denial-of-service attack](https://en.wikipedia.org/wiki/Denial-of-service_attack)
is one example. The `mid.RequestThrottler` exists to help keep a cap on how many
requests your application instance will serve concurrently.

It is recommended you create a helper function that wraps both these (along with any other middleware you are using like Hydration).

```go
// Close connection with a 503 error if not handled within 3 seconds
throttler := mid.RequestThrottler(20, 3*time.Second)

wrapper := func(function interface{}) http.Handler {
	return throttler(mid.MaxBodySize(mid.Hydrate(function), 1024*1024))
}

...

router.Handler("POST", "/:user/profile", wrapper(controller.SaveProfile))
```

## Templating

Please see https://github.com/Xeoncross/got - a minimal wrapper to improve Go's `html/template` usage by providing pre-computed inheritance with no loss of speed or modifications to the standard library's template processing.


## Performance

Much attention has been given to design mid components in a way that is 

1. Simple
2. Reduces developer work
3. Performs well

Since the main feature of mid is the input validation, it has been loosely compared to a couple other frameworks to give you an idea of the trade offs between these systems.

```
$ go test --bench=. --benchmem
goos: darwin
goarch: amd64
pkg: github.com/Xeoncross/mid/benchmarks
BenchmarkEcho-8       	  200000	      8727 ns/op	    3731 B/op	      39 allocs/op
BenchmarkGongular-8   	  100000	     18266 ns/op	    6511 B/op	      76 allocs/op
BenchmarkMid-8        	  100000	     12607 ns/op	    7174 B/op	     120 allocs/op
BenchmarkVanilla-8    	 2000000	       849 ns/op	     288 B/op	      17 allocs/op
```


Gongular is slower in these benchmarks because 1) it's a full framework with extra mux wrapping code and 2) calculations and allocs that go into handling dependency injection in a way mid is able to avoid completely by keeping the handler separate from the binding object.

Echo is the fastest, but also requires writing the most code because it is less feature-complete. Unlike the other two, the echo benchmark does not include things like URL parameter binding or standard response handling.


