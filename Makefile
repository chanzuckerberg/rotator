.PHONY: install	# tells Make that the target is not associated with a file
install:
	go install

.PHONY: test
test:
	go test -coverprofile=coverage.txt -covermode=atomic ./... 
