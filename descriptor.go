package main

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type FileDescriptor struct {
	Proto    *descriptor.FileDescriptorProto
	Messages []*MessageDescriptor
	Enums    []*EnumDescriptor
	Services []*ServiceDescriptor
}

// MessageDescriptor represents a protobuf message.
type MessageDescriptor struct {
	Proto    *descriptor.DescriptorProto
	Package  string
	Parent   *MessageDescriptor
	Nested   []*MessageDescriptor
	Enums    []*EnumDescriptor
	IsMap    bool
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

type EnumDescriptor struct {
	Proto    *descriptor.EnumDescriptorProto
	Package  string
	Parent   *MessageDescriptor
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

type ServiceDescriptor struct {
	Proto    *descriptor.ServiceDescriptorProto
	File     *FileDescriptor
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

func wrapFile(proto *descriptor.FileDescriptorProto) *FileDescriptor {
	file := &FileDescriptor{
		Proto: proto,
	}

	for _, serviceProto := range file.Proto.GetService() {
		wrapServiceDescriptor(file, serviceProto)
	}
	for _, msgProto := range file.Proto.GetMessageType() {
		wrapMessageDescriptor(file, msgProto, nil)
	}
	for _, enumProto := range file.Proto.GetEnumType() {
		wrapEnumDescriptor(file, enumProto, nil)
	}

	return file
}

func wrapServiceDescriptor(file *FileDescriptor, proto *descriptor.ServiceDescriptorProto) {
	service := &ServiceDescriptor{
		Proto:    proto,
		File:     file,
		TypeName: []string{proto.GetName()},
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), proto.GetName()),
	}
	file.Services = append(file.Services, service)
}

// wrapMessageDescriptor returns a slice containing the wrapped message and all
// of its nested messages.
func wrapMessageDescriptor(file *FileDescriptor, proto *descriptor.DescriptorProto, parent *MessageDescriptor) {
	typeName := calculateTypeName(proto.GetName(), parent)
	msg := &MessageDescriptor{
		Proto:    proto,
		Package:  file.Proto.GetPackage(),
		Parent:   parent,
		Nested:   []*MessageDescriptor{},
		IsMap:    proto.GetOptions().GetMapEntry(),
		TypeName: typeName,
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), strings.Join(typeName, ".")),
	}
	file.Messages = append(file.Messages, msg)
	if parent != nil {
		parent.Nested = append(parent.Nested, msg)
	}

	for _, nested := range proto.GetNestedType() {
		wrapMessageDescriptor(file, nested, msg)
	}
	for _, enum := range proto.GetEnumType() {
		wrapEnumDescriptor(file, enum, msg)
	}
}

func wrapEnumDescriptor(file *FileDescriptor, proto *descriptor.EnumDescriptorProto, parent *MessageDescriptor) {
	typeName := calculateTypeName(proto.GetName(), parent)
	enum := &EnumDescriptor{
		Proto:    proto,
		Package:  file.Proto.GetPackage(),
		Parent:   parent,
		TypeName: typeName,
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), strings.Join(typeName, ".")),
	}
	file.Enums = append(file.Enums, enum)
	if parent != nil {
		parent.Enums = append(parent.Enums, enum)
	}
}

func calculateTypeName(name string, parent *MessageDescriptor) []string {
	parts := []string{name}
	for ; parent != nil; parent = parent.Parent {
		parts = append(parts, parent.Proto.GetName())
	}
	for i := 0; i < len(parts)/2; i++ {
		j := len(parts) - 1 - i
		parts[i], parts[j] = parts[j], parts[i]
	}
	return parts
}
