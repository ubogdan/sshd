package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
)

//RandString .
func RandString(n int) string {
	b := make([]byte, n) //equals 8 characters
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func Sha1String(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(bs)
}
