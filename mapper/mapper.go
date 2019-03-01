package mapper

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/martinxsliu/protoc-gen-graphql/graphql"
	"github.com/martinxsliu/protoc-gen-graphql/graphqlpb"

	"github.com/martinxsliu/protoc-gen-graphql/descriptor"
)

type Mapper struct {
	FilePbs []*pb.FileDescriptorProto
	Params  *Parameters

	// Maps file names to descriptors.
	Files map[string]*descriptor.File
	// Maps qualified protobuf names to descriptors.
	Messages map[string]*descriptor.Message
	Enums    map[string]*descriptor.Enum

	// Set of protobuf messages with no fields. Values are always true.
	EmptyMessages map[string]bool

	// Maps protobuf types to graphql types.
	// e.m. ".google.protobuf.StringValue" -> "GoogleProtobuf_StringValue"
	MessageMappers map[string]*MessageMapper
	EnumMappers    map[string]*EnumMapper
	ServiceMappers map[string]*ServiceMapper
}

type MessageMapper struct {
	Descriptor *descriptor.Message
	Object     *graphql.Object
	Input      *graphql.Input
	Oneofs     []*OneofMapper
}

type OneofMapper struct {
	Descriptor *descriptor.Oneof
	Union      *graphql.Union
	Objects    []*graphql.Object
	Input      *graphql.Input
}

type EnumMapper struct {
	Descriptor *descriptor.Enum
	Enum       *graphql.Enum
}

type ServiceMapper struct {
	Descriptor          *descriptor.Service
	QueriesObject       *graphql.Object
	MutationsObject     *graphql.Object
	SubscriptionsObject *graphql.Object
}

// New creates a new Mapper with all mappings populated from the provided file
// descriptors. The provided file descriptors must be in topological order.
func New(filePbs []*pb.FileDescriptorProto, params *Parameters) *Mapper {
	m := &Mapper{
		FilePbs:        filePbs,
		Params:         params,
		Files:          make(map[string]*descriptor.File),
		Messages:       make(map[string]*descriptor.Message),
		Enums:          make(map[string]*descriptor.Enum),
		EmptyMessages:  make(map[string]bool),
		MessageMappers: make(map[string]*MessageMapper),
		EnumMappers:    make(map[string]*EnumMapper),
		ServiceMappers: make(map[string]*ServiceMapper),
	}

	for _, filePb := range filePbs {
		file := descriptor.WrapFile(filePb)

		// Build descriptor maps.
		m.Files[filePb.GetName()] = file
		for _, message := range file.Messages {
			m.Messages[message.FullName] = message
		}
		for _, enum := range file.Enums {
			m.Enums[enum.FullName] = enum
		}

		// Build protobuf to graphql mappers.

		// Build enum mapper first as it has no dependencies.
		for _, enum := range file.Enums {
			m.buildEnumMapper(enum)
		}

		// Build message mapper, first sort messages in topological order.
		g := NewGraph(file.Messages)
		messages, err := g.Sort()
		if err != nil {
			panic(err)
		}

		for _, message := range messages {
			m.buildMessageMapper(message, false)
		}

		for _, service := range file.Services {
			// Build inputs for service methods.
			var inputs []*descriptor.Message
			for _, method := range service.Proto.GetMethod() {
				inputs = append(inputs, m.Messages[method.GetInputType()])
			}

			inputMessages, err := g.SortTo(inputs)
			if err != nil {
				panic(err)
			}

			for _, message := range inputMessages {
				m.buildMessageMapper(message, true)
			}

			// Build service mapper last, after all dependencies are mapped.
			m.buildServiceMapper(service)
		}
	}

	return m
}

