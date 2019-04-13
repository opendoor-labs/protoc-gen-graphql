package graphql

import "strings"

// TypeDef returns the schema definition language (SDL) representation
// of the GraphQL type.
func TypeDef(graphqlType Type) string {
	switch graphqlType := graphqlType.(type) {
	case *Scalar:
		return typeDefScalar(graphqlType)
	case *Object:
		return typeDefObject(graphqlType)
	case *ExtendObject:
		return typeDefExtendObject(graphqlType)
	case *Input:
		return typeDefInput(graphqlType)
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

func typeDefObject(object *Object) string {
	var b strings.Builder
	b.WriteString("type ")
	b.WriteString(object.Name)

	// Omit braces if we don't have any fields, e.g. `type Empty`.
	if len(object.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range object.Fields {
			b.WriteString("  ")
			b.WriteString(typeDefField(field))
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefField(field *Field) string {
	typeName := field.TypeName
	if field.Modifiers&TypeModifierNonNull > 0 {
		typeName = typeName + "!"
	}
	if field.Modifiers&TypeModifierList > 0 {
		// Protobuf repeated values are always non-null, but can have length 0.
		typeName = "[" + typeName + "]!"
	}

	var b strings.Builder
	b.WriteString(field.Name)

	if len(field.Arguments) != 0 {
		b.WriteString("(")
		for i, arg := range field.Arguments {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(typeDefArgument(arg))
		}
		b.WriteString(")")
	}

	b.WriteString(": ")
	b.WriteString(typeName)

	return b.String()
}

func typeDefArgument(argument *Argument) string {
	typeName := argument.TypeName
	if argument.Modifiers&TypeModifierNonNull > 0 {
		typeName = typeName + "!"
	}

	var b strings.Builder
	b.WriteString(argument.Name)
	b.WriteString(": ")
	b.WriteString(typeName)

	if argument.Default != "" {
		b.WriteString(" = ")
		b.WriteString(argument.Default)
	}

	return b.String()
}

func typeDefExtendObject(object *ExtendObject) string {
	var b strings.Builder
	b.WriteString("extend type ")
	b.WriteString(object.Name)

	// Omit braces if we don't have any fields, e.g. `type Empty`.
	if len(object.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range object.Fields {
			b.WriteString("  ")
			b.WriteString(typeDefField(field))
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefInput(input *Input) string {
	var b strings.Builder
	b.WriteString("input ")
	b.WriteString(input.Name)

	// Omit braces if we don't have any fields, e.g. `input Empty`.
	if len(input.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range input.Fields {
			b.WriteString("  ")
			b.WriteString(typeDefField(field))
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

func typeDefEnum(enum *Enum) string {
	var b strings.Builder
	b.WriteString("enum ")
	b.WriteString(enum.Name)
	b.WriteString(" {\n")
	for _, value := range enum.Values {
		b.WriteString("  ")
		b.WriteString(value)
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func typeDefUnion(union *Union) string {
	var b strings.Builder
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
