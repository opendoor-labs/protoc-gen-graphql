.PHONY: build
build: protoc
	go build -o bin/protoc-gen-graphql *.go

.PHONY: install
install: protoc
	go install .

.PHONY: test
test: build
	go test ./...

.PHONY: protoc
protoc:
	protoc -I protobuf --go_out=paths=source_relative:protobuf protobuf/graphql/*.proto

.PHONY: protoc-wkt
protoc-wkt: build
	protoc -I protobuf \
		--plugin=bin/protoc-gen-graphql \
		--graphql_out=input_mode=all:protobuf \
		protobuf/google/protobuf/**/*.proto
