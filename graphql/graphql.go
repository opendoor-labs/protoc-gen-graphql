package graphql

type Type interface {
	Kind() Kind
	TypeName() string
}

type Kind uint32

const (
	KindScalar Kind = iota + 1
	KindObject
	KindInput
	KindEnum
	KindUnion
)

type TypeModifier uint32

const (
	// When combining non-null and list modifiers, the non-null modifier only
	// refers to the items inside the list, while the non-null list modifier
	// refers to the nullity of the list itself.
	TypeModifierNonNull = 1 << iota
	TypeModifierList
	TypeModifierNonNullList
)

type Scalar struct {
	Name        string
	Description string
}

func (g *Scalar) Kind() Kind       { return KindScalar }
func (g *Scalar) TypeName() string { return g.Name }
func (g *Scalar) String() string   { return g.Name }

type Object struct {
	Name        string
	Description string
	Fields      []*Field
}

func (g *Object) Kind() Kind       { return KindObject }
func (g *Object) TypeName() string { return g.Name }
func (g *Object) String() string   { return g.Name }

type ExtendObject struct {
	Name   string
	Fields []*Field
}

func (g *ExtendObject) Kind() Kind       { return KindObject }
func (g *ExtendObject) TypeName() string { return g.Name }
func (g *ExtendObject) String() string   { return g.Name }

type Input struct {
	Name        string
	Description string
	Fields      []*Field
}

func (g *Input) Kind() Kind       { return KindInput }
func (g *Input) TypeName() string { return g.Name }
func (g *Input) String() string   { return g.Name }

type Field struct {
	Name        string
	Description string
	TypeName    string
	Arguments   []*Argument
	Modifiers   TypeModifier
	Directives  []string
}

type Argument struct {
	Name        string
	Description string
	TypeName    string
	Default     string
	Modifiers   TypeModifier
}

type Enum struct {
	Name        string
	Description string
	Values      []*EnumValue
}

func (g *Enum) Kind() Kind       { return KindEnum }
func (g *Enum) TypeName() string { return g.Name }
func (g *Enum) String() string   { return g.Name }

type EnumValue struct {
	Name        string
	Description string
	Directives  []string
}

type Union struct {
	Name        string
	Description string
	TypeNames   []string
}

func (g *Union) Kind() Kind       { return KindUnion }
func (g *Union) TypeName() string { return g.Name }
func (g *Union) String() string   { return g.Name }
