package udprelay

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"main/crypts"
)

var passwd_salt []byte = []byte("11223344__passwd_salt__x")

type Packet struct {
	MsgType   uint8
	SessionID uint16
	Data      []byte
	SN        uint32
}

//完整报文
// |randomIV|  加密数据	 |
// | 16bit  |  data		|
// | 2byte  |  data		|

//解密后数据
// |MSG_TYPE | COMPRESS_TYPE| saved |Session id |serial number| 压缩过的或者没有压缩的数据| CRC32 sum
// | 8bit    | 4bit         | 4bit  | 16 bit    |    32bit    |  data                   | 32bit
// | 1byte   |       1 byte         | 2byte     |    4byte    |  data                   | 4byte
//解密数据包，以后还要做解压缩
//return 解密后的数据包  数据包类型 Session 错误信息
func DecryptPacket(encrypted_packet []byte, password []byte, method int, hashHeaderOnly bool) (*Packet, error) {
	encrypted_packet_length := len(encrypted_packet)
	if encrypted_packet_length < 10 {
		return nil, errors.New("Packet length too short")
	}
	var decrypted_packet []byte
	// if password == nil { //不要解密
	// 	decrypted_packet = encrypted_packet[2:encrypted_packet_length]
	// } else {
	cryptInstance, err := crypts.NewCryption(method, append(password, encrypted_packet[0:2]...), passwd_salt)
	if err != nil {
		return nil, errors.New("Create crypt instance error: " + err.Error())
	}
	decrypted_packet, err = cryptInstance.Decrypt(encrypted_packet[2:])
	if err != nil {
		return nil, errors.New("Unable to decrypt packet:" + err.Error())
	}
	//}
	decrypted_packet_length := len(decrypted_packet)
	if decrypted_packet_length < 12 {
		return nil, errors.New("Decrypted packet too short")
	}
	//校验数据包
	crc32 := binary.BigEndian.Uint32(decrypted_packet[decrypted_packet_length-4 : decrypted_packet_length])
	var currentCRC32sum uint32
	if hashHeaderOnly {
		currentCRC32sum = crc32sum(decrypted_packet[:8])
	} else {
		currentCRC32sum = crc32sum(decrypted_packet[:decrypted_packet_length-4])
	}
	if crc32 != currentCRC32sum {
		return nil, errors.New("wrong CRC32 sum")
	}
	packet := new(Packet)
	packet.MsgType = uint8(decrypted_packet[0])
	packet.SessionID = binary.BigEndian.Uint16(decrypted_packet[2:4])
	packet.SN = uint32(binary.BigEndian.Uint32(decrypted_packet[4:8]))
	packet.Data = decrypted_packet[8 : decrypted_packet_length-4]
	return packet, nil
}

//加密数据包，以后还要做压缩
// arg 未加密数据包 密码 数据包类型 session 压缩类型
//return 数据包
func (this *Packet) EncryptPacket(password []byte, method int, compress_type uint8, hashHeaderOnly bool) ([]byte, error) {
	var iv []byte
	iv = make([]byte, 2)
	var encrypted_packet []byte

	packet_header := make([]byte, 8)
	packet_header[0] = this.MsgType                                //消息类型
	packet_header[1] = compress_type << 4                          // 压缩类型
	binary.BigEndian.PutUint16(packet_header[2:4], this.SessionID) //session
	binary.BigEndian.PutUint32(packet_header[4:8], this.SN)        //序列号
	decrypted_packet := append(packet_header, this.Data...)        //数据
	var _packetCRC32sum uint32
	if hashHeaderOnly {
		_packetCRC32sum = crc32sum(decrypted_packet[:8])
	} else {
		_packetCRC32sum = crc32sum(decrypted_packet)
	}

	packetCRC32sum := make([]byte, 4)
	binary.BigEndian.PutUint32(packetCRC32sum, _packetCRC32sum)

	decrypted_packet = append(decrypted_packet, packetCRC32sum...)
	//if password != nil {
	rand.Read(iv)
	cryptInstance, err := crypts.NewCryption(method, append(password, iv...), passwd_salt)
	if err != nil {
		return nil, errors.New("Create encrypt instance error: " + err.Error())
	}
	encrypted_packet, err = cryptInstance.Encrypt(decrypted_packet)
	if err != nil {
		return nil, err
	}
	return append(iv, encrypted_packet...), nil
	//}
	return append(iv, decrypted_packet...), nil
}

