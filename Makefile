.PHONY: setup
setup: # setup development dependencies
	export GO111MODULE=on
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest
	go get -u github.com/haya14busa/goverage
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -L https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh

.PHONY: install	# .PHONY tells Make that the target is not associated with a file
install:
	go install

.PHONY: test
test:
	gotest -v -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test-all
test-all:
	gotest -v -coverprofile=coverage.txt -covermode=atomic ./... -tags=integration

.PHONY: test-coverage
test-coverage:  ## run the test with proper coverage reporting
	goverage -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint: # run the fast go linters
	golangci-lint run --no-config \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck --deadline=5m

.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: release
release: ## run a release
	bff bump
	git push
	goreleaser release
