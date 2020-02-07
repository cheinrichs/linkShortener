package main

import (
	"testing"
)

func TestDecodeID(t *testing.T) {
	var input = "SQ=="
	var expected = 73
	decoded, err := DecodeID(input)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	if decoded != expected {
		t.Errorf("Decoding , result: %d, expected: %d.", decoded, expected)
	}
}

func TestEncodeID(t *testing.T) {
	var input = 73
	var expected = "SQ=="
	encoded := EncodeID(input)
	if encoded != expected {
		t.Errorf("Decoding , result: %s, expected: %s.", encoded, expected)
	}
}
