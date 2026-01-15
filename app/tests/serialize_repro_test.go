package tests

import (
	"testing"

	"github.com/bvisness/flowshell/app/core"
)

type TestStruct struct {
	Value int
	Kind  int
}

func (t *TestStruct) Serialize(s *core.Serializer) bool {
	core.SInt(s, &t.Value)
	core.SInt(s, &t.Kind)
	return s.Ok()
}

func TestSMaybeThing_Encode(t *testing.T) {
	// Create an object with non-zero values
	obj := &TestStruct{Value: 42, Kind: 7}

	// Encode it
	encoder := core.NewEncoder(1)
	if !core.SMaybeThing(encoder, &obj) {
		t.Fatalf("Failed to encode")
	}

	// Check if obj was modified (clobbered)
	if obj.Value != 42 || obj.Kind != 7 {
		t.Errorf("Object was clobbered during Encode! Value=%d, Kind=%d", obj.Value, obj.Kind)
	}

	// Decode it
	data := encoder.Bytes()
	decoder := core.NewDecoder(data)
	var decoded *TestStruct
	if !core.SMaybeThing(decoder, &decoded) {
		t.Fatalf("Failed to decode")
	}

	// Verify values
	if decoded == nil {
		t.Fatalf("Decoded object is nil")
	}
	if decoded.Value != 42 {
		t.Errorf("Expected Value=42, got %d", decoded.Value)
	}
	if decoded.Kind != 7 {
		t.Errorf("Expected Kind=7, got %d", decoded.Kind)
	}
}
