package mapper

import (
	"github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"

	graphqlpb "github.com/martinxsliu/protoc-gen-graphql/protobuf/graphql"
)

func getServiceOptions(service *pb.ServiceDescriptorProto) *graphqlpb.ServiceOptions {
	options := service.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_ServiceOptions) {
		ext, err := proto.GetExtension(options, graphqlpb.E_ServiceOptions)
		if err != nil {
			panic(err)
		}
		return ext.(*graphqlpb.ServiceOptions)
	}
	return &graphqlpb.ServiceOptions{}
}

func getMessageOptions(method *pb.MethodDescriptorProto) *graphqlpb.MethodOptions {
	options := method.GetOptions()
	if proto.HasExtension(options, graphqlpb.E_MethodOptions) {
		ext, err := proto.GetExtension(options, graphqlpb.E_MethodOptions)
		if err != nil {
			panic(err)
		}
		return ext.(*graphqlpb.MethodOptions)
	}
	return &graphqlpb.MethodOptions{}
}
