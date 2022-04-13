package udprelay

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net"
	"sync"
	"time"
)

type AdvancedRelayServerStat struct {
	LocalAddr         string
	PublicAddr        string
	SessionTimeout    int
	SaveClosedSession int
	StunServer        string
	Password          string
	Target            string
	TargetIPVersion   string
	LocalName         string
	EncryptMethod     int
	EncryptHeaderOnly bool
	HashHeaderOnly    bool
	OtherData         string
	Closed            bool
	BufSize           int
	ConnectionStat    []*ConnectionStat
}

type AdvancedRelayServer struct {
	conn              *net.UDPConn
	localPublicAddr   string
	session           map[string]*StunSession       //只存储stun连接
	clientConn        map[string]*AdvancedRelayConn // 存储客户端连接 k:客户端ip和端口 v:一个mux连接
	clientConnLock    sync.RWMutex
	bufSize           int
	saveClosedSession int
	sessionTimeout    int
	stunServer        string
	password          []byte
	encryptMethod     int
	encryptHeaderOnly bool
	hashHeaderOnly    bool
	target            string //如果不指定目标，则可以由客户端进行协商
	targetIPVersion   string
	localName         []byte
	otherData         []byte
	closed            bool
	trackerConfig     *TrackerConfig
	tracker           *ServerTracker
	messageChan       chan *TrackerMessage
	randomOffset      int64 //一个普通的随机值
}

//流程：初始化服务端，设置stun服务器，上报ip地址，等待连接
func NewAdvancedRelayServer(localPort int, bufSize int, sessionTimeout int, saveClosedSession int, password []byte, encryptMethod int, encryptHeaderOnly bool, HashHeaderOnly bool, localName []byte, localOtherData []byte) (*AdvancedRelayServer, error) {
	if len(localOtherData) > 1024 {
		return nil, errors.New("other data should not longer than 1024 byte")
	}
	if len(localName) > 255 {
		return nil, errors.New("local name should not longer than 255 byte")
	}
	relay := &AdvancedRelayServer{
		bufSize:           bufSize,
		password:          password,
		sessionTimeout:    sessionTimeout,
		saveClosedSession: saveClosedSession,
		localName:         localName,
		otherData:         localOtherData,
		encryptMethod:     encryptMethod,
		encryptHeaderOnly: encryptHeaderOnly,
		hashHeaderOnly:    HashHeaderOnly,
		messageChan:       make(chan *TrackerMessage),
		session:           make(map[string]*StunSession),
		clientConn:        make(map[string]*AdvancedRelayConn),
	}
	var err error
	relay.conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: localPort,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Server %s created", relay.localName)
	_t, _ := rand.Int(rand.Reader, big.NewInt(19260817))
	relay.randomOffset = _t.Int64()
	go relay.recv_udp_proc()
	go relay.check_session_proc()
	go relay.trackerProc()
	return relay, nil
}

func (this *AdvancedRelayServer) GetSessionDict() map[string][]*SessionStat {
	var sessionDict map[string][]*SessionStat
	sessionDict = make(map[string][]*SessionStat)
	this.clientConnLock.RLock()
	defer this.clientConnLock.RUnlock()
	for k := range this.clientConn {
		sessionDict[this.clientConn[k].remoteAddrString] = this.clientConn[k].GetSessionList()
	}
	return sessionDict
}

//从track接收信息，准备发起主动连接
func (this *AdvancedRelayServer) trackerProc() {
	for this.closed != true {
		msg, ok := <-this.messageChan
		if !ok {
			return
		}
		if msg.MessageType == TK_MSG_TYPE_CONNECT {
			for i := range msg.AddrList {
				addr, err := ResolveUDPAddr("udp", msg.AddrList[i], 5)
				if err == nil {
					this.ReqClientConn(addr)
				}
			}
		} else if msg.MessageType == TK_MSG_TYPE_CONNECT_SUCCESS {
			log.Printf("Server %s connected to tracker success", this.localName)
			this.SendAddrtoTracker()
			//tracker成功连接的时候汇报一波
		}
	}
}
func (this *AdvancedRelayServer) GetTrackerConfig() *TrackerConfig {
	return this.trackerConfig
}
func (this *AdvancedRelayServer) GetTrackerStat() (string, int) {
	return this.tracker.GetStat()
}
func (this *AdvancedRelayServer) SetTrackerConfig(config *TrackerConfig) error {
	if this.tracker != nil {
		this.tracker.Close()
	}
	if config != nil {
		if config.ServerID == "" && config.ServerURL == "" && config.UserID == "" {
			config = nil
		} else {
			if !IsValidUUID(config.ServerID) || !IsValidUUID(config.UserID) {
				return errors.New("invalid ServerID or UserID")
			}
			this.tracker = NewServerTracker(config, this.messageChan)
		}
	}
	this.trackerConfig = config
	return nil
}

