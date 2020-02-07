package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeID(t *testing.T) {
	var input = "SQ=="
	var expected = 73
	result, err := DecodeID(input)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	assert.Equal(t, expected, result, "Expecting `73`")
}

func TestEncodeID(t *testing.T) {
	var input = 73
	var expected = "SQ=="
	result := EncodeID(input)

	assert.Equal(t, expected, result, "Expecting `SQ==`")
}
