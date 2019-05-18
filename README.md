# mid

Simple Go HTTP middleware for reducing code substantially when building a web app. Simply tell mid what type of struct you are expecting (and how the client should send it) and it will only call your handler if the validation for that struct passes.

```
// Post is one of your app entities with validation defined
type Post struct {
	Title   string `valid:"alphanum,required"`
	Message string `valid:"ascii,required"`
}

// PostInput defines where we should look for the Post object (a HTTP Form)
type PostInput struct {
	Form Post
}
// Other options include: Param, Query, and Body (JSON payloads)

// You can also combine everything if you don't have an existing domain entity
// type Post struct {
// 	Form struct {
//   	Title   string `valid:"alphanum,required"`
//   	Message string `valid:"ascii,required"`
//   }
// }


router := httprouter.New()
router.POST("/validate", mid.Validate(handlers.validateHandler, &PostInput{}))
```

See the [examples](https://github.com/Xeoncross/mid/tree/master/examples).


## Warning

This is beta quality software. The API might change.


## Why?

I really don't like typing the same HTTP handler validation logic over-and-over. This library provides automatic user input processing/validation and population of my domain objects. Better than echo.Bind, Gongular, or any other libraries I looked at over the last year.

Any invalid requests will receive a JSON response stating which fields have invalid values. If you want to handle the response yourself, you can set a special `nojson bool` property on your struct.


## Supported Validations

https://github.com/asaskevich/govalidator#list-of-functions


# Benchmarks

The performance of Mid is almost twice of that of Gongular. However, part of this is that Gongular is a full framework (lots of extra wrappers and allocs). Mid is simply a chainable middleware, trying to stay out of the way. Gongular's approach is more clunky trying to take over the http.Handler while still preserving it's dependencies. This causes extra delay for each request.

```
$ go test --bench=. --benchmem
goos: darwin
goarch: amd64
pkg: github.com/Xeoncross/mid/benchmarks
BenchmarkGongular-8   	  200000	      7494 ns/op	    7062 B/op	      61 allocs/op
BenchmarkMid-8        	  500000	      3390 ns/op	    2227 B/op	      32 allocs/op
BenchmarkVanilla-8    	 2000000	       633 ns/op	     160 B/op	      12 allocs/op
```


# Templates

Please use https://github.com/Xeoncross/got - a minimal wrapper to improve Go `html/template` usage with no loss of speed.


## Alternative Method(s)

This library is basically the best of both two approaches: a separate struct, which is self-describing and automatically validated before calling the handler.

- http://github.com/mustafaakin/gongular: The handler _is_ the validation schema
- https://github.com/mholt/binding: separate struct for the validation mapping so that multiple handlers can share. Requires wiring and repeated binding configuration for each struct.


## Related Projects

These projects are related in the sense of returning of structs/errors/maps directly from HTTP handlers and providing automatic input validation.

- [Gongular](https://github.com/mustafaakin/gongular#how-to-use) (more features, uses reflection)
- [Macaron](https://go-macaron.com/docs/intro/core_concepts)
- [Tango](https://github.com/tango-contrib/binding)


## Reading

- https://justinas.org/writing-http-middleware-in-go/
- https://hackernoon.com/simple-http-middleware-with-go-79a4ad62889b
- https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
- https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702
- https://stackoverflow.com/questions/6365535/http-handlehandler-or-handlerfunc
- https://www.nicolasmerouze.com/middlewares-golang-best-practices-examples/
- http://www.alexedwards.net/blog/making-and-using-middleware
- https://gist.github.com/nilium/f2ec7dcd54accd23532e82b04f1df7de
- https://github.com/rsc/tiddly/
- https://www.reddit.com/r/golang/comments/6fl86p/wrapping_httpresponsewriter_for_middleware/
- https://www.jtolds.com/2017/01/writing-advanced-web-applications-with-go/
- https://github.com/mholt/caddy/blob/master/caddyhttp/httpserver/middleware.go
- http://www.akshaydeo.com/blog/2017/12/23/How-did-I-improve-latency-by-700-percent-using-syncPool/
- https://golang.org/pkg/sync/#example_Pool
- https://github.com/go-chi/chi/blob/master/_examples/rest/main.go
- https://blog.golang.org/error-handling-and-go#TOC_3.
- https://www.reddit.com/r/golang/comments/7yt1w2/experiments_with_httphandler/
- https://cryptic.io/go-http/
- https://gist.github.com/husobee/fd23681261a39699ee37
- https://www.reddit.com/r/golang/comments/7umarx/http_input_validation/
