.PHONY: clean test pbx

BIN_DIR := bin

all: lint pbx

clean:
	@rm -rf $(BIN_DIR)

lint:
	@gofmt -w .
	@go fix ./...

test:
	@go mod tidy
	@go mod verify
	@go vet ./...
	@go test -v ./test

pbx:
	@mkdir -p $(BIN_DIR)
	@go build -ldflags "-s -w" -o $(BIN_DIR)/$@ ./cmd/$@/...