//只有报文头部加密的报文
// | iv		| 加密头 	| 数据
// | 16bit	| 			| data
// | 2byte	| 16byte	| data
//解密头结构
// |MSG_TYPE | COMPRESS_TYPE| saved |Session id |serial number| Header CRC32 sum	|
// | 8bit    | 4bit         | 4bit  | 16 bit    |  32bit      | 32bit				|
// | 1byte   |       1 byte         | 2byte     |4byte        | 4byte				|

//只加密数据包头部
func (this *Packet) EncryptPacketHeader(password []byte, method int, compress_type uint8) ([]byte, error) {
	packetHeader := make([]byte, 12)
	packetHeader[0] = this.MsgType
	packetHeader[1] = compress_type << 4
	binary.BigEndian.PutUint16(packetHeader[2:4], this.SessionID)
	binary.BigEndian.PutUint32(packetHeader[4:8], this.SN)
	crc32 := crc32sum(packetHeader[0:8])
	binary.BigEndian.PutUint32(packetHeader[8:12], crc32)
	var iv []byte
	iv = make([]byte, 2)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	cryptInstance, err := crypts.NewCryption(method, append(password, iv...), passwd_salt)
	if err != nil {
		return nil, errors.New("Create crypt instance error: " + err.Error())
	}
	encryptedHeader, err := cryptInstance.Encrypt(packetHeader)
	if err != nil {
		return nil, errors.New("Encrypt packet header error: " + err.Error())
	}
	encryptedHeader = append(iv, encryptedHeader...)
	//log.Println("encryptedHeader length: ", len(encryptedHeader))

	return append(encryptedHeader, this.Data...), nil
}

//只解密数据包头部，以后还要做解压缩
//return packet 错误信息
func DecryptPacketHeader(packet []byte, password []byte, method int) (*Packet, error) {
	if len(packet) < 18 {
		return nil, errors.New("Encrypted packet length too short")
	}

	cryptInstance, err := crypts.NewCryption(method, append(password, packet[0:2]...), passwd_salt)
	if err != nil {
		return nil, errors.New("Create crypt instance error: " + err.Error())
	}
	decryptedPacketHeader, err := cryptInstance.Decrypt(packet[2:18])
	if err != nil {
		return nil, errors.New("Decrypt packet header error: " + err.Error())
	}
	crc32 := crc32sum(decryptedPacketHeader[0:8])
	if crc32 != binary.BigEndian.Uint32(decryptedPacketHeader[8:12]) {
		return nil, errors.New("wrong packet header crc32 sum")
	}
	_packet := &Packet{
		MsgType:   decryptedPacketHeader[0],
		SessionID: binary.BigEndian.Uint16(decryptedPacketHeader[2:4]),
		SN:        binary.BigEndian.Uint32(decryptedPacketHeader[4:8]),
		Data:      packet[18:],
	}
	return _packet, nil
}

type CreateConnInfo struct {
	ReqCompressType uint8
	NetworkType     string
	TargetAddr      []byte
	TimeStamp       uint64
	PeerName        []byte
	OtherData       []byte
	NewPasswd       []byte
}

//获取连接创建报文
// arg 压缩类型 网络类型 地址
func (this *CreateConnInfo) PackCreateConnInfo() ([]byte, error) {
	packet := make([]byte, 14)
	packet[0] = this.ReqCompressType
	var networkType uint8
	switch this.NetworkType {
	case "udp":
		networkType = PROTO_UDP
	case "tcp":
		networkType = PROTO_TCP
	case "udp4":
		networkType = PROTO_UDP4
	case "tcp4":
		networkType = PROTO_TCP4
	case "udp6":
		networkType = PROTO_UDP6
	case "tcp6":
		networkType = PROTO_TCP6
	default:
		return nil, errors.New("unknown network type")
	}
	packet[1] = networkType << 4
	binary.BigEndian.PutUint64(packet[2:10], this.TimeStamp)
	if len(this.TargetAddr) > 255 {
		return nil, errors.New("地址长度过长")
	}
	if len(this.NewPasswd) != 16 {
		return nil, errors.New("new passwd must be 128bit ")
	}
	packet[10] = uint8(len(this.TargetAddr))
	if len(this.PeerName) > 255 {
		return nil, errors.New("server name longer than 255")
	}
	if len(this.OtherData) > 1024 {
		return nil, errors.New("other data longer than 1024")
	}
	packet[11] = uint8(len(this.PeerName))
	binary.BigEndian.PutUint16(packet[12:14], uint16(len(this.OtherData)))
	packet = append(packet, this.TargetAddr...)
	packet = append(packet, this.PeerName...)
	packet = append(packet, this.OtherData...)
	packet = append(packet, this.NewPasswd...)
	return packet, nil
}

