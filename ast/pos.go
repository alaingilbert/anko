package ast

import (
	"bytes"
	"encoding/gob"
)

// Position provides interface to store code locations.
type Position struct {
	Line   int
	Column int
}

// Pos interface provides two functions to get/set the position for expression or statement.
type Pos interface {
	Position() Position
	SetPosition(Position)
}

// PosImpl provides commonly implementations for Pos.
type PosImpl struct {
	pos Position
}

func (d *PosImpl) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.pos)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (d *PosImpl) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(&d.pos)
}

// Position return the position of the expression or statement.
func (x *PosImpl) Position() Position {
	return x.pos
}

// SetPosition is a function to specify position of the expression or statement.
func (x *PosImpl) SetPosition(pos Position) {
	x.pos = pos
}
