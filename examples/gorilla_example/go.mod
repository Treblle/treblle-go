module github.com/Treblle/treblle-go/examples/gorilla_example

go 1.21

require (
	github.com/Treblle/treblle-go/v2 v2.0.0
	github.com/gorilla/mux v1.8.1
)

require golang.org/x/sync v0.11.0 // indirect

replace github.com/Treblle/treblle-go/v2 => ../../
