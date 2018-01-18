# mid

Simple Go HTTP middleware for reducing code substantially when building a web app. `net/http` compatible.

See the [examples](http://github.com/xeoncross/mid/examples).

## Why?

Most middleware libraries solve easy problems like error recovery and logging. I wanted something that would help me validate user input, render nested templates, return JSON responses, and other common tasks.

### Mid is

- Fast
- Simple and free from slow, magic packages like `reflect`.
- DRY ([Don't repeat yourself](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself))
- Compatible with the big three http routers/multiplexers:
  - `net/http`
  - https://github.com/gorilla/mux
  - https://github.com/julienschmidt/httprouter (TODO)

## Related Projects

- [Gongular](https://github.com/mustafaakin/gongular#how-to-use) (more features, uses reflection)


## Reading

- https://justinas.org/writing-http-middleware-in-go/
- https://hackernoon.com/simple-http-middleware-with-go-79a4ad62889b
- https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
- https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702
- https://stackoverflow.com/questions/6365535/http-handlehandler-or-handlerfunc
- https://www.nicolasmerouze.com/middlewares-golang-best-practices-examples/
- http://www.alexedwards.net/blog/making-and-using-middleware
- https://gist.github.com/nilium/f2ec7dcd54accd23532e82b04f1df7de
