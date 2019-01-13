package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ajwerner/monkey/ast"
)

type BuiltinFunction func(args ...Object) Object

//go:generate stringer -type ObjectType

type ObjectType int

const (
	_ ObjectType = iota
	INTEGER
	BOOLEAN
	NULL
	ERROR
	FUNCTION
	STRING
	BUILTIN
	RETURN_VALUE
)

func NewEnclosedEnvironment(parent *Environment) *Environment {
	env := NewEnvironment()
	env.parent = parent
	return env
}

func NewEnvironment() *Environment {
	return &Environment{
		store: map[string]Object{},
	}
}

type Environment struct {
	store  map[string]Object
	parent *Environment
}

func (e Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.parent != nil {
		obj, ok = e.parent.Get(name)
	}
	return obj, ok
}

func (e Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

type String string

func (s String) Type() ObjectType { return STRING }
func (s String) Inspect() string  { return string(s) }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type ReturnValue struct {
	Value Object
}

func (rv ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (rv ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Err error
}

func (e Error) Type() ObjectType { return ERROR }
func (e Error) Inspect() string  { return e.Err.Error() }

type Integer int64

func (i Integer) Type() ObjectType { return INTEGER }
func (i Integer) Inspect() string  { return fmt.Sprintf("%d", i) }

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Boolean bool

func (b Boolean) Type() ObjectType { return BOOLEAN }
func (b Boolean) Inspect() string  { return fmt.Sprintf("%t", b) }

type Null struct{}

func (n Null) Type() ObjectType { return NULL }
func (n Null) Inspect() string  { return "NULL" }

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN }
func (b *Builtin) Inspect() string  { return "builtin function" }
