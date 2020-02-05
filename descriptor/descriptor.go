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
	Comments string
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
	ForeignKey *ForeignKey
	Comments   string
}

type ForeignKey struct {
	// Fully qualified name starting with a '.' including the package name.
	FullName  string
	FieldName string
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
	Comments string
}

type EnumValue struct {
	Proto    *descriptor.EnumValueDescriptorProto
	Options  *graphqlpb.EnumValueOptions
	Comments string
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
	Comments string
}

type Method struct {
	Proto    *descriptor.MethodDescriptorProto
	Options  *graphqlpb.MethodOptions
	Service  *Service
	Loaders  []*Loader
	Comments string
}

type Loader struct {
	// Fully qualified name of the loaded message starting with a '.' including the package name.
	FullName          string
	Many              bool
	RequestFieldPath  []string
	ResponseFieldPath []string
	ObjectKeyPath []string
	Method            *Method
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
	setComments(file)

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
		options := getMethodOptions(proto)
		method := &Method{
			Proto:   proto,
			Options: options,
			Service: service,
		}
		if loader := getLoaderOption(method, options.GetLoadOne(), false); loader != nil {
			method.Loaders = append(method.Loaders, loader)
		}
		if loader := getLoaderOption(method, options.GetLoadMany(), true); loader != nil {
			method.Loaders = append(method.Loaders, loader)
		}
		service.Methods = append(service.Methods, method)
	}
}

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
			options := getFieldOptions(fieldProto)
			parent.Fields = append(parent.Fields, &Field{
				Name:       fieldProto.GetName(),
				Proto:      fieldProto,
				Options:    options,
				Parent:     parent,
				ForeignKey: getForeignKeyOption(options.GetForeignKey()),
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

func setComments(file *File) {
	for _, location := range file.Proto.GetSourceCodeInfo().GetLocation() {
		if location.GetLeadingComments() == "" && location.GetTrailingComments() == "" {
			continue
		}

		// We need at least 2 elements in the path to describe a definition.
		// The first identifies the file's field number, and the second identifies
		// the index within the field.
		if len(location.Path) < 2 {
			continue
		}

		switch location.Path[0] {
		case 4: // Message
			messageProto := file.Proto.MessageType[location.Path[1]]
			setMessageComments(file, messageProto, location, location.Path[2:])
		case 5: // Enum
			enumProto := file.Proto.EnumType[location.Path[1]]
			setEnumComments(file, enumProto, location, location.Path[2:])
		case 6: // Service
			var service *Service
			serviceProto := file.Proto.Service[location.Path[1]]
			for _, s := range file.Services {
				if s.Proto == serviceProto {
					service = s
				}
			}

			if len(location.Path) == 2 {
				// This is a comment for the service.
				service.Comments = combineComments(location)
				continue
			}

			switch location.Path[2] {
			case 2: // Method
				var method *Method
				enumValueProto := service.Proto.Method[location.Path[3]]
				for _, m := range service.Methods {
					if m.Proto == enumValueProto {
						method = m
					}
				}

				if len(location.Path) == 4 {
					// This is a comment for the method.
					method.Comments = combineComments(location)
				}
			}
		}
	}
}

func setMessageComments(file *File, proto *descriptor.DescriptorProto, location *descriptor.SourceCodeInfo_Location, relativePath []int32) {
	var message *Message
	for _, m := range file.Messages {
		if m.Proto == proto {
			message = m
		}
	}

	if len(relativePath) == 0 {
		// This is a comment for the message.
		message.Comments = combineComments(location)
		return
	}

	switch relativePath[0] {
	case 2: // Field
		var field *Field
		fieldProto := message.Proto.Field[relativePath[1]]

		fields := message.Fields
		if fieldProto.OneofIndex != nil {
			oneof := message.Oneofs[fieldProto.GetOneofIndex()]
			fields = oneof.Fields
		}
		for _, f := range fields {
			if f.Proto == fieldProto {
				field = f
			}
		}

		if len(relativePath) == 2 {
			// This is a comment for the field.
			field.Comments = combineComments(location)
		}
	case 3: // Nested message
		messageProto := message.Proto.NestedType[relativePath[1]]
		setMessageComments(file, messageProto, location, relativePath[2:])
	case 4: // Nested enum
		enumProto := message.Proto.EnumType[relativePath[1]]
		setEnumComments(file, enumProto, location, relativePath[2:])
	case 8: // Oneof
		for _, f := range message.Fields {
			if f.IsOneof && f.OneofIndex == relativePath[1] {
				f.Comments = combineComments(location)
			}
		}
	}
}

func setEnumComments(file *File, proto *descriptor.EnumDescriptorProto, location *descriptor.SourceCodeInfo_Location, relativePath []int32) {
	var enum *Enum
	for _, e := range file.Enums {
		if e.Proto == proto {
			enum = e
		}
	}

	if len(relativePath) == 0 {
		// This is a comment for the enum.
		enum.Comments = combineComments(location)
		return
	}

	switch relativePath[0] {
	case 2: // Enum value
		var enumValue *EnumValue
		enumValueProto := enum.Proto.Value[relativePath[1]]
		for _, v := range enum.Values {
			if v.Proto == enumValueProto {
				enumValue = v
			}
		}

		if len(relativePath) == 2 {
			// This is a comment for the enum value.
			enumValue.Comments = combineComments(location)
		}
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

func combineComments(location *descriptor.SourceCodeInfo_Location) string {
	// Ignore leading detached comments because it is likely that the comment
	// does not relate directly to the definition.
	leading := formatComments(location.GetLeadingComments())
	trailing := formatComments(location.GetTrailingComments())

	var sep string
	if leading != "" && trailing != "" {
		sep = "\n"
	}

	return strings.TrimSpace(leading + sep + trailing)
}

func formatComments(comment string) string {
	lines := strings.Split(comment, "\n")
	for i, line := range lines {
		// For block comments enclosed between /* and */, the Protobuf compiler
		// will strip away all leading whitespaces for each line, and formatting
		// is not preserved. There is nothing we can do about that.
		//
		// For line comments beginning with //, the Protobuf compiler will
		// preserve all whitespaces immediately following the double slashes,
		// including the first space that is usually placed between the slashes
		// and the first word of the line (e.g this comment). To account for this,
		// we will heuristically remove the first space character if it is
		// present so that the generated descriptions are properly formatted.
		if strings.HasPrefix(line, " ") {
			line = line[1:]
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
