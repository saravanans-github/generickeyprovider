package main

import (
	"helper/encode"
	"testing"
)

func TestGenerateWidevinePssh(t *testing.T) {

	contentIdInBin := []byte("test123")
	keyIdInBin := encode.HexStringToBin("eee405eddea64f378e51ec167eca8d33")
	data, err := generateWidevinePssh([][]byte{keyIdInBin}, contentIdInBin, "widevine_test", "SD")

	t.Log(encode.BytesToBase64(data))

	if encode.BytesToBase64(data) != "EhDu5AXt3qZPN45R7BZ+yo0zGg13aWRldmluZV90ZXN0Igd0ZXN0MTIz" || err != nil {
		t.FailNow()
	}
}

func TestGenerateMp4Pssh(t *testing.T) {
	contentIdInBin := []byte("test123")
	keyIdInBin := encode.HexStringToBin("eee405eddea64f378e51ec167eca8d33")
	widevinePssh, _ := generateWidevinePssh([][]byte{keyIdInBin}, contentIdInBin, "widevine_test", "SD")

	mp4Pssh, err := generateMp4Pssh([][]byte{keyIdInBin}, "edef8ba979d64acea3c827dcd51d21ed", widevinePssh)
	if encode.BytesToBase64(mp4Pssh) != "AAAAXnBzc2gBAAAA7e+LqXnWSs6jyCfc1R0h7QAAAAHu5AXt3qZPN45R7BZ+yo0zAAAAKhIQ7uQF7d6mTzeOUewWfsqNMxoNd2lkZXZpbmVfdGVzdCIHdGVzdDEyMw==" || err != nil {
		t.FailNow()
	}
}
