SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=false
# TODO add release flag
GO_PACKAGE=$(shell go list)
LDFLAGS=-ldflags "-w -s -X $(GO_PACKAGE)/util.GitSha=${SHA} -X $(GO_PACKAGE)/util.Version=${VERSION} -X $(GO_PACKAGE)/util.Dirty=${DIRTY}"
export GO111MODULE=on

setup: # setup development dependencies
	export GO111MODULE=on
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh -s -- v0.9.14
.PHONY: setup

install:
	go install
.PHONY: install

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

test-all:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./... -tags=integration
.PHONY: test-all

test-coverage:  ## run the test with proper coverage reporting
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out
.PHONY: test-coverage

test-coverage-integration:  ## run the test with proper coverage reporting
	go test -coverprofile=coverage.out -covermode=atomic ./... -tags=integration
	go tool cover -html=coverage.out
.PHONY: test-coverage-all

# lint: # run the fast go linters
# 	./bin/golangci-lint run --no-config \
# 		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
# 		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
# 		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck --deadline=5m
# .PHONY: lint

lint: ## run the fast go linters on the diff from master
	./bin/reviewdog -conf .reviewdog.yml  -diff "git diff master"
.PHONY: lint

lint-ci: ## run the fast go linters
	./bin/reviewdog -conf .reviewdog.yml  -reporter=github-pr-review
.PHONY: lint-ci

lint-all: ## run the fast go linters
	# doesn't seem to be a way to get reviewdog to not filter by diff
	golangci-lint run
.PHONY: lint-all

deps:
	go get -u ./...
	go mod tidy
.PHONY: deps


release: ## run a release
	bff bump
	git push
	goreleaser release --rm-dist
.PHONY: release

build: ## build the binary
	go build .
.PHONY: build

release-prerelease: setup build ## release to github as a 'pre-release'
	go build ${LDFLAGS} .
	version=`./rotator version`; \
	git tag v"$$version"; \
	git push
	git push --tags
	goreleaser release -f .goreleaser.prerelease.yml --debug
.PHONY: release-prerelease

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod

fmt:
	goimports -w -d $$(find . -type f -name '*.go' -not -path "./vendor/*")
.PHONY: fmt
