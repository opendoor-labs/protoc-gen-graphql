.PHONY: build
build: protoc
	GO111MODULE=on go build -o bin/protoc-gen-graphql *.go

.PHONY: install
install: protoc
	GO111MODULE=on go install .

.PHONY: test
test: build
	find testdata -name "*.graphql" -type f -delete
	GO111MODULE=on go test ./...

.PHONY: protoc
protoc:
	protoc -I protobuf --go_out=paths=source_relative:protobuf protobuf/graphql/*.proto

.PHONY: protoc-wkt
protoc-wkt: build
	protoc -I protobuf \
		--plugin=bin/protoc-gen-graphql \
		--graphql_out=input_mode=all:protobuf \
		protobuf/google/protobuf/**/*.proto
