.PHONY: build test protoc install

build:
	go build ./...

test:
	go test ./...

protoc:
	protoc --go_out=. --go_opt=module=github.com/JasperLabs/grpcerrors \
		-I proto \
		proto/errors/errors.proto

install:
	go install ./cmd/protoc-gen-go-errors/
