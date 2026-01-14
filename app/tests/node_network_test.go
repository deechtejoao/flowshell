package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func TestHTTPRequestNode(t *testing.T) {
	// Start a local test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello World"))
		} else if r.URL.Path == "/post" && r.Method == "POST" {
			body, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Posted: " + string(body)))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("GET Request", func(t *testing.T) {
		node := nodes.NewHTTPRequestNode()
		action := node.Action.(*nodes.HTTPRequestAction)

		setupGraph(node, core.NewStringValue(server.URL+"/test"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, int64(200), res.Outputs[0].Int64Value)
		assert.Equal(t, "Hello World", string(res.Outputs[1].BytesValue))
	})

	t.Run("POST Request", func(t *testing.T) {
		node := nodes.NewHTTPRequestNode()
		action := node.Action.(*nodes.HTTPRequestAction)

		setupGraph(node,
			core.NewStringValue(server.URL+"/post"),
			core.NewStringValue("POST"),
			core.NewStringValue("data"),
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, int64(201), res.Outputs[0].Int64Value)
		assert.Equal(t, "Posted: data", string(res.Outputs[1].BytesValue))
	})
}
