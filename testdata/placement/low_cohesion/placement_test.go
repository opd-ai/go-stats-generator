package placement

import (
	"strings"
	"testing"
)

func TestHandleUser(t *testing.T) {
	u := &User{ID: 7, Name: "Alice"}
	got := HandleUser(u)
	if !strings.Contains(got, "7") {
		t.Errorf("HandleUser(%+v) = %q, expected output containing user ID", u, got)
	}
}

func TestHandleProduct(t *testing.T) {
	p := &Product{ID: 3, Title: "Widget", Price: 9.99}
	got := HandleProduct(p)
	if !strings.Contains(got, "3") {
		t.Errorf("HandleProduct(%+v) = %q, expected output containing product ID", p, got)
	}
}

func TestHandleOrder(t *testing.T) {
	o := &Order{ID: 5, UserID: 1, ProductID: 2}
	got := HandleOrder(o)
	if !strings.Contains(got, "5") {
		t.Errorf("HandleOrder(%+v) = %q, expected output containing order ID", o, got)
	}
}

func TestFormatUser(t *testing.T) {
	u := &User{ID: 1, Name: "Bob"}
	got := FormatUser(u)
	want := HandleUser(u)
	if got != want {
		t.Errorf("FormatUser = %q, want %q", got, want)
	}
}

func TestFormatProduct(t *testing.T) {
	p := &Product{ID: 2, Title: "Gadget", Price: 19.99}
	got := FormatProduct(p)
	want := HandleProduct(p)
	if got != want {
		t.Errorf("FormatProduct = %q, want %q", got, want)
	}
}

func TestFormatOrder(t *testing.T) {
	o := &Order{ID: 10, UserID: 1, ProductID: 2}
	got := FormatOrder(o)
	want := HandleOrder(o)
	if got != want {
		t.Errorf("FormatOrder = %q, want %q", got, want)
	}
}

func TestProcessAll(t *testing.T) {
	// ProcessAll prints to stdout; just verify no panic
	ProcessAll()
}

func TestProcess1Through6(t *testing.T) {
	// Process1-6 are trivial wrappers that each delegate to ProcessAll.
	// They all produce identical output; verify none of them panic.
	Process1()
	Process2()
	Process3()
	Process4()
	Process5()
	Process6()
}
