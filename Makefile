
SHELL := /bin/sh

# The name of the executable (default is current directory name)
TARGET := vctui
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := 0.1.0
BUILD := `git rev-parse HEAD`

# Operating System Default (LINUX)
TARGETOS=linux

# Use linker flags to provide version/build settings to the target
#STATIC flags for (scratch and CGO_Enabled)
LDFLAGS=-ldflags "-s -w -X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -linkmode external -extldflags '-static'"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

REPO ?= thebsdbox
DOCKERREPO ?= $(REPO)
DOCKERTAG ?= $(VERSION)

.PHONY: all build clean install uninstall fmt simplify check run

all: check install

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)

install:
	@echo Building and Installing project
	@go install $(LDFLAGS)

uninstall: clean
	@rm -f $$(which ${TARGET})

fmt:
	@gofmt -l -w $(SRC)

dockerx86:
	@docker buildx build  --platform linux/amd64 --load -t $(DOCKERREPO)/$(TARGET):$(DOCKERTAG) -f Dockerfile .
	@echo New Multi Architecture Docker image created

docker:
	@docker buildx build  --platform linux/amd64,linux/arm64,linux/arm/v7 --push -t $(DOCKERREPO)/$(TARGET):$(DOCKERTAG) -f Dockerfile .
	@echo New Multi Architecture Docker image created

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${SRC}

run: install
	@$(TARGET)
