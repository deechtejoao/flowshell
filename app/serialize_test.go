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
