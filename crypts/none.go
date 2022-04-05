package crypts

import (
	"errors"
)

type NONECryption struct {
}

func (this *NONECryption) SetPassword(password []byte, salt []byte) error {
	return nil
}

//AesEncrypt 加密
func (this *NONECryption) Encrypt(data []byte) ([]byte, error) {
	//判断加密快的大小
	blockSize := 16
	//填充
	encryptBytes := pkcs7Padding(data, blockSize)
	return encryptBytes, nil
}

//AesDecrypt 解密
func (this *NONECryption) Decrypt(data []byte) ([]byte, error) {
	//获取块的大小
	blockSize := 16
	if len(data)%blockSize != 0 {
		//log.Printf("加密包长度%d 块长度%d", len(data), this.block.BlockSize())
		return nil, errors.New("wrong block size")
	}
	//去除填充
	crypted, err := pkcs7UnPadding(data)
	return crypted, err
}
