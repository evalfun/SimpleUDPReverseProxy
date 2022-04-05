package crypts

import (
	"errors"
	"fmt"
)

const (
	CRYPT_METHOD_AES  int = 9
	CRYPT_METHOD_NONE int = 7
)

type Cryption interface {
	SetPassword([]byte, []byte) error
	Decrypt([]byte) ([]byte, error)
	Encrypt([]byte) ([]byte, error)
}

func NewCryption(method int, password []byte, salt []byte) (Cryption, error) {
	var instance Cryption
	if method == CRYPT_METHOD_AES {
		instance = &AESCryption{}
		err := instance.SetPassword(password, salt)
		return instance, err
	} else if method == CRYPT_METHOD_NONE {
		instance = &NONECryption{}
		return instance, nil
	}
	return nil, errors.New(fmt.Sprintf("No such crypt method %d", method))
}

func GetCryptMethodCode(s string) int {
	switch s {
	case "AES":
		return CRYPT_METHOD_AES
	case "NONE":
		return CRYPT_METHOD_NONE
	}
	return 0
}

func GetCryptMethodStr(s int) string {
	switch s {
	case CRYPT_METHOD_AES:
		return "AES"
	case CRYPT_METHOD_NONE:
		return "NONE"
	}
	return "unkonwn"
}
