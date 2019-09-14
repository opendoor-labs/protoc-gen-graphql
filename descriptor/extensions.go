package descriptor

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"

	graphqlpb "github.com/martinxsliu/protoc-gen-graphql/protobuf/graphql"
)

func getFileOptions(file *pb.FileDescriptorProto) *graphqlpb.FileOptions {
	options := file.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_File) {
		ext, err := proto.GetExtension(options, graphqlpb.E_File)
		if err != nil {
			panic(fmt.Sprintf("error getting file options: %s", err.Error()))
		}
		return ext.(*graphqlpb.FileOptions)
	}
	return &graphqlpb.FileOptions{}
}

func getMessageOptions(message *pb.DescriptorProto) *graphqlpb.MessageOptions {
	options := message.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_Message) {
		ext, err := proto.GetExtension(options, graphqlpb.E_Message)
		if err != nil {
			panic(fmt.Sprintf("error getting message options: %s", err.Error()))
		}
		return ext.(*graphqlpb.MessageOptions)
	}
	return &graphqlpb.MessageOptions{}
}

func getFieldOptions(field *pb.FieldDescriptorProto) *graphqlpb.FieldOptions {
	options := field.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_Field) {
		ext, err := proto.GetExtension(options, graphqlpb.E_Field)
		if err != nil {
			panic(fmt.Sprintf("error getting field options: %s", err.Error()))
		}
		return ext.(*graphqlpb.FieldOptions)
	}
	return &graphqlpb.FieldOptions{}
}

func getEnumOptions(enum *pb.EnumDescriptorProto) *graphqlpb.EnumOptions {
	options := enum.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_Enum) {
		ext, err := proto.GetExtension(options, graphqlpb.E_Enum)
		if err != nil {
			panic(fmt.Sprintf("error getting enum options: %s", err.Error()))
		}
		return ext.(*graphqlpb.EnumOptions)
	}
	return &graphqlpb.EnumOptions{}
}

func getEnumValueOptions(enumValue *pb.EnumValueDescriptorProto) *graphqlpb.EnumValueOptions {
	options := enumValue.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_EnumValue) {
		ext, err := proto.GetExtension(options, graphqlpb.E_EnumValue)
		if err != nil {
			panic(fmt.Sprintf("error getting enum value options: %s", err.Error()))
		}
		return ext.(*graphqlpb.EnumValueOptions)
	}
	return &graphqlpb.EnumValueOptions{}
}

func getServiceOptions(service *pb.ServiceDescriptorProto) *graphqlpb.ServiceOptions {
	options := service.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_Service) {
		ext, err := proto.GetExtension(options, graphqlpb.E_Service)
		if err != nil {
			panic(fmt.Sprintf("error getting service options: %s", err.Error()))
		}
		return ext.(*graphqlpb.ServiceOptions)
	}
	return &graphqlpb.ServiceOptions{}
}

func getMethodOptions(method *pb.MethodDescriptorProto) *graphqlpb.MethodOptions {
	options := method.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_Method) {
		ext, err := proto.GetExtension(options, graphqlpb.E_Method)
		if err != nil {
			panic(fmt.Sprintf("error getting method options: %s", err.Error()))
		}
		return ext.(*graphqlpb.MethodOptions)
	}
	return &graphqlpb.MethodOptions{}
}
