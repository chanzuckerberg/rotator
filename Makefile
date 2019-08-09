.PHONY: install	# tells Make that the target is not associated with a file

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
