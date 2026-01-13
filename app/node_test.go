package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeNodes(t *testing.T) {
	t.Run("LoadFileAction", func(t *testing.T) {
		before := NewLoadFileNode("foo/bar")

		enc := NewEncoder(1)
		assert.True(t, SThing(enc, before))
		assert.True(t, enc.Ok())

		buf := enc.Bytes()
		t.Log("encoded:", buf)

		dec := NewDecoder(buf)
		var after Node
		assert.True(t, SThing(dec, &after))

		assert.True(t, dec.Ok())
		assert.Equal(t, before.ID, after.ID)
		assert.Equal(t, before.Pos, after.Pos)
		assert.Equal(t, before.Name, after.Name)
		assert.Equal(t, before.Pinned, after.Pinned)
		assert.Equal(t, before.InputPorts, after.InputPorts)
		assert.Equal(t, before.OutputPorts, after.OutputPorts)
		assert.Equal(t, before.Action, after.Action)
	})
}
