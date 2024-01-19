package main

import (
	"testing"

	"golang.org/x/example/hello/reverse"
)

func TestReverse(t *testing.T) {
	in, want := "Hello, OTUS!", "!SUTO ,olleH"

	got := reverse.String(in)

	if got != want {
		t.Errorf("String(%q) == %q, want %q", in, got, want)
	}
}
