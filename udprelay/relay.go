package udprelay

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type UDPRelay struct {
	Conn              *net.UDPConn
	LocalPublicAddr   string
	Target            *net.UDPAddr
	TargetIPVersion   string
	Session           sync.Map
	buf_size          int
	SaveClosedSession int
	SessionTimeout    int
	StunServer        string
}

func NewUDPRelay(local_port int, buf_size int) (*UDPRelay, error) {
	udp_relay := &UDPRelay{
		buf_size: buf_size,
		//Session:  make(map[string]CommonSession),
	}
	if buf_size < 128 {
		return nil, errors.New("bufSize can not smaller than 128")
	}
	udp_relay.SessionTimeout = 300
	var err error
	udp_relay.Conn, err = net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: local_port,
	})
	if err != nil {
		return nil, err
	}
	go udp_relay.recv_udp_proc()
	go udp_relay.check_session_proc()
	return udp_relay, nil
}

func (this *UDPRelay) SetTimeout(timeout int) {
	this.SessionTimeout = timeout
}

func (this *UDPRelay) SetSessionSave(timeout int) {
	this.SaveClosedSession = timeout
}

//设置目标服务器
func (this *UDPRelay) SetTargetAddr(target_ip_version string, target_addr string) error {
	var target *net.UDPAddr
	var err error
	target, err = ResolveUDPAddr(target_ip_version, target_addr, 100)
	if err != nil {
		return err
	}
	this.TargetIPVersion = target_ip_version
	this.Target = target
	return nil
}

//设置stun服务器
func (this *UDPRelay) SetStunServer(stun_server string) error {

	target, err := ResolveUDPAddr("udp4", stun_server, 100)
	if err != nil {
		return err
	}
	stun_session := NewStunSession(target, this.Conn)
	log.Println("创建stun连接", target.String())
	this.Session.Store(target.String(), stun_session)
	//this.Session[target.String()] = stun_session
	this.StunServer = target.String()
	go stun_session.SendStunBindReqProc()
	return nil
}

//从目标服务器读取数据发送udp数据 给客户端
func (this *UDPRelay) send_udp_proc(clientAddr *net.UDPAddr, session CommonSession) {
	data := make([]byte, this.buf_size)
	for {
		count, err := session.Recv(data)
		if err != nil {
			return
		}
		_, err = this.Conn.WriteToUDP(data[:count], clientAddr)
	}
}

//从客户端接收udp数据 发送给目标服务器
func (this *UDPRelay) recv_udp_proc() {
	data := make([]byte, this.buf_size)
	for {
		read_count, remoteAddr, err := this.Conn.ReadFromUDP(data)
		if err != nil {
			return
		}
		//查找现有的session
		session := this.GetSession(remoteAddr.String())
		if session == nil {
			//创建新的session
			if remoteAddr.Port == 3479 || remoteAddr.Port == 3478 { // stun session !
				var session CommonSession
				session = NewStunSession(remoteAddr, this.Conn)
				session.Send(data[:read_count])
				log.Println("创建stun连接", remoteAddr.String(), this.LocalPublicAddr)
				this.Session.Store(remoteAddr.String(), session)
			} else { // 普通session
				var session CommonSession
				session, err = NewUDPSession(this.TargetIPVersion, this.Target, this.buf_size)
				if err != nil {
					log.Println("连接创建失败", err.Error())
					continue
				}
				log.Println("连接创建", remoteAddr.String(), session.GetLocalAddr())
				this.Session.Store(remoteAddr.String(), session)
				go this.send_udp_proc(remoteAddr, session)
			}

		} else {
			err = session.Send(data[:read_count])
			if err != nil {
				log.Printf("接收数据%s错误%s\n", remoteAddr.String(), err.Error())
			}
			if remoteAddr.Port == 3478 {
				stun_session := session.(*StunSession)
				this.LocalPublicAddr = stun_session.GetPublicAddr()
			}
		}

	}
}

func (this *UDPRelay) GetSession(client_addr string) CommonSession {
	_session, ok := this.Session.Load(client_addr)
	if !ok {
		return nil
	}
	return _session.(CommonSession)
}

func (this *UDPRelay) Close() {
	this.Session.Range(func(key, value interface{}) bool {
		session := value.(CommonSession)
		session.Close("server instance closed")
		return true
	})
	this.Conn.Close()
}

//检查资源线程
func (this *UDPRelay) check_session_proc() {
	dead_conn := 0
	all_conn := 0
	for {
		if this.Conn == nil {
			return
		}
		current_time := time.Now().Unix()
		dead_conn = 0
		all_conn = 0
		this.Session.Range(func(k, v interface{}) bool {
			all_conn = all_conn + 1
			session := v.(CommonSession)
			SendBytes, RecvBytes, CreateTime, ClosedTime, LastRecv, LastSend := session.GetSessionInfo()
			var LastActive int64
			if LastRecv > LastSend {
				LastActive = LastRecv
			} else {
				LastActive = LastSend
			}

			log.Printf("连接创建时间%d 最后活跃%d 关闭时间 %d 发送%7d  接收%7d  %s\n",
				CreateTime, LastActive, ClosedTime, SendBytes, RecvBytes, k)

			if session.IsClosed() {
				dead_conn = dead_conn + 1
			}
			if session.IsClosed() && current_time-ClosedTime > int64(this.SaveClosedSession) {
				log.Println("连接清除", k, session.GetLocalAddr())
				this.Session.Delete(k)
			} else {
				if current_time-LastRecv > int64(this.SessionTimeout) &&
					current_time-LastSend > int64(this.SessionTimeout) {
					if !session.IsClosed() {
						session.Close("连接超时")
					}
				}
			}
			return true
		})
		log.Printf("连接数%3d 活动连接%3d\n", all_conn, all_conn-dead_conn)
		time.Sleep(10 * time.Second)
	}
}
