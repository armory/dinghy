
all: build test

# Build the Dinghy binary
build: dinghy.go cmd pkg
	@go build -v .

# Test this project
test: build
	@go test -v -race -covermode atomic -coverprofile=profile.cov ./...

run: build
	@./dinghy

# Remove build artifacts from the working tree
clean:
	@rm -rf ./dinghy
