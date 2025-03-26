module github.com/Treblle/treblle-go/examples/treblle-go-sdk-example

go 1.23.0

toolchain go1.23.2

require (
	github.com/Treblle/treblle-go/v2 v2.0.0
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
)

require golang.org/x/sync v0.12.0 // indirect

replace github.com/Treblle/treblle-go/v2 => ../../
