package object

import (
	"fmt"
	"testing"
)

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello world"}
	hello2 := &String{Value: "Hello world"}

	diff1 := &String{Value: "My name is Johnny"}

	fmt.Printf("%+v\n", hello1)

	if hello1.GetHash() != hello2.GetHash() {
		t.Errorf("strings with same content have different keys.")
	}

	if hello1.GetHash() == diff1.GetHash() {
		t.Errorf("strings with different content have the same keys")
	}
}
