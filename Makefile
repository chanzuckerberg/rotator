.PHONY: install	# tells Make that the target is not associated with a file
install:
	go install

.PHONY: test
test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

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
