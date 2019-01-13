.PHONY: build
build: protoc
	go build -o bin/protoc-gen-graphql *.go

.PHONY: install
install: protoc
	go install .

.PHONY: test
test: build
	ginkgo ./...

.PHONY: protoc
protoc:
	protoc -I graphqlpb --go_out=paths=source_relative:graphqlpb graphqlpb/*.proto

.PHONY: protoc-wkt
protoc-wkt: build
	protoc -I graphqlpb \
		--plugin=bin/protoc-gen-graphql \
		--graphql_out=graphqlpb \
		graphqlpb/google/protobuf/*.proto
