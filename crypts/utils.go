package crypts

import (
	"bytes"
	"errors"
)

//pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

//pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("UnPadding fail.Wrong padding length")
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	if length-unPadding <= 0 {
		return nil, errors.New("UnPadding fail.Wrong padding length")
	}
	return data[:(length - unPadding)], nil
}