func (m *Mapper) buildMessageMapper(message *descriptor.Message, input bool) {
	if len(message.Fields) == 0 {
		m.EmptyMessages[message.FullName] = true
		m.MessageMappers[message.FullName] = &MessageMapper{Descriptor: message}
		return
	}

	mapper := &MessageMapper{
		Descriptor: message,
		Object: &graphql.Object{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:    message.Package,
				TypeName:   message.TypeName,
				Input:      false,
				IsProtoMap: message.IsMap,
			}),
			Fields: m.graphqlFields(message, false),
		},
	}

	if input {
		mapper.Input = &graphql.Input{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:    message.Package,
				TypeName:   message.TypeName,
				Input:      true,
				IsProtoMap: message.IsMap,
			}),
			Fields: m.graphqlFields(message, true),
		}
	}

	for _, oneof := range message.Oneofs {
		mapper.Oneofs = append(mapper.Oneofs, m.buildOneofMapper(oneof, input))
	}

	m.MessageMappers[message.FullName] = mapper
}

func (m *Mapper) graphqlFields(message *descriptor.Message, input bool) []*graphql.Field {
	var fields []*graphql.Field
	for _, field := range message.Fields {
		if field.IsOneof {
			fields = append(fields, &graphql.Field{
				Name: field.OneofName,
				TypeName: BuildGraphqlTypeName(&GraphqlTypeNameParts{
					Package:  message.Package,
					TypeName: append(message.TypeName, field.OneofName),
					Input:    input,
				}),
			})
			continue
		}

		fields = append(fields, m.graphqlField(field.Proto, fieldOptions{Input: input}))
	}
	return fields
}

type fieldOptions struct {
	Input           bool
	NullableScalars bool
}

func (m *Mapper) graphqlField(proto *pb.FieldDescriptorProto, options fieldOptions) *graphql.Field {
	field := &graphql.Field{
		Name: proto.GetName(),
	}

	switch proto.GetType() {
	case pb.FieldDescriptorProto_TYPE_FLOAT, pb.FieldDescriptorProto_TYPE_DOUBLE,
		pb.FieldDescriptorProto_TYPE_UINT32, pb.FieldDescriptorProto_TYPE_SINT32,
		pb.FieldDescriptorProto_TYPE_FIXED32, pb.FieldDescriptorProto_TYPE_SFIXED32:

		field.TypeName = graphql.ScalarFloat.TypeName()
		field.Modifiers = graphql.TypeModifierNonNull

	case pb.FieldDescriptorProto_TYPE_STRING, pb.FieldDescriptorProto_TYPE_BYTES,
		pb.FieldDescriptorProto_TYPE_INT64, pb.FieldDescriptorProto_TYPE_UINT64, pb.FieldDescriptorProto_TYPE_SINT64,
		pb.FieldDescriptorProto_TYPE_FIXED64, pb.FieldDescriptorProto_TYPE_SFIXED64:

		field.TypeName = graphql.ScalarString.TypeName()
		if !options.NullableScalars {
			field.Modifiers = graphql.TypeModifierNonNull
		}

	case pb.FieldDescriptorProto_TYPE_INT32:
		field.TypeName = graphql.ScalarInt.TypeName()
		if !options.NullableScalars {
			field.Modifiers = graphql.TypeModifierNonNull
		}

	case pb.FieldDescriptorProto_TYPE_BOOL:
		field.TypeName = graphql.ScalarBoolean.TypeName()
		if !options.NullableScalars {
			field.Modifiers = graphql.TypeModifierNonNull
		}

	case pb.FieldDescriptorProto_TYPE_ENUM:
		field.TypeName = m.EnumMappers[proto.GetTypeName()].Enum.Name
		if !options.NullableScalars {
			field.Modifiers = graphql.TypeModifierNonNull
		}

	case pb.FieldDescriptorProto_TYPE_MESSAGE:
		if m.EmptyMessages[proto.GetTypeName()] {
			field.TypeName = graphql.ScalarBoolean.TypeName()
			break
		}

		if options.Input {
			if m.MessageMappers[proto.GetTypeName()].Input == nil {
				panic(fmt.Sprintf("%s: %+v", proto.GetTypeName(), m.MessageMappers[proto.GetTypeName()]))
			}
			field.TypeName = m.MessageMappers[proto.GetTypeName()].Input.Name
		} else {
			if m.MessageMappers[proto.GetTypeName()] == nil {
				panic(fmt.Sprintf("%s: %v", proto.GetTypeName(), m.MessageMappers))
			}
			field.TypeName = m.MessageMappers[proto.GetTypeName()].Object.Name
		}

		// Map elements are non-nullable.
		if m.Messages[proto.GetTypeName()].IsMap {
			field.Modifiers = graphql.TypeModifierNonNull
		}

	default:
		panic(fmt.Sprintf("unexpected protobuf descriptor type: %s", proto.GetType().String()))
	}

	if proto.GetLabel() == pb.FieldDescriptorProto_LABEL_REPEATED {
		field.Modifiers = field.Modifiers | graphql.TypeModifierList
	}

	return m.graphqlSpecialTypes(field, proto.GetTypeName())
}

