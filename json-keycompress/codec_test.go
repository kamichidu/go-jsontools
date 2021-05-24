package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	var enc Encoder
	v, err := enc.Encode([]byte(`{
		"object": {
			"hoge": 0,
			"fuga": 1
		},
		"array": [{
			"hoge": 1,
			"fuga": 2
		}, {
			"hoge": 2,
			"fuga": 3
		}]
	}`))
	if !assert.NoError(t, err) {
		return
	}
	assert.JSONEq(t, string(`{
		"*": {
			"@D": {
				"@C": 0,
				"@B": 1
			},
			"@A": [{
				"@C": 1,
				"@B": 2
			}, {
				"@C": 2,
				"@B": 3
			}]
		},
		"@": {
			"A": "array",
			"B": "fuga",
			"C": "hoge",
			"D": "object"
		}
	}`), string(v))
}

func TestDecoder(t *testing.T) {
	var dec Decoder
	v, err := dec.Decode([]byte(`{
		"*": {
			"@a": {
				"@b": 0,
				"@c": 1
			},
			"@d": [{
				"@b": 1,
				"@c": 2
			}, {
				"@b": 2,
				"@c": 3
			}]
		},
		"@": {
			"a": "object",
			"b": "hoge",
			"c": "fuga",
			"d": "array"
		}
	}`))
	if !assert.NoError(t, err) {
		return
	}
	assert.JSONEq(t, string(`{
		"object": {
			"hoge": 0,
			"fuga": 1
		},
		"array": [{
			"hoge": 1,
			"fuga": 2
		}, {
			"hoge": 2,
			"fuga": 3
		}]
	}`), string(v))
}
