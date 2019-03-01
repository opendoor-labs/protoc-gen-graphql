package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fail("error reading input: " + err.Error())
	}

	req := &plugin.CodeGeneratorRequest{}
	resp := &plugin.CodeGeneratorResponse{}

	if err := proto.Unmarshal(input, req); err != nil {
		fail("error parsing input: " + err.Error())
	}

	generator := New(req, resp)

	err = generator.Generate()
	if err != nil {
		writeErr(resp, err)
	}

	writeOutput(resp)
}

func writeErr(resp *plugin.CodeGeneratorResponse, err error) {
	msg := err.Error()
	resp.Error = &msg
	writeOutput(resp)
	os.Exit(0)
}

func writeOutput(resp *plugin.CodeGeneratorResponse) {
	output, err := proto.Marshal(resp)
	if err != nil {
		fail("error serializing output: " + err.Error())
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		fail("error writing output: " + err.Error())
	}
}

func fail(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "protoc-gen-graphql: %s", msg)
	os.Exit(1)
}
