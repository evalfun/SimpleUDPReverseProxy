package udprelay

import (
	"SimpleUDPReverseProxy/stun"
	"fmt"
	"log"
	"net"
	"time"
)

// type CommonSession interface {
// 	Send([]byte) error
// 	Close(string)
// 	IsClosed() bool
// 	GetSessionInfo() (int64, int64, int64, int64, int64, int64)
// 	GetSourceAddr() *net.UDPAddr
// 	GetTargetAddr() *net.UDPAddr
// }

type StunSession struct {
	conn           *net.UDPConn
	stunServerAddr *net.UDPAddr
	publicAddr     string
	SendBytes      int64
	RecvBytes      int64
	CreateTime     int64
	ClosedTime     int64
	LastRecv       int64
	LastSend       int64
}

func NewStunSession(stun_server *net.UDPAddr, conn *net.UDPConn) *StunSession {
	session := &StunSession{
		conn:           conn,
		stunServerAddr: stun_server,
		CreateTime:     time.Now().Unix(),
		LastRecv:       time.Now().Unix(),
		LastSend:       time.Now().Unix(),
	}
	return session
}

func (this *StunSession) GetTargetAddr() string {
	return this.stunServerAddr.String()
}

func (this *StunSession) GetLocalAddr() string {
	return this.conn.LocalAddr().String()
}

func (this *StunSession) SendStunBindReqProc() {
	for {
		if this.IsClosed() {
			return
		}

		for i := 0; i < 3; i++ {
			err := this.SendBindReq(false, false)
			if err != nil {
				log.Printf("send stin bind request fail: %s\n", err.Error())
			}
			time.Sleep(1000 * time.Microsecond)
			//this.SendBindReq(false, false)
			//time.Sleep(1000 * time.Microsecond)
			//this.SendBindReq(false, true)
		}
		time.Sleep(1 * time.Minute)
	}
}

// you should not use this method
func (this *StunSession) Recv(data []byte) (int, error) {
	for {
		time.Sleep(114514 * time.Microsecond)
		if this.IsClosed() {
			break
		}
	}
	return 0, nil
}

//从stun server收到了数据, 由rely server的监听端口转发而来
func (this *StunSession) Send(data []byte) error {
	packet, err := stun.NewPacketFromBytes(data)
	if err != nil {
		fmt.Println("stun packet error:", err.Error(), this.stunServerAddr.String())
		return err
	}
	//log.Println("stun: public ip addr: ", packet.GetMappedAddr().String())
	this.publicAddr = packet.GetMappedAddr().String()
	this.RecvBytes = this.RecvBytes + int64(len(data))
	this.LastRecv = time.Now().Unix()
	return nil
}

func (this *StunSession) SendBindReq(changeIP bool, changePort bool) error {
	// Construct packet.
	pkt, err := stun.NewPacket()
	if err != nil {
		return err
	}
	pkt.Types = stun.TypeBindingRequest
	attribute := stun.NewSoftwareAttribute("ur20230324x")
	pkt.AddAttribute(*attribute)
	if changeIP || changePort {
		attribute = stun.NewChangeReqAttribute(changeIP, changePort)
		pkt.AddAttribute(*attribute)
	}
	// length of fingerprint attribute must be included into crc,
	// so we add it before calculating crc, then subtract it after calculating crc.
	pkt.Length += 8
	attribute = stun.NewFingerprintAttribute(pkt)
	pkt.Length -= 8
	pkt.AddAttribute(*attribute)
	// Send packet.
	data := pkt.Bytes()
	writeCount, err := this.conn.WriteToUDP(data, this.stunServerAddr)
	this.LastSend = time.Now().Unix()
	this.SendBytes = this.SendBytes + int64(writeCount)
	return err
}

func (this *StunSession) Close(r string) {
	this.ClosedTime = time.Now().Unix()
	return
}

func (this *StunSession) IsClosed() bool {
	if this.ClosedTime > 0 {
		return true
	}
	return false
}

func (this *StunSession) GetSessionInfo() (int64, int64, int64, int64, int64, int64) {
	return this.SendBytes, this.RecvBytes, this.CreateTime, this.ClosedTime, this.LastRecv, this.LastSend
}

func (this *StunSession) GetSourceAddr() *net.UDPAddr {
	return this.stunServerAddr
}

func (this *StunSession) GetPublicAddr() string {
	return this.publicAddr
}
