PROJECT_NAME := github.com/QinYuuuu/abvss

.PHONY: init

init:
	@echo "initialize Go module..."
	@go mod init $(PROJECT_NAME)
	@go mod tidy
	@echo "Go module initialization finished"

proto_gen:
	cd pkg/protobuf && protoc --go_out=. --go_opt=paths=source_relative Message.proto

release:
	cd harts/cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o harts-DKG-test main.go && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o harts-DKG-test.exe main.go