.PHONY: build clean test test-race

VERSION=1.4.2
BIN=simplyddns
DIR_SRC=./cmd/simplyddns
DOCKER_CMD=docker

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags="-X main.BuildVersion=$(VERSION) -X 'main.BuildTime=`date`' -extldflags -static"
GO=$(GO_ENV) $(shell which go)

build: test $(DIR_SRC)/main.go
	@$(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN) $(DIR_SRC)

darwin_universal: test $(DIR_SRC)/main.go
	@GOOS=darwin GOARCH=arm64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN)_arm64 $(DIR_SRC)
	@GOOS=darwin GOARCH=amd64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN)_amd64 $(DIR_SRC)
	@lipo -create -output $(BIN) $(BIN)_arm64 $(BIN)_amd64 
	@rm -f $(BIN)_arm64 $(BIN)_amd64 

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
