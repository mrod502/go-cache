package gocache

import "testing"

type TestStruct struct {
	Foo string
	Bar int
}

func TestNewInterfaceCache(t *testing.T) {

	c := NewInterfaceCache()

	if c == nil {
		t.Fatal("returned cache is nil")
	}
}

func TestCrud(t *testing.T) {
	c := NewInterfaceCache()
	key := "some-key"
	obj := TestStruct{Foo: "hello", Bar: 1 << 8}
	if c.Exists(key) {
		t.Fatal("returned true for nonexistent key")
	}

	if v, err := c.Get(key); v != nil || err != ErrKey {
		t.Fatal("something returned that shouldn't have been returned")
	}
	if err := c.Set(key, obj); err != nil {
		t.Fatal(err)
	}

	if !c.Exists(key) {
		t.Fatal("returned true for nonexistent key")
	}

	if v, err := c.Get(key); v != obj || err != nil {
		t.Fatal("returned object not equal to original", err)
	}

	if err := c.Delete(key); err != nil {
		t.Fatalf("error deleting %s: %s", key, err)
	}

	if c.Exists(key) {
		t.Fatal("returned true for nonexistent key")
	}

}

func TestWhere(t *testing.T) {
	c := NewInterfaceCache()
	key := "some-key"
	obj := TestStruct{Foo: "hello", Bar: 1 << 8}

	if err := c.Set(key, obj); err != nil {
		t.Fatal(err)
	}

	results, err := c.Where(func(i interface{}) bool {

		return i == obj
	})
	if err != nil {
		t.Fatal(err)
	}
	val := results[0].(TestStruct)

	if val != obj {
		t.Fatalf("query result not equal to expected:\n\texpected:{%+v}, got:%+v", obj, results)
	}

}
