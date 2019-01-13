package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func findProtoFiles(name string) []string {
	protos, err := filepath.Glob(filepath.Join("testdata", name, "*.proto"))
	Expect(err).NotTo(HaveOccurred())
	Expect(protos).NotTo(BeEmpty())
	return protos
}

func runProtoc(protoFiles []string, parameter string) {
	args := append([]string{
		"-I", "testdata",
		"--plugin=bin/protoc-gen-graphql",
		fmt.Sprintf("--graphql_out=%s:testdata", parameter),
	}, protoFiles...)
	cmd := exec.Command("protoc", args...)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
}

func itGeneratesTheCorrectOutput(name string) {
	protoFiles := findProtoFiles(name)
	runProtoc(protoFiles, "")

	for _, proto := range protoFiles {
		graphql, err := ioutil.ReadFile(strings.TrimSuffix(proto, ".proto") + "_pb.graphql")
		Expect(err).NotTo(HaveOccurred())
		expected, err := ioutil.ReadFile(strings.TrimSuffix(proto, ".proto") + ".golden")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(graphql)).To(Equal(string(expected)))
	}
}

var _ = Describe("Plugin", func() {
	It("generates basic protobuf types", func() {
		itGeneratesTheCorrectOutput("basic")
	})

	It("generates input types for gRPC services", func() {
		itGeneratesTheCorrectOutput("grpc")
	})
})
