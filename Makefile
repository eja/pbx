.PHONY: clean test pbx

BUILD_DIR := build

all: lint pbx

clean:
	@rm -rf $(BUILD_DIR)

lint:
	@gofmt -w .
	@go fix ./...

test:
	@go mod tidy
	@go mod verify
	@go vet ./...
	@go test -v ./test

pbx:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-s -w" -o $(BUILD_DIR)/$@ ./cmd/$@/...
