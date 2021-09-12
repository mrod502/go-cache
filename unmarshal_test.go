package gocache

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	v := NewBoolCache()
	b := []byte(`{"a":true}`)
	err := json.Unmarshal(b, v)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v.Get("a"))
}
