
test:
	go test ./...
	cd examples && go run .. stdlib_test.ge

benchmarks:
	go test -bench=.

cov:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

staticcheck:
	staticcheck ./...

cycl:
	gocyclo -top 10 .

repl:
	@ go run .

clean:
	go mod tidy
	go fmt
	rm -rf *.out *.html *.prof *.test
	go clean -testcache

dev:
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.57.1
    go get honnef.co/go/tools/cmd/staticcheck
    go get github.com/fzipp/gocyclo/cmd/gocyclo@latest
