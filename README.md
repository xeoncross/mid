# mid

Simple Go HTTP middleware for reducing code substantially when building a web app.

- `net/http` compatible.
- Biggest feature is automatic input validation.
- No framework lock-in

See the [examples](https://github.com/Xeoncross/mid/tree/master/examples).

###Warning

This is alpha quality software. The unit tests aren't finished and the API might change.

## Why?

Most middleware libraries solve easy problems like error recovery and logging. I wanted something that would help me validate user input, return JSON/gRPC responses, and other common tasks.

### Mid is

- Fast
- Simple
- DRY ([Don't repeat yourself](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself))
- Compatible with the big three http routers/multiplexers:
  - `net/http`
  - https://github.com/gorilla/mux (TODO)
  - https://github.com/julienschmidt/httprouter

## Usage

You create a handler which has struct properties that match the parameters you are expecting. These values can be form fields, JSON bodies, or URL params.

```
type MyHandler struct {
	Body struct {
		Bio string
		Age int `valid:"required"`
	}
	Param struct {
		Name string `valid:"alpha"`
	}
}

func (h MyHandler) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	fmt.Printf("Validation Errors: %+v\n", ValidationErrors)
	fmt.Printf("Populated Handler: %+v\n", h)
	return nil
}
```

Then you ask Mid to validate all requests so that only ones matching the values + rules of the handlers run the handler. In the route below, we expect a `HTTP POST` request to the URL path `/hello/[alpha]` with the two values `Bio` and `Age` either sent as `multipart/form-data` or an `application/json` body.

```
router := httprouter.New()
router.POST("/hello/:Name", Validate(&MyHandler{}, false, nil))
```

Any invalid requests will receive a JSON response stating which fields have invalid values. If you want to handle the response yourself, you can set a special `nojson bool` property on your handler struct.

# Templates

Please use https://github.com/Xeoncross/got - a minimal wrapper to improve Go `html/template` usage with no loss of speed.

# Benchmarks

The performance of Mid is almost twice of that of Gongular. However, part of this is that Gongular is a full framework (lots of extra wrappers and allocs). Mid is simply a chainable middleware, trying to stay out of the way.

One big feature (incomplete in mid) is dependency injection.

```
$ go test --bench=. --benchmem
goos: darwin
goarch: amd64
pkg: github.com/Xeoncross/mid/benchmarks
BenchmarkGongular-8   	  200000	      7481 ns/op	    7332 B/op	      60 allocs/op
BenchmarkMid-8        	  300000	      4363 ns/op	    2568 B/op	      38 allocs/op
BenchmarkVanilla-8    	 3000000	       460 ns/op	      64 B/op	       6 allocs/op
```

# Background

[Gongular](github.com/mustafaakin/gongular) is a neat framework that handles input validation and DI for the http.Handler. However, they don't support HTTP templates. It also adds noticeable overhead.

My goals were:

 - to simplify the code
 - increase performance
 - support non-JSON response bodies (especially `html/template`)

## Inspiration

Originally, I wanted to add template support to gongular. However, this idea seems to be a flop unless we assume:

1. a non-javascript site
2. only apply validation if a POST/PUT/DELETE request.
3. Avoid auto-loading templates

The original idea of a single handler that returns JSON or HTML depending on if
a `template.Template` is set doesn't make much sense.

- Endpoints should be JSON only not rendering pages (Modern web apps).
- HTML pages that don't use AJAX need to run the handler again, so mid can't provide
  any kind of short cut.

Originally I was going to have any request that does not contain all required data
bypass the handler and return the errors.

For HTML pages this meant skipping straight to loading the template. However,
the template often relies on data provided by the handler, so it is not very
useful to skip that part.

Consider a form that needs to be rendered on first load + some select lists or
other data provided by the handler. Validation only applies on a subsequent POST
request, but we would still need the extra data provided by the handler.

## Solution

Validation defaults to returning JSON objects on failure (never calling the
handler). However, if a `NOJSON` function/property is defined on the struct then
we call the handler providing the results of the validation and let it run
normally.

This allows handlers to render XML, templates, or anything else along with the
pre-validated input information.

## TODO

Need to make the validation handler re-attach any struct properties that contain
non-zero values. This is so you can set database handles or other things on the
handler and have them passed onto the new copy when the handler is cloned.


## Related Projects

These projects are related in the sense of returning of structs/errors/maps directly from HTTP handlers and providing automatic input validation.

- [Gongular](https://github.com/mustafaakin/gongular#how-to-use) (more features, uses reflection)
- [Macaron](https://go-macaron.com/docs/intro/core_concepts)
- [Tango](https://github.com/lunny/tango)


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
