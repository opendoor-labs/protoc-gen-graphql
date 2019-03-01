package descriptor

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type File struct {
	Proto *descriptor.FileDescriptorProto
	// All the protobuf types defined in this file, including nested types.
	Messages []*Message
	Enums    []*Enum
	Services []*Service
}

// Message represents a protobuf message.
type Message struct {
	Proto   *descriptor.DescriptorProto
	Package string
	// nil if Message is a top level message (not nested).
	Parent *Message
	Nested []*Message
	Enums  []*Enum
	// Effective fields of the protobuf message, where fields belonging to a
	// oneof fields are represented as a single field. Note that this
	// behaviour is different from Proto.Field.
	Fields   []*Field
	Oneofs   []*Oneof
	IsMap    bool
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

// Field represents a protobuf field.
type Field struct {
	// nil if IsOneof is true.
	Proto     *descriptor.FieldDescriptorProto
	Parent    *Message
	IsOneof   bool
	OneofName string
}

type Oneof struct {
	Proto       *descriptor.OneofDescriptorProto
	Parent      *Message
	FieldProtos []*descriptor.FieldDescriptorProto
}

type Enum struct {
	Proto   *descriptor.EnumDescriptorProto
	Package string
	// nil if Enum is a top level enum (not nested).
	Parent   *Message
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

type Service struct {
	Proto    *descriptor.ServiceDescriptorProto
	Package  string
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

func WrapFile(proto *descriptor.FileDescriptorProto) *File {
	file := &File{
		Proto: proto,
	}

	for _, serviceProto := range file.Proto.GetService() {
		wrapService(file, serviceProto)
	}
	for _, msgProto := range file.Proto.GetMessageType() {
		wrapMessage(file, msgProto, nil)
	}
	for _, enumProto := range file.Proto.GetEnumType() {
		wrapEnum(file, enumProto, nil)
	}

	return file
}

func wrapService(file *File, proto *descriptor.ServiceDescriptorProto) {
	service := &Service{
		Proto:    proto,
		Package:  file.Proto.GetPackage(),
		TypeName: []string{proto.GetName()},
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), proto.GetName()),
	}
	file.Services = append(file.Services, service)
}

// wrapMessage returns a slice containing the wrapped message and all
// of its nested messages.
func wrapMessage(file *File, proto *descriptor.DescriptorProto, parent *Message) {
	typeName := calculateTypeName(proto.GetName(), parent)
	msg := &Message{
		Proto:    proto,
		Package:  file.Proto.GetPackage(),
		Parent:   parent,
		IsMap:    proto.GetOptions().GetMapEntry(),
		TypeName: typeName,
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), strings.Join(typeName, ".")),
	}
	file.Messages = append(file.Messages, msg)
	if parent != nil {
		parent.Nested = append(parent.Nested, msg)
	}

	wrapFields(msg)
	wrapOneof(msg)
	for _, nested := range proto.GetNestedType() {
		wrapMessage(file, nested, msg)
	}
	for _, enum := range proto.GetEnumType() {
		wrapEnum(file, enum, msg)
	}
}

func wrapFields(parent *Message) {
	seenOneofs := make(map[int32]bool)
	for _, fieldProto := range parent.Proto.GetField() {
		// Handle normal field.
		if fieldProto.OneofIndex == nil {
			parent.Fields = append(parent.Fields, &Field{
				Proto:  fieldProto,
				Parent: parent,
			})
			continue
		}

		// Handle field that belongs to a oneof. We only want to append the oneof field
		// the first time we encounter it.
		index := *fieldProto.OneofIndex
		if seenOneofs[index] {
			continue
		}
		seenOneofs[index] = true

		name := parent.Proto.GetOneofDecl()[index].GetName()
		parent.Fields = append(parent.Fields, &Field{
			Parent:    parent,
			IsOneof:   true,
			OneofName: name,
		})
	}
}

func wrapOneof(parent *Message) {
	for _, oneofProto := range parent.Proto.GetOneofDecl() {
		parent.Oneofs = append(parent.Oneofs, &Oneof{
			Proto:  oneofProto,
			Parent: parent,
		})
	}

	for _, fieldProto := range parent.Proto.GetField() {
		if fieldProto.OneofIndex != nil {
			index := *fieldProto.OneofIndex
			parent.Oneofs[index].FieldProtos = append(parent.Oneofs[index].FieldProtos, fieldProto)
		}
	}
}

func wrapEnum(file *File, proto *descriptor.EnumDescriptorProto, parent *Message) {
	typeName := calculateTypeName(proto.GetName(), parent)
	enum := &Enum{
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

func calculateTypeName(name string, parent *Message) []string {
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
