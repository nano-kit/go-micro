package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/protobuf/proto"
	go_api "github.com/micro/go-micro/v2/api/proto"
	"github.com/micro/go-micro/v2/registry"
)

func TestRequestPayloadFromRequest(t *testing.T) {

	// our test event so that we can validate serialising / deserializing of true protos works
	protoEvent := go_api.Event{
		Name: "Test",
	}

	protoBytes, err := proto.Marshal(&protoEvent)
	if err != nil {
		t.Fatal("Failed to marshal proto", err)
	}

	jsonBytes, err := json.Marshal(protoEvent)
	if err != nil {
		t.Fatal("Failed to marshal proto to JSON ", err)
	}

	jsonUrlBytes := []byte(`{"key1":"val1","key2":"val2","name":"Test"}`)

	t.Run("extracting a json from a POST request with url params", func(t *testing.T) {
		r, err := http.NewRequest("POST", "http://localhost/my/path?key1=val1&key2=val2", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to created http.Request: %v", err)
		}

		extByte, err := requestPayload(r)
		if err != nil {
			t.Fatalf("Failed to extract payload from request: %v", err)
		}
		if string(extByte) != string(jsonUrlBytes) {
			t.Fatalf("Expected %v and %v to match", string(extByte), jsonUrlBytes)
		}
	})

	t.Run("extracting a proto from a POST request", func(t *testing.T) {
		r, err := http.NewRequest("POST", "http://localhost/my/path", bytes.NewReader(protoBytes))
		if err != nil {
			t.Fatalf("Failed to created http.Request: %v", err)
		}

		extByte, err := requestPayload(r)
		if err != nil {
			t.Fatalf("Failed to extract payload from request: %v", err)
		}
		if string(extByte) != string(protoBytes) {
			t.Fatalf("Expected %v and %v to match", string(extByte), string(protoBytes))
		}
	})

	t.Run("extracting JSON from a POST request", func(t *testing.T) {
		r, err := http.NewRequest("POST", "http://localhost/my/path", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Fatalf("Failed to created http.Request: %v", err)
		}

		extByte, err := requestPayload(r)
		if err != nil {
			t.Fatalf("Failed to extract payload from request: %v", err)
		}
		if string(extByte) != string(jsonBytes) {
			t.Fatalf("Expected %v and %v to match", string(extByte), string(jsonBytes))
		}
	})

	t.Run("extracting params from a GET request", func(t *testing.T) {

		r, err := http.NewRequest("GET", "http://localhost/my/path", nil)
		if err != nil {
			t.Fatalf("Failed to created http.Request: %v", err)
		}

		q := r.URL.Query()
		q.Add("name", "Test")
		r.URL.RawQuery = q.Encode()

		extByte, err := requestPayload(r)
		if err != nil {
			t.Fatalf("Failed to extract payload from request: %v", err)
		}
		if string(extByte) != string(jsonBytes) {
			t.Fatalf("Expected %v and %v to match", string(extByte), string(jsonBytes))
		}
	})

	t.Run("GET request with no params", func(t *testing.T) {

		r, err := http.NewRequest("GET", "http://localhost/my/path", nil)
		if err != nil {
			t.Fatalf("Failed to created http.Request: %v", err)
		}

		extByte, err := requestPayload(r)
		if err != nil {
			t.Fatalf("Failed to extract payload from request: %v", err)
		}
		if string(extByte) != "" {
			t.Fatalf("Expected %v and %v to match", string(extByte), "")
		}
	})
}

func TestSelectorStrategy(t *testing.T) {
	services := []*registry.Service{
		{
			Nodes: []*registry.Node{
				{Id: "abcd"},
				{Id: "efgh"},
				{Id: "ijkl"},
			},
		},
	}

	route := make(map[string]string) // route is a map from key to node Id
	moved := 0 // moved is a counter of moved route after adding a node
	dist := make(map[string]int)
	for i:=0; i<1000; i++ {
		key := "00001"
		node, _ := strategy(key, services)(nil)()
		dist[node.Id]++
	}
	t.Log("fixed key dist", dist)

	dist = make(map[string]int)
	for i:=0; i<1000; i++ {
		key := fmt.Sprintf("%05d", i)
		node, _ := strategy(key, services)(nil)()
		route[key] = node.Id
		dist[node.Id]++
	}
	t.Log("rand key dist", dist)

	// add a node
	services = []*registry.Service{
		{
			Nodes: []*registry.Node{
				{Id: "abcd"},
				{Id: "efgh"},
				{Id: "ijkl"},
				{Id: "mnop"},
			},
		},
	}
	dist = make(map[string]int)
	moved = 0
	for i:=0; i<1000; i++ {
		key := fmt.Sprintf("%05d", i)
		node, _ := strategy(key, services)(nil)()
		if route[key] != node.Id {
			moved++
		}
		dist[node.Id]++
	}
	t.Log("add a node dist", dist, "moved", moved)

	// del a node
	services = []*registry.Service{
		{
			Nodes: []*registry.Node{
				{Id: "abcd"},
				//{Id: "efgh"},
				{Id: "ijkl"},
			},
		},
	}
	dist = make(map[string]int)
	moved = 0
	for i:=0; i<1000; i++ {
		key := fmt.Sprintf("%05d", i)
		node, _ := strategy(key, services)(nil)()
		if route[key] != node.Id {
			moved++
		}
		dist[node.Id]++
	}
	t.Log("del a node dist", dist, "moved", moved)
}
