package crypts

import (
	"errors"
	"fmt"
)

const (
	CRYPT_METHOD_AES_GCM    int = 9
	CRYPT_METHOD_NONE_SHA1  int = 7
	CRYPT_METHOD_NONE_CRC32 int = 8
)

type Cryption interface {
	SetPassword([]byte, []byte) error
	Decrypt([]byte) ([]byte, error)
	Encrypt([]byte) ([]byte, error)
}

func NewCryption(method int, password []byte, salt []byte) (Cryption, error) {
	var instance Cryption
	if method == CRYPT_METHOD_AES_GCM {
		instance = &AESCryption{}
	} else if method == CRYPT_METHOD_NONE_SHA1 {
		instance = &NONESHA1Cryption{}
	} else if method == CRYPT_METHOD_NONE_CRC32 {
		instance = &NONECRC32Cryption{}
	} else {
		return nil, errors.New(fmt.Sprintf("No such crypt method %d", method))
	}
	err := instance.SetPassword(password, salt)
	return instance, err

}

func GetCryptMethodCode(s string) int {
	switch s {
	case "AES_GCM":
		return CRYPT_METHOD_AES_GCM
	case "NONE_SHA1":
		return CRYPT_METHOD_NONE_SHA1
	case "NONE_CRC32":
		return CRYPT_METHOD_NONE_CRC32
	}
	return 0
}

func GetCryptMethodStr(s int) string {
	switch s {
	case CRYPT_METHOD_AES_GCM:
		return "AES_GCM"
	case CRYPT_METHOD_NONE_SHA1:
		return "NONE_SHA1"
	case CRYPT_METHOD_NONE_CRC32:
		return "NONE_CRC32"
	}
	return "unkonwn"
}
