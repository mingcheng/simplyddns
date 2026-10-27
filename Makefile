.DEFAULT_GOAL := build
.PHONY: all build install clean test race vet fmt fmt-check mod-check lint check vulncheck \
	version act darwin_universal release dist verify-dist build_docker_image

VERSION ?= 2.0.0
BIN ?= simplyddns
OUTPUT ?= $(BIN)
DIR_SRC := ./cmd/simplyddns
MISE ?= mise
DOCKER_CMD ?= docker
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
DIST_DIR ?= dist
RELEASE_TARGETS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

GO := $(MISE) exec go -- go
GO_FLAGS = -trimpath -ldflags="-s -w -X main.BuildVersion=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.BuildCommit=$(GIT_COMMIT)"

build:
	@CGO_ENABLED=0 $(GO) build $(GO_FLAGS) -o "$(OUTPUT)" $(DIR_SRC)

install:
	@CGO_ENABLED=0 $(GO) install $(GO_FLAGS) $(DIR_SRC)

version:
	@printf '%s\n' "$(VERSION)"

test:
	@CGO_ENABLED=0 $(GO) test ./...

race:
	@CGO_ENABLED=1 $(GO) test -race -covermode=atomic -coverprofile=coverage.out ./...

vet:
	@$(GO) vet ./...

fmt:
	@$(GO) fmt ./...

fmt-check:
	@set -e; files=$$($(MISE) exec go -- gofmt -l .); \
	if [ -n "$$files" ]; then printf 'Run mise exec -- make fmt:\n%s\n' "$$files"; exit 1; fi

mod-check:
	@$(GO) mod tidy -diff

lint:
	@$(MISE) exec actionlint shellcheck -- actionlint .github/workflows/*.yml .gitea/workflows/*.yaml
	@$(MISE) exec shellcheck -- shellcheck scripts/*.sh
	@$(MISE) exec hadolint -- hadolint Dockerfile

vulncheck:
	@$(MISE) exec go go:golang.org/x/vuln/cmd/govulncheck -- govulncheck ./...

check: fmt-check mod-check vet test race lint

darwin_universal:
	@$(MAKE) build GOOS=darwin GOARCH=arm64 OUTPUT="$(BIN)_arm64"
	@$(MAKE) build GOOS=darwin GOARCH=amd64 OUTPUT="$(BIN)_amd64"
	@lipo -create -output "$(OUTPUT)" "$(BIN)_arm64" "$(BIN)_amd64"
	@rm -f "$(BIN)_arm64" "$(BIN)_amd64"

build_docker_image:
	@$(MISE) exec docker-cli -- $(DOCKER_CMD) build \
		--build-arg VERSION="$(VERSION)" \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg BUILD_COMMIT="$(GIT_COMMIT)" \
		-f Dockerfile -t "simplyddns:$(VERSION)" .

release: check vulncheck verify-dist

dist:
	@mkdir -p "$(DIST_DIR)"
	@set -eu; stage=$$(mktemp -d "$(DIST_DIR)/.stage.XXXXXX"); archive=""; \
	cleanup() { \
		rm -f "$$stage/$(BIN)" "$$stage/$(BIN).exe" "$$stage/LICENSE" \
			"$$stage/README.md" "$$stage/README_ZH.md" "$$stage/simplyddns.yml"; \
		if [ -n "$$archive" ]; then rm -f "$$stage/$$archive"; fi; \
		rmdir "$$stage"; \
	}; \
	trap cleanup EXIT; \
	cp LICENSE README.md README_ZH.md "$$stage/"; \
	cp example/basic.yml "$$stage/simplyddns.yml"; \
	set --; \
	for target in $(RELEASE_TARGETS); do \
		os=$${target%/*}; arch=$${target#*/}; \
		name="$(BIN)_$(VERSION)_$${os}_$${arch}"; \
		bin="$(BIN)"; archive="$$name.tar.gz"; \
		if [ "$$os" = windows ]; then bin="$(BIN).exe"; archive="$$name.zip"; fi; \
		GOOS="$$os" GOARCH="$$arch" CGO_ENABLED=0 $(GO) build $(GO_FLAGS) -o "$$stage/$$bin" $(DIR_SRC); \
		if [ "$$os" = windows ]; then \
			(cd "$$stage" && zip -q "$$archive" "$$bin" LICENSE README.md README_ZH.md simplyddns.yml); \
		else \
			(cd "$$stage" && COPYFILE_DISABLE=1 tar -czf "$$archive" "$$bin" LICENSE README.md README_ZH.md simplyddns.yml); \
		fi; \
		mv "$$stage/$$archive" "$(DIST_DIR)/$$archive"; \
		rm -f "$$stage/$$bin"; \
		set -- "$$@" "$$archive"; \
	done; \
	test "$$#" -gt 0; \
	(cd "$(DIST_DIR)" && shasum -a 256 "$$@" > checksums.txt)

verify-dist: dist
	@SIMPLYDDNS_DIST_DIR="$(abspath $(DIST_DIR))" SIMPLYDDNS_VERSION="$(VERSION)" \
		SIMPLYDDNS_BIN="$(BIN)" SIMPLYDDNS_TARGETS="$(RELEASE_TARGETS)" \
		$(GO) test . -run '^TestReleaseArtifacts$$' -count=1

clean:
	@rm -f "$(OUTPUT)" "$(BIN)_arm64" "$(BIN)_amd64" coverage.out
	@if [ -d "$(DIST_DIR)" ]; then \
		find "$(DIST_DIR)" -maxdepth 1 -type f \
			\( -name "$(BIN)_*.tar.gz" -o -name "$(BIN)_*.zip" -o -name checksums.txt \) -delete; \
	fi

act:
	@$(MISE) exec act -- act -W .github/workflows/ci.yml --container-architecture linux/amd64 -j checks

all: check build
