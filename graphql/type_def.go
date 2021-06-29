package graphql

import (
	"github.com/martinxsliu/protoc-gen-graphql/parameters"
	"strings"
)

// TypeDef returns the schema definition language (SDL) representation
// of the GraphQL type.
func TypeDef(graphqlType Type, params *parameters.Parameters) string {
	switch graphqlType := graphqlType.(type) {
	case *Scalar:
		return typeDefScalar(graphqlType)
	case *Object:
		return typeDefObject(graphqlType, params.NullableListTypes)
	case *ExtendObject:
		return typeDefExtendObject(graphqlType, params.NullableListTypes)
	case *Input:
		return typeDefInput(graphqlType, params.NullableListTypes)
	case *Enum:
		return typeDefEnum(graphqlType)
	case *Union:
		return typeDefUnion(graphqlType)
	default:
		return ""
	}
}

func typeDefScalar(scalar *Scalar) string {
	return scalar.Name
}

func typeDefObject(object *Object, nullableListTypes bool) string {
	b := &strings.Builder{}

	if object.Description != "" {
		writeDescription(b, object.Description, 0)
	}

	b.WriteString("type ")
	b.WriteString(object.Name)

	// Omit braces if we don't have any fields, e.g. `type Empty`.
	if len(object.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range object.Fields {
			typeDefField(b, field, nullableListTypes)
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefField(b *strings.Builder, field *Field, nullableListTypes bool) {
	typeName := field.TypeName
	if field.Modifiers&TypeModifierNonNull > 0 {
		typeName = typeName + "!"
	}
	if field.Modifiers&TypeModifierList > 0 {
		// Protobuf repeated values are always non-null, but can have length 0.
		typeName = "[" + typeName + "]"
		if !nullableListTypes && field.Modifiers&TypeModifierNonNullList > 0 {
			typeName = typeName + "!"
		}
	}

	if field.Description != "" {
		writeDescription(b, field.Description, 2)
	}

	b.WriteString("  ")
	b.WriteString(field.Name)

	if len(field.Arguments) != 0 {
		b.WriteString("(")
		for i, arg := range field.Arguments {
			if i != 0 {
				b.WriteString(", ")
			}
			typeDefArgument(b, arg)
		}
		b.WriteString(")")
	}

	b.WriteString(": ")
	b.WriteString(typeName)

	for _, directive := range field.Directives {
		b.WriteString(" @")
		b.WriteString(directive)
	}
}

func typeDefArgument(b *strings.Builder, argument *Argument) string {
	typeName := argument.TypeName
	if argument.Modifiers&TypeModifierNonNull > 0 {
		typeName = typeName + "!"
	}

	b.WriteString(argument.Name)
	b.WriteString(": ")
	b.WriteString(typeName)

	if argument.Default != "" {
		b.WriteString(" = ")
		b.WriteString(argument.Default)
	}

	return b.String()
}

func typeDefExtendObject(object *ExtendObject, nullableListTypes bool) string {
	b := &strings.Builder{}
	b.WriteString("extend type ")
	b.WriteString(object.Name)

	// Omit braces if we don't have any fields, e.g. `type Empty`.
	if len(object.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range object.Fields {
			typeDefField(b, field, nullableListTypes)
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefInput(input *Input, nullableListTypes bool) string {
	b := &strings.Builder{}

	if input.Description != "" {
		writeDescription(b, input.Description, 0)
	}

	b.WriteString("input ")
	b.WriteString(input.Name)

	// Omit braces if we don't have any fields, e.g. `input Empty`.
	if len(input.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range input.Fields {
			typeDefField(b, field, nullableListTypes)
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefEnum(enum *Enum) string {
	b := &strings.Builder{}

	if enum.Description != "" {
		writeDescription(b, enum.Description, 0)
	}

	b.WriteString("enum ")
	b.WriteString(enum.Name)
	b.WriteString(" {\n")
	for _, value := range enum.Values {
		typeDefEnumValue(b, value)
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func typeDefEnumValue(b *strings.Builder, value *EnumValue) {
	if value.Description != "" {
		writeDescription(b, value.Description, 2)
	}

	b.WriteString("  ")
	b.WriteString(value.Name)
	for _, directive := range value.Directives {
		b.WriteString(" @")
		b.WriteString(directive)
	}
}

func typeDefUnion(union *Union) string {
	b := &strings.Builder{}

	if union.Description != "" {
		writeDescription(b, union.Description, 0)
	}

	b.WriteString("union ")
	b.WriteString(union.Name)
	b.WriteString(" = ")
	for i, name := range union.TypeNames {
		if i != 0 {
			b.WriteString(" | ")
		}
		b.WriteString(name)
	}
	return b.String()
}

func writeDescription(b *strings.Builder, description string, indent int) {
	lines := strings.Split(description, "\n")
	prefix := strings.Repeat(" ", indent)
	b.WriteString(prefix)
	b.WriteString("\"\"\"\n")
	for _, line := range lines {
		// Don't output the whitespaces if the line is only comprised of whitespaces.
		if strings.TrimSpace(line) != "" {
			b.WriteString(prefix)
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	b.WriteString(prefix)
	b.WriteString("\"\"\"\n")
}
