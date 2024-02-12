package compiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompile1(t *testing.T) {
	by, err := Compile("a = 1")
	assert.NoError(t, err)
	Decode(by)
	//by, err = Compile(`"a" in ["a"]`)
	//assert.NoError(t, err)
	//Decode(by)
	//by, err = Compile(`a = make(chan bool); b = func (c) { c <- true }; go b(a); <- a`)
	//assert.NoError(t, err)
	//Decode(by)
	//by, err = Compile(`make(foo)`)
	//assert.NoError(t, err)
	//Decode(by)

	//src := `func(){}()`
	//stmts1, _ := parser.ParseSrc(src)
	//by, _ := Compile(src)
	//stmts2 := Decode(by)
	////pprint(stmts1)
	////pprint(stmts2)
	//assert.True(t, reflect.DeepEqual(stmts1, stmts2))
}
