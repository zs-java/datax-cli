BIN_NAME=datax-cli
USR_BIN_PATH=/usr/local/bin

build:
	@go build -o $(BIN_NAME) .

build-all:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BIN_NAME)-linux-amd64 main.go && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/$(BIN_NAME)-darwin-amd64 main.go && \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/$(BIN_NAME)-win-amd64.exe main.go

install: $(BIN_NAME)
	@cp $(BIN_NAME) $(USR_BIN_PATH)

uninstall: $(USR_BIN_PATH)/$(BIN_NAME)
	@rm -rf $(USR_BIN_PATH)/$(BIN_NAME)

clean:
	@rm -rf ./dist
	@rm -rf ./logs
	@rm -rf ./bin
	@rm -rf $(BIN_NAME)

.PHONY: build build-all clean