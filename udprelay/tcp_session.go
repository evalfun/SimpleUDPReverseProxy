package udprelay

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type TCPSession struct {
	Conn       *net.TCPConn
	TargetAddr *net.TCPAddr // 目标服务器的地址
	SendBytes  int64
	RecvBytes  int64
	CreateTime int64
	ClosedTime int64
	LastRecv   int64
	LastSend   int64
}

func (this *TCPSession) GetLocalAddr() string {
	return this.Conn.LocalAddr().String()
}

func (this *TCPSession) GetSessionInfo() (int64, int64, int64, int64, int64, int64) {
	return this.SendBytes, this.RecvBytes, this.CreateTime, this.ClosedTime, this.LastRecv, this.LastSend
}

func (this *TCPSession) GetTargetAddr() string {
	return this.TargetAddr.String()
}

func NewTCPSession(ip_version string, target_addr *net.TCPAddr, buf_size int) (*TCPSession, error) {
	conn, err := net.DialTCP(ip_version, nil, target_addr)
	if err != nil {
		return nil, err
	}
	session := &TCPSession{
		Conn:       conn,
		TargetAddr: target_addr,
	}
	session.LastSend = time.Now().Unix()
	session.CreateTime = session.LastSend
	session.LastRecv = session.LastSend
	return session, nil
}

func (this *TCPSession) Close(r string) {
	if this.ClosedTime > 0 {
		return
	}
	this.ClosedTime = time.Now().Unix()
	fmt.Println("连接关闭", this.Conn.LocalAddr().String(), r)
	this.Conn.Close()
}

func (this *TCPSession) IsClosed() bool {
	if this.ClosedTime > 0 {
		return true
	}
	return false
}

func (this *TCPSession) Send(data []byte) error {
	if this.ClosedTime != 0 {
		return errors.New("connection closed")
	}
	this.SendBytes = this.SendBytes + int64(len(data))
	this.LastSend = time.Now().Unix()
	_, err := this.Conn.Write(data)
	if err != nil {
		this.Close(err.Error())
		return err
	}

	return nil
}

func (this *TCPSession) Recv(data []byte) (int, error) {
	read, err := this.Conn.Read(data)
	if err != nil {
		this.Close(err.Error())
		return 0, err
	}
	this.LastRecv = time.Now().Unix()
	this.RecvBytes = int64(read) + this.RecvBytes
	return read, err
}
