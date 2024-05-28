package helper

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash"
)

// MD5 is md5
func MD5(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

// HMAC is HMAC
func HMAC(h func() hash.Hash, payload []byte, secret []byte) []byte {
	mac := hmac.New(h, secret)
	mac.Write(payload)
	return mac.Sum(nil)
}

// Sha1 is sha1
func Sha1(text string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(text)))
}
