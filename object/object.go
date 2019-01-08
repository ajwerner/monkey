package object

import "fmt"

//go:generate stringer -type ObjectType

type ObjectType int

const (
	_ ObjectType = iota
	INTEGER
	BOOLEAN
	NULL
)

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
func (n Null) Inspect() string  { return "null" }
