BINDIR      := $(CURDIR)/bin
BINNAME     ?= testprog
SRC        := $(shell find . -type f -name '*.go' -print)

GOBIN = $(shell go env GOBIN)
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$(BINDIR)/$(BINNAME)"  -ldflags '-extldflags "-static"' .

.PHONY: build-linux
build-linux: $(BINDIR)/$(BINNAME)

.PHONY: clean
clean:
	@rm -rf '$(BINDIR)'