//连接创建包报文  此报文在解密后报文的数据字段中
// |请求压缩类型 | 网络类型 |saved|timestamp|地址长度|  pn len | od len |  地址   | peername |other data |new passwd
// |4bit        | 4bit    | 4bit|   64bit | 16bit | 8bit     |  16bit |  xxxx	|   xxx   | xxx       | 128bit
// |1byte       |  1byte        |  8byte  | 1byte | 1byte    | 2byte  |  xxxx	|   xxx   | xxx       | 16byte
//解析创建连接报文
//return 请求压缩类型  网络类型  目标地址 错误信息
func UnpackCreateConnInfo(packet []byte) (*CreateConnInfo, error) {
	packet_length := len(packet)
	if packet_length < 16 {
		return nil, errors.New("Create Conn packet too short")
	}
	createConnInfo := &CreateConnInfo{}
	createConnInfo.ReqCompressType = uint8(packet[0] >> 4)
	_networkType := uint8(packet[1] >> 4)
	createConnInfo.TimeStamp = binary.BigEndian.Uint64(packet[2:10])
	addr_length := int(packet[10])
	var networkType string
	switch _networkType {
	case PROTO_TCP:
		networkType = "tcp"
	case PROTO_UDP:
		networkType = "udp"
	case PROTO_TCP4:
		networkType = "tcp4"
	case PROTO_UDP4:
		networkType = "udp4"
	case PROTO_TCP6:
		networkType = "tcp6"
	case PROTO_UDP6:
		networkType = "udp6"
	default:
		return nil, errors.New("Unknown network type")
	}
	createConnInfo.NetworkType = networkType
	peerNameLength := packet[11]
	otherDataLength := binary.BigEndian.Uint16(packet[12:14])
	if packet_length < 14+addr_length+int(peerNameLength)+int(otherDataLength)+16 {
		return nil, errors.New("Create Conn addr too short")
	}
	createConnInfo.TargetAddr = packet[14 : 14+addr_length]
	createConnInfo.PeerName = packet[14+addr_length : 14+addr_length+int(peerNameLength)]
	createConnInfo.OtherData = packet[14+addr_length+int(peerNameLength) : 14+addr_length+int(peerNameLength)+int(otherDataLength)]
	createConnInfo.NewPasswd = packet[14+addr_length+int(peerNameLength)+int(otherDataLength):]
	return createConnInfo, nil
}

//确认报文
// |result |  saved |pn len |od len  | peer name  | other data
// | 8bit  |  16bit | 8bit  | 16bit  | xx         | xxx
// | 1byte |  2byte | 1byte | 2byte  | xxx        | xxx

type AckInfo struct {
	Result    uint8
	PeerName  []byte
	OtherData []byte
}

func UnpackAckInfo(data []byte) (*AckInfo, error) {
	if len(data) < 6 {
		return nil, errors.New("ack into packet too short")
	}
	ackInfo := new(AckInfo)
	ackInfo.Result = data[0]
	peerNameLength := data[3]
	otherDataLength := binary.BigEndian.Uint16(data[4:6])
	if len(data) < 6+int(peerNameLength)+int(otherDataLength) {
		return nil, errors.New("ack into packet too short_")
	}
	ackInfo.PeerName = data[6 : 6+peerNameLength]
	ackInfo.OtherData = data[6+peerNameLength : 6+int(peerNameLength)+int(otherDataLength)]
	return ackInfo, nil
}

func (this *AckInfo) PackAckInfo() ([]byte, error) {
	if len(this.PeerName) > 255 {
		return nil, errors.New("server name longer than 255")
	}
	if len(this.OtherData) > 1024 {
		return nil, errors.New("other data longer than 1024")
	}
	data := make([]byte, 6)
	data[0] = this.Result
	data[3] = uint8(len(this.PeerName))
	binary.BigEndian.PutUint16(data[4:6], uint16(len(this.OtherData)))
	data = append(data, this.PeerName...)
	data = append(data, this.OtherData...)
	return data, nil
}
