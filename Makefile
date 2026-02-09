.DEFAULT_GOAL := run
.PHOHY: run build test vet fmt

fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

# create binary in ./bin directory
build: vet
	@go build -o ./bin/out ./cmd/main.go

run: build
	@./bin/out


