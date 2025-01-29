BIN_NAME := lnkr

test:
	@echo "Running go test"
	go test ./...

build:
	@echo "Running go build"
	go build -o $(BIN_NAME) ./cmd/

clean:
	rm $(BIN_NAME)
