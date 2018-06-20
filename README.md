# mid

Simple Go HTTP middleware for reducing code substantially when building a web app. `net/http` compatible.

See the [examples](https://github.com/Xeoncross/mid/tree/master/examples).

## Why?

Most middleware libraries solve easy problems like error recovery and logging. I wanted something that would help me validate user input, render nested templates, return JSON responses, and other common tasks.

### Mid is

- Fast
- Simple
- DRY ([Don't repeat yourself](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself))
- Compatible with the big three http routers/multiplexers:
  - `net/http`
  - https://github.com/gorilla/mux
  - https://github.com/julienschmidt/httprouter (TODO)

## Thoughts

Adding template support to gongular seems to be a flop unless we assume:

1. a non-javascript site
2. only apply validation if a POST/PUT/DELETE request.
3. Avoid auto-loading templates

The original idea of a single handler that returns JSON or HTML depending on if
a `template.Template` is set doesn't make much sense.

- Endpoints should be JSON only not rendering pages (Modern web apps).
- HTML pages that don't use AJAX should not send back AJAX.

Originally I was going to have any request that does not contain all required data
bypass the handler and return the errors.

For HTML pages this meant skipping straight to loading the template. However,
the template often relies on data provided by the handler, so it is not very
useful to skip that part.

Consider a form that needs to be rendered on first load + some select lists or
other data provided by the handler. Validation only applies on a subsequent POST
request, but we would still need the extra data provided by the handler.



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
