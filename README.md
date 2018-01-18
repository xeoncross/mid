# mid

Simple Go HTTP middleware for reducing code substantially when building a web app. `net/http` compatible.

## Why?

Most middleware libraries solve easy problems like error recovery and logging. I wanted something that would help me validate user input, render nested templates, return JSON responses, and other common tasks.

### Mid is

- Fast
- Simple and free from slow, magic packages like `reflect`.
- DRY ([Don't repeat yourself](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself))
- Compatible with the big three http routers/multiplexers:
  - `net/http`
  - https://github.com/julienschmidt/httprouter
  - https://github.com/gorilla/mux
