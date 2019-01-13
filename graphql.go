package main

import (
	"strings"
)

type GraphqlType interface {
	ToGQL() string
}

type GraphqlTypeClass uint32

const (
	// Scalar types
	GraphqlTypeClassInt = iota + 1
	GraphqlTypeClassFloat
	GraphqlTypeClassString
	GraphqlTypeClassBoolean
	GraphqlTypeClassID

	// Non-scalar types
	GraphqlTypeClassObject
	GraphqlTypeClassInput
	GraphqlTypeClassEnum
	GraphqlTypeClassUnion
)

var scalarNames = map[GraphqlTypeClass]string{
	GraphqlTypeClassInt:     "Int",
	GraphqlTypeClassFloat:   "Float",
	GraphqlTypeClassString:  "String",
	GraphqlTypeClassBoolean: "Boolean",
	GraphqlTypeClassID:      "ID",
}

type GraphqlTypeModifier uint32

const (
	// When combining non-null and list modifiers, the non-null modifier only
	// refers to the items inside the list. The list itself should always be
	// non-null since protobuf repeated fields are not optional.
	GraphqlTypeModifierNonNull = 1 << iota
	GraphqlTypeModifierList
)

type GraphqlObject struct {
	Name   string
	Fields []*GraphqlField
}

func (g *GraphqlObject) ToGQL() string {
	var b strings.Builder
	b.WriteString("type ")
	b.WriteString(g.Name)

	// Omit braces if we don't have any fields, e.g. `type Empty`.
	if len(g.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range g.Fields {
			b.WriteString("  ")
			b.WriteString(field.ToGQL())
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

type GraphqlInput struct {
	Name   string
	Fields []*GraphqlField
}

func (g *GraphqlInput) ToGQL() string {
	var b strings.Builder
	b.WriteString("input ")
	b.WriteString(g.Name)

	// Omit braces if we don't have any fields, e.g. `input Empty`.
	if len(g.Fields) > 0 {
		b.WriteString(" {\n")
		for _, field := range g.Fields {
			b.WriteString("  ")
			b.WriteString(field.ToGQL())
			b.WriteString("\n")
		}
		b.WriteString("}")
	}

	return b.String()
}

type GraphqlField struct {
	Name      string
	Type      GraphqlTypeClass
	TypeName  string // Populated if the Type is non-scalar.
	Arguments []*GraphqlArgument
	Modifiers GraphqlTypeModifier
}

func (g *GraphqlField) ToGQL() string {
	typeName, ok := scalarNames[g.Type]
	if !ok {
		typeName = g.TypeName
	}

	if g.Modifiers&GraphqlTypeModifierNonNull > 0 {
		typeName = typeName + "!"
	}
	if g.Modifiers&GraphqlTypeModifierList > 0 {
		// Protobuf repeated values are always non-null.
		typeName = "[" + typeName + "]!"
	}

	var b strings.Builder
	b.WriteString(g.Name)

	if len(g.Arguments) != 0 {
		b.WriteString("(")
		for i, arg := range g.Arguments {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(arg.ToGQL())
		}
		b.WriteString(")")
	}

	b.WriteString(": ")
	b.WriteString(typeName)

	return b.String()
}

type GraphqlArgument struct {
	Name      string
	Type      GraphqlTypeClass
	TypeName  string // Populated if the Type is non-scalar.
	Default   string
	Modifiers GraphqlTypeModifier
}

func (g *GraphqlArgument) ToGQL() string {
	typeName, ok := scalarNames[g.Type]
	if !ok {
		typeName = g.TypeName
	}

	var b strings.Builder
	b.WriteString(g.Name)
	b.WriteString(": ")
	b.WriteString(typeName)

	if g.Modifiers&GraphqlTypeModifierNonNull > 0 {
		b.WriteString("!")
	}

	if g.Default != "" {
		b.WriteString(" = ")
		b.WriteString(g.Default)
	}

	return b.String()
}

type GraphqlEnum struct {
	Name   string
	Values []string
}

func (g *GraphqlEnum) ToGQL() string {
	var b strings.Builder
	b.WriteString("enum ")
	b.WriteString(g.Name)
	b.WriteString(" {\n")
	for _, value := range g.Values {
		b.WriteString("  ")
		b.WriteString(value)
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

type GraphqlUnion struct {
	Name      string
	TypeNames []string
}

func (g *GraphqlUnion) ToGQL() string {
	var b strings.Builder
	b.WriteString("union ")
	b.WriteString(g.Name)
	b.WriteString(" = ")
	for i, name := range g.TypeNames {
		if i != 0 {
			b.WriteString(" | ")
		}
		b.WriteString(name)
	}
	return b.String()
}
