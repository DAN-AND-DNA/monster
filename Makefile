all: clean test

.PHONY: clean
clean:
	@go clean -testcache

.PHONY: test
test:
	@CGO_ENABLED=1 go test -v  ./pkg/errcode ./pkg/utils ./pkg/version
