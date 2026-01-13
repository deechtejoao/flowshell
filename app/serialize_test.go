package app

import (
	"testing"
)

func TestSerializer_Basics(t *testing.T) {
	// Encode
	s := NewEncoder(1)
	valInt := 42
	valBool := true
	str := "hello"

	SInt(s, &valInt)
	SBool(s, &valBool)
	SStr(s, &str)

	if !s.Ok() {
		t.Fatal("Encode failed")
	}

	data := s.Bytes()

	// Decode
	d := NewDecoder(data)
	if d.Version != 1 {
		t.Errorf("Expected version 1, got %d", d.Version)
	}

	var outInt int
	var outBool bool
	var outStr string

	SInt(d, &outInt)
	SBool(d, &outBool)
	SStr(d, &outStr)

	if !d.Ok() {
		t.Fatal("Decode failed")
	}

	if outInt != valInt {
		t.Errorf("Int mismatch: %d != %d", outInt, valInt)
	}
	if outBool != valBool {
		t.Errorf("Bool mismatch: %v != %v", outBool, valBool)
	}
	if outStr != str {
		t.Errorf("Str mismatch: %s != %s", outStr, str)
	}
}

func TestSerializer_EdgeCases(t *testing.T) {
	// Empty buffer decode
	d := NewDecoder([]byte{})
	var i int
	SInt(d, &i)
	if d.Ok() {
		t.Error("Should fail on empty buffer")
	}

	// Truncated buffer
	s := NewEncoder(1)
	val := 12345
	SInt(s, &val)
	data := s.Bytes()

	d = NewDecoder(data[:len(data)-1]) // Cut off last byte
	var out int
	SInt(d, &out) // Version read might succeed (first byte) or fail.
	// Actually NewDecoder reads version immediately.
	// The buffer contains Version (varint) + Int (varint).
	// If we truncate, one of them will fail.

	if d.Ok() {
		// Try reading the int
		SInt(d, &out)
		if d.Ok() {
			t.Error("Should fail on truncated buffer")
		}

	}
}

type DummyStruct struct {
	Val int
}

func (d *DummyStruct) Serialize(s *Serializer) bool {
	return SInt(s, &d.Val)
}

func TestSerializer_ComplexTypes(t *testing.T) {
	s := NewEncoder(1)

	// Float
	f := 3.14
	SFloat(s, &f)

	// Slice (using DummyStruct)
	sl := []DummyStruct{{Val: 1}, {Val: 2}, {Val: 3}}
	SSlice(s, &sl)

	// Nil Slice
	var nilSl []DummyStruct
	SSlice(s, &nilSl)

	// Pointer (Fixed)
	val123 := int32(123)
	pVal := &val123
	SMaybeFixed(s, &pVal)

	// Nil Pointer
	var nilP *int32
	SMaybeFixed(s, &nilP)

	// Struct
	dummy := DummyStruct{Val: 99}
	SThing(s, &dummy)

	if !s.Ok() {
		t.Fatalf("Encode failed: %v", s.Errs)
	}

	// Decode
	d := NewDecoder(s.Bytes())

	var outF float64
	SFloat(d, &outF)
	if outF != f {
		t.Errorf("Float mismatch: %v != %v", outF, f)
	}

	var outSl []DummyStruct
	SSlice(d, &outSl)
	if len(outSl) != 3 || outSl[0].Val != 1 {
		t.Errorf("Slice mismatch: %v", outSl)
	}

	var outNilSl []DummyStruct
	SSlice(d, &outNilSl)
	if outNilSl != nil {
		t.Error("Nil slice became non-nil")
	}

	var outP *int32
	SMaybeFixed(d, &outP)
	if outP == nil || *outP != 123 {
		t.Error("Pointer mismatch")
	}

	var outNilP *int32
	SMaybeFixed(d, &outNilP)
	if outNilP != nil {
		t.Error("Nil pointer became non-nil")
	}

	var outDummy DummyStruct
	SThing(d, &outDummy)
	if outDummy.Val != 99 {
		t.Error("Struct mismatch")
	}
}
