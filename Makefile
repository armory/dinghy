# Borrowed from:
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


BINARY = dinghy
VET_REPORT = vet.report
TEST_REPORT = tests.xml
GOARCH = amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REPO=$(shell basename $(shell git rev-parse --show-toplevel))
#select all packages except a few folders because it's an integration test
PKGS := $(shell go list ./... | grep -v -e /integration -e /vendor)
INTEGRATION_PKGS := $(shell go list ./... | grep /integration)
ORG=armory-io
BUILD_DIR=$(shell pwd)/build

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	LDFLAGS = -ldflags "-X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -linkmode external -extldflags -static -s -w"
endif

# Build the project
all: clean dependencies test vet build

dependencies:
	go mod tidy && \
	go mod vendor

run:
	go run ./${BINARY}.go

build: ./${BINARY}.go
	go build -i ${LDFLAGS} -o ${BUILD_DIR}/${BINARY} ./${BINARY}.go

test: dependencies
	go test -cover -v $(PKGS)

# The go test tool won't create a coverage profile if you give it multiple
# packages. Recommendation is to run the coverage for each package and merge
# the coverage profiles (which are just text files). So that's what we do,
# first generating the header line in our target coverage profile, and then
# grabbing just the guts of each individual coverage file and appending.
#
# After you generage the coverage report  you can use the go tooling to view
# it in a browser:
#   go tool cover -html build/coverage/coverage.out
coverage: clean dependencies
	mkdir -p ${BUILD_DIR}/coverage ; \
	echo 'mode: set' > ${BUILD_DIR}/coverage/coverage.out ; \
	for TESTPKG in $(PKGS); do \
		go test --coverprofile=${BUILD_DIR}/coverage/coverage.tmp -v $$TESTPKG ; \
		if [ -f ${BUILD_DIR}/coverage/coverage.tmp ]; then \
			tail -n +2 ${BUILD_DIR}/coverage/coverage.tmp >> ${BUILD_DIR}/coverage/coverage.out ; \
			rm ${BUILD_DIR}/coverage/coverage.tmp ; \
		fi ; \
	done

integration: dependencies
	mkdir -p ${BUILD_DIR}/tests ; \
	for TESTPKG in $(INTEGRATION_PKGS); do \
	    EXENAME=$$(echo "$$TESTPKG" | tr / _) ; \
	    go test -c -o ${BUILD_DIR}/tests/$$EXENAME $$TESTPKG ; \
	done

golint:
	go get -v golang.org/x/lint/golint

lint: golint
	@golint $(PKGS)

vet:
	go vet -v ./...

fmt:
	go fmt $$(go list ./... | grep -v /vendor/) ; \

clean:
	rm -rf ${BUILD_DIR}; \
	go clean

# mac users need to use gnu-sed
dep:
	@echo "upgrading $(dep) to $(version)"
	$(eval OLD_VERSION := $(shell grep $(dep) go.mod | awk '{print $$2}'))
	$(eval CURR_BRANCH := $(shell git rev-parse --abbrev-ref HEAD))
	sed -i -e 's/$(dep) $(OLD_VERSION)/$(dep) $(version)/g' go.mod
	go mod vendor

docker:
	rm -fr build/docker ; \
	mkdir -p build ; \
	ln -sf $(shell pwd) $(shell pwd)/build/docker ; \
	./gradlew devSnapshot docker --info -x dockerClean -x dockerPrepare -x artifactoryDeploy -x dockerPush

docker-push:
	[[ -f build/docker-labels.json ]] || (echo 'You must run make docker, before running make docker-push' && exit 1) ; \
	docker push $$(cat build/docker-labels.json | jq -r ".name")

.PHONY: lint linux darwin test vet fmt clean run

lambda:
	GOOS=linux go build -i ${LDFLAGS} -o cachelambda/main ./cachelambda/main.go
	cd cachelambda && zip lambda_function_payload.zip main

