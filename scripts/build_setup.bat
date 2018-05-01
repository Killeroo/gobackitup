:: https://golang.org/doc/install/source#environment
set GOARCH=amd64
set GOOS=windows
go tool dist install -v pkg/runtime
go install -v -a std