func (m *Mapper) graphqlSpecialTypes(field *graphql.Field, protoTypeName string) *graphql.Field {
	if protoTypeName == ".google.protobuf.Timestamp" && m.Params.TimestampTypeName != "" {
		field.TypeName = m.Params.TimestampTypeName
	}
	if protoTypeName == ".google.protobuf.Duration" && m.Params.DurationTypeName != "" {
		field.TypeName = m.Params.DurationTypeName
	}

	if m.Params.WrappersAsNull {
		switch protoTypeName {
		case ".google.protobuf.FloatValue", ".google.protobuf.DoubleValue", ".google.protobuf.UInt32Value":
			field.TypeName = graphql.ScalarFloat.TypeName()
		case ".google.protobuf.StringValue", ".google.protobuf.BytesValue", ".google.protobuf.Int64Value", ".google.protobuf.UInt64Value":
			field.TypeName = graphql.ScalarString.TypeName()
		case ".google.protobuf.Int32Value":
			field.TypeName = graphql.ScalarInt.TypeName()
		case ".google.protobuf.BoolValue":
			field.TypeName = graphql.ScalarBoolean.TypeName()
		}
	}

	return field
}

func (m *Mapper) buildOneofMapper(oneof *descriptor.Oneof, input bool) *OneofMapper {
	mapper := &OneofMapper{
		Union: &graphql.Union{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:  oneof.Parent.Package,
				TypeName: append(oneof.Parent.TypeName, oneof.Proto.GetName()),
			}),
		},
	}

	for _, fieldProto := range oneof.FieldProtos {
		typeName := BuildGraphqlTypeName(&GraphqlTypeNameParts{
			Package:  oneof.Parent.Package,
			TypeName: append(oneof.Parent.TypeName, oneof.Proto.GetName(), fieldProto.GetName()),
		})

		mapper.Union.TypeNames = append(mapper.Union.TypeNames, typeName)
		mapper.Objects = append(mapper.Objects, &graphql.Object{
			Name:   typeName,
			Fields: []*graphql.Field{m.graphqlField(fieldProto, fieldOptions{})},
		})
	}

	if !input {
		return mapper
	}

	var inputFields []*graphql.Field
	for _, fieldProto := range oneof.FieldProtos {
		inputFields = append(inputFields, m.graphqlField(fieldProto, fieldOptions{Input: true, NullableScalars: true}))
	}

	mapper.Input = &graphql.Input{
		Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
			Package:  oneof.Parent.Package,
			TypeName: append(oneof.Parent.TypeName, oneof.Proto.GetName()),
			Input:    true,
		}),
		Fields: inputFields,
	}

	return mapper
}

func (m *Mapper) buildEnumMapper(enum *descriptor.Enum) {
	var values []string
	for _, protoValue := range enum.Proto.GetValue() {
		values = append(values, protoValue.GetName())
	}

	m.EnumMappers[enum.FullName] = &EnumMapper{
		Descriptor: enum,
		Enum: &graphql.Enum{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:  enum.Package,
				TypeName: enum.TypeName,
			}),
			Values: values,
		},
	}
}

