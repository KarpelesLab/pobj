package pobj_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/KarpelesLab/pobj"
)

func TestStaticStd(t *testing.T) {
	a := func() string {
		return "hello"
	}

	st := pobj.Static(a)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err := st.Call(context.Background())
	if err != nil {
		t.Errorf("failed to perform: %s", err)
	}

	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "hello" {
		t.Fatalf("failed to execute")
	}
}

func TestStaticParam(t *testing.T) {
	a := func(in struct{ A string }) string {
		return strings.ToUpper(in.A)
	}

	st := pobj.Static(a)
	if st == nil {
		t.Fatalf("unable to gen static method for param test")
	}

	res, err := st.CallArg(context.Background(), map[string]any{"A": "hello"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "HELLO" {
		t.Errorf("failed to perform, result is %s", str)
	}

	res, err = st.CallArg(context.Background(), struct{ A string }{A: "world"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok = res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "WORLD" {
		t.Errorf("failed to perform, result is %s", str)
	}
}

type scannableObject struct {
	A string
}

func (s *scannableObject) Scan(v any) error {
	s.A = fmt.Sprintf("%#v", v)
	return nil
}

func TestStaticScanner(t *testing.T) {
	a := func(in struct{ Foo *scannableObject }) string {
		return in.Foo.A
	}

	st := pobj.Static(a)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err := st.CallArg(context.Background(), map[string]any{"Foo": "Hello"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != `"Hello"` {
		t.Errorf("failed to perform, result is %s", str)
	}

	b := func(in struct{ Foo scannableObject }) string {
		return in.Foo.A
	}

	st = pobj.Static(b)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err = st.CallArg(context.Background(), map[string]any{"Foo": "World"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok = res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != `"World"` {
		t.Errorf("failed to perform, result is %s", str)
	}
}
