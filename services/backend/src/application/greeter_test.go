package application

import "testing"

func TestGreeterServiceGreet(t *testing.T) {
	got := (&GreeterService{}).Greet("world")

	if want := "Hello world!"; got != want {
		t.Errorf("Greet(%q) = %q, want %q", "world", got, want)
	}
}
