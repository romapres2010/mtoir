go env -w GOPRIVATE=github.com
go mod tidy
rem go clean -modcache
go get -u ./../...
go mod vendor
rem make -f make vendor