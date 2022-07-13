.PHONY: build clean test test-race

VERSION=1.4.1
BIN=simplyddns
DIR_SRC=./cmd/simplyddns
DOCKER_CMD=docker

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags="-X main.BuildVersion=$(VERSION) -X 'main.BuildTime=`date`' -extldflags -static"
GO=$(GO_ENV) $(shell which go)

build: test $(DIR_SRC)/main.go
	@$(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN) $(DIR_SRC)

build_docker_image: test clean
	@$(DOCKER_CMD) build -f ./Dockerfile -t simplyddns:$(VERSION) .

install: build
	@$(GO) install $(GO_FLAGS) $(DIR_SRC)

test:
	@$(GO) test ./...

# clean all build result
clean:
	@$(GO) clean ./...
	@rm -f $(BIN)
