package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

var Secretkey = []byte("secretkeyfromme1") //ключик
var Nonce = []byte("123456123456")         //вектор инициализации

func EncryptUid(s []byte) []byte {
	aesblock, err := aes.NewCipher(Secretkey)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return []byte{}
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return []byte{}
	}
	dst := aesgcm.Seal(nil, Nonce, s, nil) // зашифровываем
	return dst
}

func DecryptUid(s []byte) []byte {
	aesblock, err := aes.NewCipher(Secretkey)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return []byte{}
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return []byte{}
	}
	src, err := aesgcm.Open(nil, Nonce, s, nil) // расшифровываем
	if err != nil {
		fmt.Printf("error: %v\n", err, src, "-----", string(s))
		return []byte{}
	}
	return src
}
