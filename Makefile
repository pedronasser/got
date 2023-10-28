BINARY_NAME=got
VERSION=0.0.0

build:
	go build -o bin/$(BINARY_NAME) .

build-debug:
	go build -gcflags="all=-N -l" -o bin/$(BINARY_NAME) .

install:
	$(eval GIT_COMMIT = $(shell git rev-list -1 HEAD))
	go install -ldflags \
		"-X main.Version=v$(VERSION)-$(GIT_COMMIT)" \
		github.com/pedronasser/got

.PHONY: build install