func (this *AdvancedRelayServer) GetConnectionList() []*ConnectionStat {
	var connstat []*ConnectionStat
	this.clientConnLock.RLock()
	defer this.clientConnLock.RUnlock()
	for k := range this.clientConn {
		connstat = append(connstat, this.clientConn[k].GetConnStat())
	}
	return connstat
}
func (this *AdvancedRelayServer) GetServerStat() *AdvancedRelayServerStat {
	stat := &AdvancedRelayServerStat{
		LocalAddr:         this.conn.LocalAddr().String(),
		PublicAddr:        this.localPublicAddr,
		SessionTimeout:    this.sessionTimeout,
		StunServer:        this.stunServer,
		Password:          string(this.password),
		TargetIPVersion:   this.targetIPVersion,
		LocalName:         string(this.localName),
		OtherData:         string(this.otherData),
		Closed:            this.closed,
		BufSize:           this.bufSize,
		SaveClosedSession: this.saveClosedSession,
		EncryptMethod:     this.encryptMethod,
		EncryptHeaderOnly: this.encryptHeaderOnly,
		HashHeaderOnly:    this.hashHeaderOnly,
	}
	stat.Target = this.target
	// for k := range this.clientConn {
	// 	stat.ConnectionStat = append(stat.ConnectionStat, this.clientConn[k].GetConnStat())
	// }
	return stat
}

func (this *AdvancedRelayServer) SendAddrtoTracker() {
	if this.tracker == nil {
		return
	}
	addrList := []string{this.localPublicAddr}
	this.tracker.SendAddrList(addrList)
}

//从客户端接收udp数据 发送给目标服务器
func (this *AdvancedRelayServer) recv_udp_proc() {
	for {
		data := make([]byte, this.bufSize)
		read_count, remoteAddr, err := this.conn.ReadFromUDP(data)
		remoteAddrString := remoteAddr.String()
		if err != nil {
			log.Printf("%s recv udp data fail: %s", this.localName, err.Error())
			return
		}
		//log.Printf("%s recv udp data from %s", this.localName, remoteAddr.String())
		if remoteAddr.Port == 3478 { // stun session !
			var session *StunSession
			session, ok := this.session[remoteAddrString]
			if !ok { // 创建新的stun session
				session = NewStunSession(remoteAddr, this.conn)
				this.session[remoteAddrString] = session
				log.Printf("Server %s created a new stun session", this.localName)
			} else {
				session.Send(data[:read_count])
				newAddr := session.GetPublicAddr()
				if this.localPublicAddr != newAddr {
					this.localPublicAddr = session.GetPublicAddr()
					this.SendAddrtoTracker()
					//收到新ip的时候就向tracker汇报
				}
			}
		} else if remoteAddr.Port == 3479 {
			continue
		} else { // 对端连接
			var clientConn *AdvancedRelayConn
			this.clientConnLock.Lock()
			clientConn = this.GetConnByRemoteAddr(remoteAddrString)
			if clientConn != nil {
				if clientConn.IsClosed() { // 收到一个已经关闭的连接，释放连接并尝试重新握手
					delete(this.clientConn, remoteAddrString)
					clientConn = nil
				}
			}
			if clientConn == nil { //新建连接
				// 先尝试解包，如果能解包就新建连接
				if this.password != nil {
					var err error
					_, err = DecryptPacket(data[:read_count], this.password, this.encryptMethod, this.hashHeaderOnly)
					//log.Printf("DecryptPacket %x, %d, %v", this.password, this.encryptMethod, this.hashHeaderOnly)

					if err != nil {
						log.Printf("Server %s received an invaild packet from %s %s", this.localName,
							remoteAddrString, err.Error())
						this.clientConnLock.Unlock()
						continue
					}
				}
				clientConn = NewAdvancedRelayConn(this.conn, remoteAddr, this.password, this.encryptMethod, this.encryptHeaderOnly, this.hashHeaderOnly, this.localName, this.otherData, this.sessionTimeout, this.bufSize)
				clientConn.SetTargetAddr(this.target, this.targetIPVersion)
				this.clientConn[remoteAddr.String()] = clientConn
				log.Printf("Server %s created a new connection", this.localName)
			}
			if clientConn.IsClosed() == false {
				clientConn.recvChan <- data[:read_count]
			}
			this.clientConnLock.Unlock()
		}
	}
}

