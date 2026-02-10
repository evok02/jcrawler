GOFLAGS = GODEBUG=http2client=0
.DEFAULT_GOAL := run
.PHOHY: run build test vet fmt

fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

# create binary in ./bin directory
build: vet
	@$(GOFLAGS) go build -o ./bin/out ./cmd/main.go

run: build
	@./bin/out

test: build
	@go test ./...
