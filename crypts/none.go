package crypts

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
)

//不加密。只是校验一下

type NONESHA1Cryption struct {
	sha1p    hash.Hash
	password []byte
}

func (this *NONESHA1Cryption) SetPassword(password []byte, salt []byte) error {
	this.sha1p = sha1.New()
	this.password = make([]byte, len(password))
	copy(this.password, password)
	this.password = append(this.password, salt...)
	return nil
}

func (this *NONESHA1Cryption) Encrypt(data []byte) ([]byte, error) {
	this.sha1p.Reset()
	this.sha1p.Write(this.password)
	this.sha1p.Write(data)
	return append(data, this.sha1p.Sum(nil)...), nil
}

func (this *NONESHA1Cryption) Decrypt(data []byte) ([]byte, error) {
	encrypted_packet_length := len(data)
	if encrypted_packet_length < 28 {
		return nil, errors.New("encrypted packet length too short")
	}
	this.sha1p.Reset()
	this.sha1p.Write(this.password)
	this.sha1p.Write(data[:encrypted_packet_length-20])
	current_sha1sum := this.sha1p.Sum(nil)
	if !bytes.Equal(current_sha1sum, data[encrypted_packet_length-20:encrypted_packet_length]) {
		return nil, errors.New("wrong sha1 sum")
	}
	return data[:encrypted_packet_length-20], nil
}

type NONECRC32Cryption struct {
	passwd []byte
}

func (this *NONECRC32Cryption) SetPassword(password []byte, salt []byte) error {
	this.passwd = password
	return nil
}

func (this *NONECRC32Cryption) Encrypt(data []byte) ([]byte, error) {
	crc32q := crc32.MakeTable(0xD5828281)
	crc32sum := make([]byte, 4)
	t := crc32.Checksum(append(data, this.passwd...), crc32q)
	binary.BigEndian.PutUint32(crc32sum, t)
	enc_pkt := append(data, crc32sum...)
	return enc_pkt, nil
}

func (this *NONECRC32Cryption) Decrypt(data []byte) ([]byte, error) {
	encrypted_packet_length := len(data)
	if encrypted_packet_length < 12 {
		return nil, errors.New("encrypted packet length too short")
	}
	crc32q := crc32.MakeTable(0xD5828281)
	crc32sum := binary.BigEndian.Uint32(data[encrypted_packet_length-4 : encrypted_packet_length])
	current_crc32sum := crc32.Checksum(append(data[:encrypted_packet_length-4], this.passwd...), crc32q)

	if current_crc32sum != crc32sum {
		t := make([]byte, 4)
		binary.BigEndian.PutUint32(t, crc32sum)
		data = append(data[:encrypted_packet_length-4], t...)
		return nil, errors.New("wrong crc32 sum")
	}
	return data[:encrypted_packet_length-4], nil
}
