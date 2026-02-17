GOFLAGS = GODEBUG=http2client=0
GCFLAGS = -gcflags="all=-N -l"

.DEFAULT_GOAL := run
.PHOHY: run build test vet fmt

fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

# create binary in ./bin directory
build: vet
	@$(GOFLAGS) go build -o ./bin/out ./cmd/main.go

build-debug: vet
	@go build $(GCFLAGS) -o ./debug/out ./cmd/main.go

run: build
	@./bin/out

test: build
	@go test ./...

debug: build-debug
	@dlv exec ./debug/out
