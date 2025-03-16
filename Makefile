PROJECT_NAME := github.com/QinYuuuu/abvss

.PHONY: init

init:
	@echo "initialize Go module..."
	@go mod init $(PROJECT_NAME)
	@go mod tidy
	@echo "Go module initialization finished"
