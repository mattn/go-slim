module github.com/mattn/go-slim/_example/webapp

go 1.15

replace github.com/mattn/go-slim => ../..

require (
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/mattn/go-slim v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20220126234351-aa10faf2a1f8 // indirect
)
