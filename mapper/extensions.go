package mapper

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"

	graphqlpb "github.com/martinxsliu/protoc-gen-graphql/protobuf/graphql"
)

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

func getMessageOptions(method *pb.MethodDescriptorProto) *graphqlpb.MethodOptions {
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
