package ast

// TypeKind is the kinds of types
type TypeKind int

const (
	// TypeDefault default type
	TypeDefault TypeKind = iota
	// TypePtr ptr type
	TypePtr
	// TypeSlice slice type
	TypeSlice
	// TypeMap map type
	TypeMap
	// TypeChan chan type
	TypeChan
	// TypeStructType struct type
	TypeStructType
)

// TypeStruct is the type and sub-types
type TypeStruct struct {
	Kind        TypeKind
	Env         []string
	Name        string
	Dimensions  int
	SubType     *TypeStruct
	Key         *TypeStruct
	StructNames []string
	StructTypes []*TypeStruct
	Mutable     bool
}
