.PHONY: install	# tells Make that the target is not associated with a file
install:
	go install

.PHONY: test
test:
	go test -coverprofile=coverage.txt -covermode=atomic ./... 

.PHONY: lint
lint: # run the fast go linters
	golangci-lint run --no-config \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck