package gocache

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGet(t *testing.T) {
	k := "hello"
	v := "world"
	c := New(map[string]string{
		k: v,
	})
	if val, err := c.Get(k); err != nil {
		t.Fatal(err)
	} else {
		if v != val {
			t.Fatalf("expected %s, got %s", v, val)
		}
	}
}

func TestSet(t *testing.T) {
	k := "hello"
	v := "world"
	newVal := "42"
	c := New(map[string]string{
		k: v,
	})
	c.Set(k, newVal)
	if val, err := c.Get(k); err != nil {
		t.Fatal(err)
	} else {
		if val != newVal {
			t.Fatalf("expected %s, got %s", v, val)
		}
	}
}

func TestDelete(t *testing.T) {
	k := "hello"
	v := "world"
	c := New(map[string]string{
		k: v,
	})
	c.Delete(k)
	if c.Exists(k) {
		t.Fatalf("value at key %s still exists", k)
	}
}

func TestCopy(t *testing.T) {

	m := map[string]string{
		"hello": "world",
	}
	mMarshaled, _ := json.Marshal(m)
	c := New(m)

	cMarshaled, _ := json.Marshal(c.Copy())

	if !bytes.Equal(mMarshaled, cMarshaled) {
		t.Fatalf("expected %s, got %s", string(mMarshaled), string(cMarshaled))
	}

}
