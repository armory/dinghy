
all: build test

# Generate vendor dependencies
vendor:
	@go mod vendor

# Build the Dinghy binary
build: dinghy.go cmd pkg vendor
	@go build -mod vendor -v .

# Test this project
test: build
	@go test -v -mod=vendor -race -covermode atomic -coverprofile=profile.cov ./...

run: build
	@./dinghy

# Remove build artifacts from the working tree
clean:
	@rm -rf ./vendor ./dinghy
