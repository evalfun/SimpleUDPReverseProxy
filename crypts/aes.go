package crypts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type AESCryption struct {
	aesgcm cipher.AEAD
	key    []byte
}

func (this *AESCryption) SetPassword(password []byte, salt []byte) error {
	//创建加密实例
	this.key = pbkdf2.Key(password, salt, 1, 16, sha1.New)
	block, err := aes.NewCipher(this.key)
	if err != nil {
		return err
	}
	this.aesgcm, err = cipher.NewGCM(block)
	return err
}

//AesEncrypt 加密
func (this *AESCryption) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, this.aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	out := this.aesgcm.Seal(nonce, nonce, data, nil)
	return out, nil
}

//AesDecrypt 解密
func (this *AESCryption) Decrypt(data []byte) ([]byte, error) {
	nonce, ciphertext := data[:this.aesgcm.NonceSize()], data[this.aesgcm.NonceSize():]
	out, err := this.aesgcm.Open(nil, nonce, ciphertext, nil)
	return out, err
}
