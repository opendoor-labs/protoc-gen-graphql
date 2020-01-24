BINARY := bin/protoc-gen-graphql

GRAPHQL_PROTOS := $(wildcard protobuf/graphql/*.proto)
GRAPHQL_PROTOS_GO_SRC := $(patsubst %.proto,%.pb.go,$(GRAPHQL_PROTOS))

GO_SRC := $(wildcard */*.go) $(wildcard *.go)

.PHONY: build
build: $(BINARY)

$(BINARY): protoc $(GO_SRC)
	GO111MODULE=on go build -o $@ *.go

.PHONY: install
install: protoc $(GO_SRC)
	GO111MODULE=on go install .

.PHONY: test
test: build
	find testdata -name "*.graphql" -type f -delete
	GO111MODULE=on go test -v ./...

.PHONY: protoc
protoc: $(GRAPHQL_PROTOS_GO_SRC)

$(GRAPHQL_PROTOS_GO_SRC): $(GRAPHQL_PROTOS)
	protoc -I protobuf --go_out=paths=source_relative:protobuf $^

.PHONY: protoc-wkt
protoc-wkt: build
	protoc -I protobuf \
		--plugin=$(BINARY) \
		--graphql_out=input_mode=all:protobuf \
		protobuf/google/protobuf/*.proto
