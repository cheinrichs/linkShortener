package main

import (
	"fmt"
	"testing"
)

func TestDecodeID(t *testing.T) {
	var input = "NA=="
	var expected = 4
	decoded, err := DecodeID(input)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	fmt.Println(decoded)
	if decoded == expected {
		t.Errorf("Decoding , result: %d, expected: %d.", decoded, expected)
	}
}
