.PHONY: build clean test act darwin_universal release dist build_docker_image

VERSION=2.0.0
BIN=simplyddns
DIR_SRC=./cmd/simplyddns
DOCKER_CMD=docker
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
DIST_DIR=dist
RELEASE_TARGETS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags="-s -w -X main.BuildVersion=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.BuildCommit=$(GIT_COMMIT) -extldflags -static"
GO=$(GO_ENV) go

build: $(DIR_SRC)/main.go
	@$(GO) build $(GO_FLAGS) -o $(BIN) $(DIR_SRC)

darwin_universal: $(DIR_SRC)/main.go
	@GOOS=darwin GOARCH=arm64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN)_arm64 $(DIR_SRC)
	@GOOS=darwin GOARCH=amd64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o $(BIN)_amd64 $(DIR_SRC)
	@lipo -create -output $(BIN) $(BIN)_arm64 $(BIN)_amd64
	@rm -f $(BIN)_arm64 $(BIN)_amd64

build_docker_image: clean
	@$(DOCKER_CMD) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILD_COMMIT=$(GIT_COMMIT) \
		-f ./Dockerfile -t simplyddns:$(VERSION) .

install: build
	@$(GO) install $(GO_FLAGS) $(DIR_SRC)

release: dist

dist: clean
	@mkdir -p $(DIST_DIR)
	@set -e; for target in $(RELEASE_TARGETS); do \
		os=$${target%/*}; \
		arch=$${target#*/}; \
		name="$(BIN)_$(VERSION)_$${os}_$${arch}"; \
		bin="$(BIN)"; \
		archive="$${name}.tar.gz"; \
		if [ "$${os}" = "windows" ]; then \
			bin="$(BIN).exe"; \
			archive="$${name}.zip"; \
		fi; \
		GOOS="$${os}" GOARCH="$${arch}" $(GO) build $(GO_FLAGS) -o "$(DIST_DIR)/$${bin}" $(DIR_SRC); \
		if [ "$${os}" = "windows" ]; then \
			(cd "$(DIST_DIR)" && zip -q "$${archive}" "$${bin}"); \
		else \
			(cd "$(DIST_DIR)" && tar -czf "$${archive}" "$${bin}"); \
		fi; \
		rm -f "$(DIST_DIR)/$${bin}"; \
	done

test:
	@$(GO) test ./...

# clean all build result
clean:
	@$(GO) clean ./...
	@rm -f $(BIN)
	@rm -rf $(DIST_DIR)

#https://github.com/nektos/act
act:
	@act -W ./.github/workflows --container-architecture linux/amd64 --pull=false -j test

all: clean test build
