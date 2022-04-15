package udprelay

import (
	"errors"
	"net"
	"time"
)

type CommonSession interface {
	Send([]byte) error
	Recv([]byte) (int, error)
	Close(string)
	IsClosed() bool
	//SendBytes, RecvBytes, CreateTime, ClosedTime, LastRecv, LastSend
	GetSessionInfo() (int64, int64, int64, int64, int64, int64)
	GetLocalAddr() string
	GetTargetAddr() string
}

type SessionStat struct {
	SendBytes  int64
	RecvBytes  int64
	CreateTime int64
	ClosedTime int64
	TargetAddr string
}

type UDPSession struct {
	Conn       *net.UDPConn
	TargetAddr *net.UDPAddr // 目标服务器的地址
	SendBytes  int64
	RecvBytes  int64
	CreateTime int64
	ClosedTime int64
	LastRecv   int64
	LastSend   int64
}

func (this *UDPSession) GetTargetAddr() string {
	return this.TargetAddr.String()
}
func (this *UDPSession) GetLocalAddr() string {
	return this.Conn.LocalAddr().String()
}

func (this *UDPSession) GetSessionInfo() (int64, int64, int64, int64, int64, int64) {
	return this.SendBytes, this.RecvBytes, this.CreateTime, this.ClosedTime, this.LastRecv, this.LastSend
}

func NewUDPSession(ip_version string, target_addr *net.UDPAddr, buf_size int) (*UDPSession, error) {
	conn, err := net.DialUDP(ip_version, nil, target_addr)
	if err != nil {
		return nil, err
	}
	udp_session := &UDPSession{
		Conn:       conn,
		TargetAddr: target_addr,
	}
	udp_session.LastSend = time.Now().Unix()
	udp_session.CreateTime = udp_session.LastSend
	udp_session.LastRecv = udp_session.LastSend
	return udp_session, nil
}

func (this *UDPSession) Close(r string) {
	if this.ClosedTime > 0 {
		return
	}
	this.ClosedTime = time.Now().Unix()
	//log.Println("Session closed ", this.TargetAddr.String(), r)
	this.Conn.Close()
}

func (this *UDPSession) IsClosed() bool {
	if this.ClosedTime > 0 {
		return true
	}
	return false
}

func (this *UDPSession) Send(data []byte) error {
	if this.ClosedTime != 0 {
		return errors.New("connection closed")
	}
	this.LastSend = time.Now().Unix()
	send, err := this.Conn.Write(data)
	if err != nil {
		this.Close(err.Error())
		return err
	}
	this.SendBytes = this.SendBytes + int64(send)
	return nil
}

func (this *UDPSession) Recv(data []byte) (int, error) {
	read, err := this.Conn.Read(data)
	if err != nil {
		this.Close(err.Error())
		return 0, err
	}
	this.LastRecv = time.Now().Unix()
	this.RecvBytes = int64(read) + this.RecvBytes
	return read, err
}
