build:
	go build -o bin/soroboru main.go

help:
	@echo "Makefile for Soroboru"
	@echo ""
	@echo "Available targets:"
	@echo "  build         Build the Soroboru binary"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make help"

run: build
	./bin/soroboru --name test
	cd test
	go fmt
	go mod tidy