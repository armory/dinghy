# Borrowed from:
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

BINARY = dinghy
VET_REPORT = vet.report
TEST_REPORT = tests.xml
GOARCH = amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
#REPO=$(shell basename $(shell pwd))
REPO=$(shell basename $(shell git rev-parse --show-toplevel))
#select all packages except a few folders because it's an integration test
PKGS := $(shell go list ./... | grep -v -e /integration -e /vendor)
INTEGRATION_PKGS := $(shell go list ./... | grep /integration)
ORG=armory-io
PROJECT_DIR=${GOPATH}/src/github.com/${ORG}/${REPO}
BUILD_DIR=${GOPATH}/src/github.com/${ORG}/${REPO}/build
CURRENT_DIR=$(shell pwd)
PROJECT_DIR_LINK=$(shell readlink ${PROJECT_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

# Build the project
all: clean dependencies lint test vet build

dependencies:
	dep ensure

run:
	go run ./cmd/${BINARY}.go

build: ./cmd/${BINARY}.go
	cd ${PROJECT_DIR}; \
	go build -i ${LDFLAGS} -o ${BUILD_DIR}/main ./cmd/${BINARY}.go ; \

test: dependencies
	go test -v $(PKGS)

integration: dependencies
	cd ${PROJECT_DIR} ; \
	mkdir -p ${BUILD_DIR}/tests ; \
	for TESTPKG in $(INTEGRATION_PKGS); do \
	    EXENAME=$$(echo "$$TESTPKG" | tr / _) ; \
	    go test -c -o ${BUILD_DIR}/tests/$$EXENAME $$TESTPKG ; \
	done

GOLINT=$(GOPATH)/bin/golint

$(GOLINT):
	go get -v golang.org/x/lint/golint

lint: $(GOLINT)
	@$(GOLINT) $(PKGS)

vet:
	cd ${PROJECT_DIR}; \
	go vet -v ./...

fmt:
	cd ${PROJECT_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \

clean:
	rm -rf ${BUILD_DIR}
	go clean

.PHONY: lint linux darwin test vet fmt clean run