func (m *Mapper) buildServiceMapper(service *descriptor.Service) {
	var (
		queries       []*graphql.Field
		mutations     []*graphql.Field
		subscriptions []*graphql.Field
	)

	for _, method := range service.Proto.GetMethod() {
		var operation string
		if proto.HasExtension(method.GetOptions(), graphqlpb.E_Operation) {
			extVal, err := proto.GetExtension(method.GetOptions(), graphqlpb.E_Operation)
			if err != nil {
				panic(err)
			}
			operation = *extVal.(*string)
		}
		if operation == "none" {
			return
		}

		field := m.graphqlFieldFromMethod(method)

		switch operation {
		case "mutation":
			mutations = append(mutations, field)
		case "subscription":
			subscriptions = append(subscriptions, field)
		default:
			queries = append(queries, field)
		}
	}

	mapper := &ServiceMapper{
		Descriptor: service,
	}
	if len(queries) > 0 {
		mapper.QueriesObject = &graphql.Object{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:  service.Package,
				TypeName: append(service.TypeName, "Query"),
			}),
			Fields: queries,
		}
	}
	if len(mutations) > 0 {
		mapper.MutationsObject = &graphql.Object{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:  service.Package,
				TypeName: append(service.TypeName, "Mutation"),
			}),
			Fields: mutations,
		}
	}
	if len(subscriptions) > 0 {
		mapper.SubscriptionsObject = &graphql.Object{
			Name: BuildGraphqlTypeName(&GraphqlTypeNameParts{
				Package:  service.Package,
				TypeName: append(service.TypeName, "Subscription"),
			}),
			Fields: subscriptions,
		}
	}

	m.ServiceMappers[service.FullName] = mapper
}

func (m *Mapper) graphqlFieldFromMethod(method *pb.MethodDescriptorProto) *graphql.Field {
	// Only add an argument if there are fields in the gRPC request message.
	var arguments []*graphql.Argument
	inputType := m.Messages[method.GetInputType()]
	if m.MessageMappers[method.GetInputType()].Input == nil {
		panic(fmt.Sprintf("%s: %+v", method.GetInputType(), m.MessageMappers[method.GetInputType()]))
	}
	if len(inputType.Fields) != 0 {
		arguments = append(arguments, &graphql.Argument{
			Name:      "input",
			TypeName:  m.MessageMappers[method.GetInputType()].Input.Name,
			Modifiers: graphql.TypeModifierNonNull,
		})
	}

	// If the response message has no fields then return a nullable Boolean.
	// It is up to the resolver's implementation whether or not to return an
	// actual boolean value or default to null.
	outputType := m.Messages[method.GetOutputType()]
	if len(outputType.Fields) == 0 {
		return &graphql.Field{
			Name:      method.GetName(),
			TypeName:  graphql.ScalarBoolean.TypeName(),
			Arguments: arguments,
		}
	}

	return &graphql.Field{
		Name:      method.GetName(),
		TypeName:  m.MessageMappers[method.GetOutputType()].Object.Name,
		Arguments: arguments,
		Modifiers: graphql.TypeModifierNonNull,
	}
}

type GraphqlTypeNameParts struct {
	Package    string
	TypeName   []string
	IsProtoMap bool
	Input      bool
}

func BuildGraphqlTypeName(parts *GraphqlTypeNameParts) string {
	var b strings.Builder
	b.WriteString(generator.CamelCaseSlice(strings.Split(parts.Package, ".")))
	for i, name := range parts.TypeName {
		if parts.IsProtoMap && i == len(parts.TypeName)-1 {
			name = strings.TrimSuffix(name, "Entry")
		}

		b.WriteString("_")
		b.WriteString(generator.CamelCase(name))
	}
	if parts.Input {
		b.WriteString("_Input")
	}
	return b.String()
}
