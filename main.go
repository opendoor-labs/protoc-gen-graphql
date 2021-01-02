package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	opts := protogen.Options{}
	opts.Run(func(gen *protogen.Plugin) error {
		return New(gen).Generate()
	})
}
