package descriptor

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	graphqlpb "github.com/martinxsliu/protoc-gen-graphql/protobuf/graphql"
)

type File struct {
	Proto   *descriptor.FileDescriptorProto
	Options *graphqlpb.FileOptions
	// All the protobuf types defined in this file, including nested types.
	Messages []*Message
	Enums    []*Enum
	Services []*Service
}

// Message represents a protobuf message.
type Message struct {
	Proto   *descriptor.DescriptorProto
	Options *graphqlpb.MessageOptions
	Package string
	File    *File
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
	Name string
	// nil if IsProtoOneof is true.
	Proto      *descriptor.FieldDescriptorProto
	Options    *graphqlpb.FieldOptions
	Parent     *Message
	IsOneof    bool
	OneofIndex int32
}

type Oneof struct {
	Proto  *descriptor.OneofDescriptorProto
	Parent *Message
	Fields []*Field
}

type Enum struct {
	Proto   *descriptor.EnumDescriptorProto
	Options *graphqlpb.EnumOptions
	Package string
	File    *File
	// nil if Enum is a top level enum (not nested).
	Parent   *Message
	Values   []*EnumValue
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
}

type EnumValue struct {
	Proto   *descriptor.EnumValueDescriptorProto
	Options *graphqlpb.EnumValueOptions
}

type Service struct {
	Proto    *descriptor.ServiceDescriptorProto
	Options  *graphqlpb.ServiceOptions
	Package  string
	File     *File
	TypeName []string
	// Fully qualified name starting with a '.' including the package name.
	FullName string
	Methods  []*Method
}

type Method struct {
	Proto   *descriptor.MethodDescriptorProto
	Options *graphqlpb.MethodOptions
}

func WrapFile(proto *descriptor.FileDescriptorProto) *File {
	file := &File{
		Proto:   proto,
		Options: getFileOptions(proto),
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
		Options:  getServiceOptions(proto),
		Package:  file.Proto.GetPackage(),
		File:     file,
		TypeName: []string{proto.GetName()},
		FullName: fmt.Sprintf(".%s.%s", file.Proto.GetPackage(), proto.GetName()),
	}
	wrapMethods(service)
	file.Services = append(file.Services, service)
}

func wrapMethods(service *Service) {
	for _, proto := range service.Proto.GetMethod() {
		service.Methods = append(service.Methods, &Method{
			Proto:   proto,
			Options: getMethodOptions(proto),
		})
	}
}

// wrapMessage returns a slice containing the wrapped message and all
// of its nested messages.
func wrapMessage(file *File, proto *descriptor.DescriptorProto, parent *Message) {
	typeName := calculateTypeName(proto.GetName(), parent)
	msg := &Message{
		Proto:    proto,
		Options:  getMessageOptions(proto),
		Package:  file.Proto.GetPackage(),
		File:     file,
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
	wrapOneofs(msg)
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
				Name:    fieldProto.GetName(),
				Proto:   fieldProto,
				Options: getFieldOptions(fieldProto),
				Parent:  parent,
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
			Name:       name,
			Parent:     parent,
			IsOneof:    true,
			OneofIndex: index,
		})
	}
}

func wrapOneofs(parent *Message) {
	for _, oneofProto := range parent.Proto.GetOneofDecl() {
		parent.Oneofs = append(parent.Oneofs, &Oneof{
			Proto:  oneofProto,
			Parent: parent,
		})
	}

	for _, fieldProto := range parent.Proto.GetField() {
		if fieldProto.OneofIndex != nil {
			index := *fieldProto.OneofIndex
			parent.Oneofs[index].Fields = append(parent.Oneofs[index].Fields, &Field{
				Name:    fieldProto.GetName(),
				Proto:   fieldProto,
				Options: getFieldOptions(fieldProto),
				Parent:  parent,
			})
		}
	}
}

func wrapEnum(file *File, proto *descriptor.EnumDescriptorProto, parent *Message) {
	typeName := calculateTypeName(proto.GetName(), parent)

	var values []*EnumValue
	for _, valueProto := range proto.GetValue() {
		values = append(values, &EnumValue{
			Proto:   valueProto,
			Options: getEnumValueOptions(valueProto),
		})
	}

	enum := &Enum{
		Proto:    proto,
		Options:  getEnumOptions(proto),
		Package:  file.Proto.GetPackage(),
		File:     file,
		Parent:   parent,
		Values:   values,
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
