package udprelay

import (
	"hash/crc32"
	"net"
)

func ResolveUDPAddr(ip_version string, server string, max_count int) (*net.UDPAddr, error) {
	var target *net.UDPAddr
	var err error
	for i := 0; i < max_count; i++ {
		target, err = net.ResolveUDPAddr("udp4", server)
		if err == nil {
			return target, nil
		}
		break
	}
	return nil, err
}

func crc32sum(data []byte) uint32 {
	crc32q := crc32.MakeTable(0xD5828281)
	return crc32.Checksum(data, crc32q)
}