func (this *AdvancedRelayServer) GetConnByRemoteAddr(addr string) *AdvancedRelayConn {
	clientConn, ok := this.clientConn[addr]
	if !ok {
		return nil
	}
	return clientConn
}

//发送主动连接，请求连接到客户端
func (this *AdvancedRelayServer) ReqClientConn(remoteAddr *net.UDPAddr) {
	remoteAddrString := remoteAddr.String()
	log.Printf("Server %s try connect to client %s", this.localName, remoteAddrString)
	this.clientConnLock.Lock()
	defer this.clientConnLock.Unlock()
	clientConn := this.GetConnByRemoteAddr(remoteAddrString)
	if clientConn != nil {
		if clientConn.IsClosed() == false {
			return
		} else { // 如果连接已经存在但关闭了，那就删除当前连接，重新建立一个
			delete(this.clientConn, remoteAddrString)
		}
	}

	clientConn = NewAdvancedRelayConn(this.conn, remoteAddr, this.password, this.encryptMethod, this.encryptHeaderOnly, this.hashHeaderOnly, this.localName, this.otherData, this.sessionTimeout, this.bufSize)
	clientConn.SetTargetAddr(this.target, this.targetIPVersion)
	this.clientConn[remoteAddrString] = clientConn
	clientConn.RequestConnect()
}

func (this *AdvancedRelayServer) SetTimeout(timeout int) {
	log.Printf("Server %s session timeout set to %d", this.localName, timeout)
	this.sessionTimeout = timeout
}

func (this *AdvancedRelayServer) SetSessionSave(timeout int) {
	log.Printf("Server %s save closed session duration set to %d", this.localName, timeout)
	this.saveClosedSession = timeout
}

//设置用户目标服务器地址
func (this *AdvancedRelayServer) SetTargetAddr(targetIPVersion string, target string) {
	log.Printf("Server %s target user server set to %s %s", this.localName, targetIPVersion, target)
	this.targetIPVersion = targetIPVersion
	this.target = target
}

//设置stun服务器
func (this *AdvancedRelayServer) SetStunServer(stun_server string) error {

	target, err := ResolveUDPAddr("udp", stun_server, 100)
	if err != nil {
		return err
	}
	stun_session := NewStunSession(target, this.conn)
	this.session[target.String()] = stun_session
	this.stunServer = target.String()
	go stun_session.SendStunBindReqProc()
	log.Printf("Server %s create stun session %s", this.localName, stun_server)
	return nil
}

//检查资源线程
func (this *AdvancedRelayServer) check_session_proc() {
	var stunSession CommonSession
	var currentTime int64
	for this.closed == false {
		currentTime = time.Now().Unix()

		for k := range this.session {
			_, _, _, ClosedTime, LastRecv, LastSend := this.session[k].GetSessionInfo()
			if currentTime-LastRecv > int64(this.sessionTimeout) && currentTime-LastSend > int64(this.sessionTimeout) {
				stunSession.Close("session timeout")
				log.Printf("Client %s stun session timeout", this.localName)
			}
			if currentTime-ClosedTime > int64(this.saveClosedSession) && ClosedTime != 0 {
				delete(this.session, k)
				log.Printf("Client %s delete closed stun session", this.localName)
			}
		}
		this.clientConnLock.Lock()
		for k := range this.clientConn {
			lastSend, lastRecv, stat, _, _, _ := this.clientConn[k].GetConnInfo()
			if stat == ARC_STAT_CLOSED && currentTime-lastSend > int64(this.saveClosedSession+this.sessionTimeout) &&
				currentTime-lastRecv > int64(this.saveClosedSession+this.sessionTimeout) {
				delete(this.clientConn, k)
			}
		}
		this.clientConnLock.Unlock()
		if (currentTime+this.randomOffset)%1200 == 0 {
			//向tracker汇报自己的ip
			//为了避免所有服务端同时向tracker汇报，就加个随机值
			this.SendAddrtoTracker()
		}
		time.Sleep(1 * time.Second)
	}
}

func (this *AdvancedRelayServer) Close() {
	this.closed = true
	this.conn.Close()
	if this.tracker != nil {
		this.tracker.Close()
	}
	for k := range this.session {
		this.session[k].Close("advanced relay server closed")
	}
	this.clientConnLock.Lock()
	defer this.clientConnLock.Unlock()
	for k := range this.clientConn {
		this.clientConn[k].Close("server stop")
	}
	close(this.messageChan)
}
