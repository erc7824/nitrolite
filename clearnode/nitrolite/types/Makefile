# Protocol buffer generation
PROTO_DIR = .
PROTO_OUT_DIR = proto

.PHONY: proto
proto: install-protoc-gen-go
	@echo "Generating Go structs from protobuf..."
	protoc \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) \
		$(PROTO_DIR)/*.proto
	@echo "Protobuf generation complete"

.PHONY: install-protoc-gen-go
install-protoc-gen-go:
	@which protoc-gen-go > /dev/null || (echo "Installing protoc-gen-go..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)

.PHONY: clean-proto
clean-proto:
	@echo "Cleaning generated protobuf files..."
	rm -f $(PROTO_OUT_DIR)/*.pb.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto              - Generate Go structs from protobuf files"
	@echo "  install-protoc-gen-go - Install protoc-gen-go if not present"
	@echo "  clean-proto        - Remove generated protobuf files"
	@echo "  help               - Show this help message"