## Intro

Trebble middleware for Go works with applications based on `net/http`.

## Installation

```shell
go get github.com/treblle/treblle-go
```

Trebble uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies.


## Basic configuration

Configure Treblle at the start of your `main()` function:

```go
import "github.com/treblle/treblle-go"

func main() {
    treblle.Configure(treblle.Configuration{
        APIKey:          "YOUR API KEY HERE",
        ProjectID:    "YOUR PROJECT ID HERE",
        // 
        KeysToMask: []string{"password", "card_number"},
    })

    // rest of your program.
}

```


After that, just use the middleware with any of your handlers
 ```go
import "github.com/treblle/treblle-go"


mux := http.NewServeMux()
mux.Handle("/", treblle.Middleware(yourHandler))
```