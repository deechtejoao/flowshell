package tests

import (
	"testing"

	
	"github.com/stretchr/testify/assert"
	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func TestSerializeNodes(t *testing.T) {
	t.Run("LoadFileAction", func(t *testing.T) {
		before := nodes.NewLoadFileNode("foo/bar")

		enc := core.NewEncoder(1)
		assert.True(t, core.SThing(enc, before))
		assert.True(t, enc.Ok())

		buf := enc.Bytes()
		t.Log("encoded:", buf)

		dec := core.NewDecoder(buf)
		var after core.Node
		assert.True(t, core.SThing(dec, &after))

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