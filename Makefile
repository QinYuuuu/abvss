PROJECT_NAME := github.com/QinYuuuu/abvss

.PHONY: init

init:
	@echo "initialize Go module..."
	@go mod init $(PROJECT_NAME)
	@go mod tidy
	@echo "Go module initialization finished"

proto_gen:
	@cd pkg/protobuf
	@protoc --go_out=. --go-grpc_out=. *.proto