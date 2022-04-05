package udprelay

import (
	"errors"
	"log"
	"net"
	"time"
)

type IncomeUDPSession struct {
	*UDPSession
}

func (this *IncomeUDPSession) InitUDPSession(conn *net.UDPConn, target_addr *net.UDPAddr) {
	this.UDPSession = &UDPSession{}
	this.Conn = conn
	this.TargetAddr = target_addr
	this.CreateTime = time.Now().Unix()
	this.LastRecv = time.Now().Unix()
	this.LastSend = time.Now().Unix()
}

func (this *IncomeUDPSession) Send(data []byte) error {
	if this.ClosedTime != 0 {
		return errors.New("connection closed")
	}
	this.SendBytes = this.SendBytes + int64(len(data))
	this.LastSend = time.Now().Unix()
	_, err := this.Conn.WriteToUDP(data, this.TargetAddr)
	if err != nil {
		this.Close(err.Error())
		return err
	}
	return nil
}

func (this *IncomeUDPSession) Close(r string) {
	if this.ClosedTime > 0 {
		return
	}
	this.ClosedTime = time.Now().Unix()
	log.Println("UDPIncome Session closed ", this.TargetAddr.String(), r)
	//this.Conn.Close()
}

// you should not use this method
func (this *IncomeUDPSession) Recv(data []byte) (int, error) {
	for {
		time.Sleep(114514 * time.Microsecond)
		if this.IsClosed() {
			break
		}
	}
	return 0, nil
}
