package crypts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

type AESCryption struct {
	block cipher.Block
	key   []byte
}

func (this *AESCryption) SetPassword(password []byte, salt []byte) error {
	//创建加密实例
	var err error
	this.key = pbkdf2.Key(password, salt, 1, 16, sha1.New)
	this.block, err = aes.NewCipher(this.key)
	return err
}

//AesEncrypt 加密
func (this *AESCryption) Encrypt(data []byte) ([]byte, error) {
	//判断加密快的大小
	blockSize := this.block.BlockSize()
	//填充
	encryptBytes := pkcs7Padding(data, blockSize)
	//初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	//使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(this.block, this.key)
	//执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	//log.Printf("加密包长度%d 块长度%d", len(crypted), this.block.BlockSize())
	return crypted, nil
}

//AesDecrypt 解密
func (this *AESCryption) Decrypt(data []byte) ([]byte, error) {
	//获取块的大小
	blockSize := this.block.BlockSize()
	if len(data)%blockSize != 0 {
		//log.Printf("加密包长度%d 块长度%d", len(data), this.block.BlockSize())
		return nil, errors.New("wrong block size")
	}
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(this.block, this.key[:blockSize])
	//初始化解密数据接收切片
	crypted := make([]byte, len(data))
	//执行解密
	blockMode.CryptBlocks(crypted, data)
	//去除填充
	var err error
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
