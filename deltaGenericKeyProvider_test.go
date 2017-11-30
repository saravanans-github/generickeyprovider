package main

import (
	"helper/encode"
	"testing"
)

func TestGenerateProtoBuf(t *testing.T) {

	data, err := generateProtoBuf(encode.HexStringToBin("79221c6e568d6dbec38e2eac02d299cd"),
		[]byte("key-id:eSIcblaNbb7Dji6sAtKZzQ=="),
		"widevine_test",
		"SD")

	if encode.BytesToBase64(data) != "CAESEHkiHG5WjW2+w44urALSmc0aDXdpZGV2aW5lX3Rlc3QiH2tleS1pZDplU0ljYmxhTmJiN0RqaTZzQXRLWnpRPT0qAlNEMgA=" || err != nil {
		t.FailNow()
	}
}
