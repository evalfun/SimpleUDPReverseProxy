package udprelay

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"sync"
	"time"
)

type AdvancedRelayClientStat struct {
	LocalAddr         string
	ListenerAddr      string
	PublicAddr        string
	SessionTimeout    int
	BufSize           int
	StunServer        string
	Password          string
	CryptMethod       int
	EncryptHeaderOnly bool
	HashHeaderOnly    bool
	Target            string
	TargetIPVersion   string
	LocalName         string
	OtherData         string
	Stat              int
	SaveClosedSession int
	ConnectionStat    []*ConnectionStat
	ServerAddrList    []string
}

type AdvancedRelayClient struct {
	conn              *net.UDPConn
	serverAddrList    []*net.UDPAddr //候选服务器地址
	clientListener    *net.UDPConn
	localPublicAddr   string
	session           map[string]*StunSession       // 只存储stun连接
	serverConn        map[string]*AdvancedRelayConn // 存储客户端连接 k:客户端ip和端口 v:一个mux连接
	serverConnLock    sync.RWMutex
	bufSize           int
	saveClosedSession int
	sessionTimeout    int
	stunServer        string
	password          []byte //初始密码
	encryptMethod     int
	encryptHeaderOnly bool
	hashHeaderOnly    bool
	target            string //需要初始化
	targetIPVersion   string
	localName         []byte
	otherData         []byte
	stat              int
	compressType      uint8
	trackerConfig     *TrackerConfig
	trackerMessage    string
	serverAddrListft  []string //从tracker获取到的服务器地址
	randomOffset      int64    //一个普通的随机值
}

//流程：新建客户端，设置stun服务器，(有了公网ip，请求对方的公网ip)，设置对方地址列表，开始连接
func NewAdvancedRelayClient(listenerPort int, localPort int, bufSize int, target string, targetIPVersion string, sessionTimeout int, saveClosedSession int, password []byte, encryptMethod int, encryptHeaderOnly bool, hashHeaderOnly bool, localName []byte, localOtherData []byte, compressType uint8) (*AdvancedRelayClient, error) {
	if len(localOtherData) > 1024 {
		return nil, errors.New("other data should not longer than 1024 byte")
	}
	if len(localName) > 255 {
		return nil, errors.New("local name should not longer than 255 byte")
	}
	relay := &AdvancedRelayClient{
		bufSize:           bufSize,
		target:            target,
		targetIPVersion:   targetIPVersion,
		password:          password,
		sessionTimeout:    sessionTimeout,
		saveClosedSession: saveClosedSession,
		localName:         localName,
		otherData:         localOtherData,
		stat:              ARCL_STAT_INIT,
		compressType:      compressType,
		encryptMethod:     encryptMethod,
		encryptHeaderOnly: encryptHeaderOnly,
		hashHeaderOnly:    hashHeaderOnly,
		serverConn:        make(map[string]*AdvancedRelayConn),
		session:           make(map[string]*StunSession),
	}
	var err error
	if localPort != 0 {
		relay.conn, err = net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: localPort,
		})
	} else {
		relay.conn, err = net.ListenUDP("udp", nil)
	}
	if err != nil {
		return nil, errors.New("listen local port failed:" + err.Error())
	}
	relay.clientListener, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: listenerPort,
	})
	if err != nil {
		return nil, errors.New("listen listener port failed:" + err.Error())
	}
	_t, _ := rand.Int(rand.Reader, big.NewInt(19260817))
	relay.randomOffset = _t.Int64()
	go relay.recv_udp_proc()
	go relay.check_session_proc()
	log.Printf("Client %s created.", relay.localName)

	return relay, nil
}

func (this *AdvancedRelayClient) RestartConnection() {
	this.serverConnLock.Lock()
	for k := range this.serverConn {
		this.serverConn[k].Close("advanced relay client closed")
		log.Printf("Client %s closed.", this.localName)
	}
	this.serverConn = make(map[string]*AdvancedRelayConn)
	this.stat = ARCL_STAT_CONNECTINT
	this.serverConnLock.Unlock()
	this.StartConnect()
	this.GetAddrListFromTrackerAndConnect()
	this.RequestServerConnectClient()
}

func (this *AdvancedRelayClient) SetServerAddrList(addrList []*net.UDPAddr) {
	this.serverAddrList = addrList
	var serverList string
	for i := range this.serverAddrList {
		serverList = serverList + " " + this.serverAddrList[i].String()
	}
	//log.Printf("Client %s server list set to: %s", this.localName, serverList)
}

func (this *AdvancedRelayClient) GetTrackerConfig() *TrackerConfig {
	return this.trackerConfig
}
func (this *AdvancedRelayClient) SetTrackerConfig(config *TrackerConfig) {
	if config != nil {
		if config.ServerID == "" && config.ServerURL == "" && config.UserID == "" {
			this.trackerConfig = nil
			return
		}
	}
	this.trackerConfig = config
	this.GetAddrListFromTrackerAndConnect()
}
func (this *AdvancedRelayClient) GetTrackerStat() (string, []string) {
	return this.trackerMessage, this.serverAddrListft
}

//从tracker 获取地址并连接
func (this *AdvancedRelayClient) GetAddrListFromTrackerAndConnect() {
	if this.trackerConfig == nil {
		this.trackerMessage = "no tracker config"
		return
	}
	var err error
	this.serverAddrListft, err = this.trackerConfig.GetServerAddrList()
	if err != nil {
		this.trackerMessage = fmt.Sprintf("Get addrList from tracker failed: %s %s", err.Error(), time.Now().Format("2006.01.02 15:04:05"))
		return
	}
	this.trackerMessage = fmt.Sprintf("Get addrList from tracker success at %s", time.Now().Format("2006.01.02 15:04:05"))
	var addrList []*net.UDPAddr
	for i := range this.serverAddrListft {
		addr, err := ResolveUDPAddr("udp", this.serverAddrListft[i], 4)
		if err != nil {
			log.Printf("Client %s can not resolve addr %s from tracker: %s", this.localName, this.serverAddrListft[i], err.Error())
			continue
		}
		addrList = append(addrList, addr)
	}
	if this.stat == ARCL_STAT_CONNECTED || this.stat == ARCL_STAT_CLOSED {
		return
	}
	this.serverConnLock.Lock()
	defer this.serverConnLock.Unlock()
	for i := range addrList {
		this.connectToServer(addrList[i])
	}
	this.stat = ARCL_STAT_CONNECTINT
	return
}

//开始连接
func (this *AdvancedRelayClient) StartConnect() {
	this.serverConnLock.Lock()
	defer this.serverConnLock.Unlock()
	if this.stat == ARCL_STAT_CONNECTED || this.stat == ARCL_STAT_CLOSED {
		return
	}
	for i := 0; i < len(this.serverAddrList); i++ {
		this.connectToServer(this.serverAddrList[i])
	}
	this.stat = ARCL_STAT_CONNECTINT
}

func (this *AdvancedRelayClient) GetClientStat() *AdvancedRelayClientStat {
	stat := &AdvancedRelayClientStat{
		LocalAddr:         this.conn.LocalAddr().String(),
		ListenerAddr:      this.clientListener.LocalAddr().String(),
		PublicAddr:        this.localPublicAddr,
		SessionTimeout:    this.sessionTimeout,
		StunServer:        this.stunServer,
		Password:          string(this.password),
		CryptMethod:       this.encryptMethod,
		EncryptHeaderOnly: this.encryptHeaderOnly,
		HashHeaderOnly:    this.hashHeaderOnly,
		Target:            this.target,
		TargetIPVersion:   this.targetIPVersion,
		LocalName:         string(this.localName),
		OtherData:         string(this.otherData),
		Stat:              this.stat,
		BufSize:           this.bufSize,
		SaveClosedSession: this.saveClosedSession,
	}
	this.serverConnLock.RLock()
	defer this.serverConnLock.RUnlock()
	for k := range this.serverConn {
		stat.ConnectionStat = append(stat.ConnectionStat, this.serverConn[k].GetConnStat())
	}

	for i := range this.serverAddrList {
		stat.ServerAddrList = append(stat.ServerAddrList, this.serverAddrList[i].String())
	}
	return stat
}

func (this *AdvancedRelayClient) GetSessionDict() map[string][]*SessionStat {
	var sessionDict map[string][]*SessionStat
	sessionDict = make(map[string][]*SessionStat)
	this.serverConnLock.RLock()
	for k := range this.serverConn {
		sessionDict[this.serverConn[k].remoteAddrString] = this.serverConn[k].GetSessionList()
	}
	defer this.serverConnLock.RUnlock()
	return sessionDict
}
func (this *AdvancedRelayClient) GetConnectionList() []*ConnectionStat {
	var connstat []*ConnectionStat
	this.serverConnLock.RLock()
	for k := range this.serverConn {
		connstat = append(connstat, this.serverConn[k].GetConnStat())
	}
	defer this.serverConnLock.RUnlock()
	return connstat
}

func (this *AdvancedRelayClient) RequestServerConnectClient() {
	if this.localPublicAddr != "" && this.trackerConfig != nil {
		addrList := []string{this.localPublicAddr}
		err := this.trackerConfig.ReqServerConnect(addrList)
		if err != nil {
			log.Printf("Client %s Request server connect client failed: %s", this.localName,
				err.Error())
			this.trackerMessage = fmt.Sprintf("Request server connect failed: %s %s", err.Error(), time.Now().Format("2006.01.02 15:04:05"))
			return
		}
		this.trackerMessage = fmt.Sprintf("Request server connect success %s", time.Now().Format("2006.01.02 15:04:05"))
		log.Printf("Client %s Request server connect client success", this.localName)
	}
}

//从服务器接收udp数据 发送给目标连接
//每个client实例下，此协程只需开启一个
func (this *AdvancedRelayClient) recv_udp_proc() {
	for {
		data := make([]byte, this.bufSize)
		read_count, remoteAddr, err := this.conn.ReadFromUDP(data)
		remoteAddrString := remoteAddr.String()
		if err != nil {
			return
		}
		if remoteAddr.Port == 3478 { // stun session !
			var session *StunSession
			session, ok := this.session[remoteAddrString]
			if !ok { // 创建新的stun session
				session = NewStunSession(remoteAddr, this.conn)
				go session.SendStunBindReqProc()
				this.session[remoteAddrString] = session
			} else {
				session.Send(data[:read_count])
				//log.Printf("%s public ip from stun: %s", this.localName, session.GetPublicAddr())
				if this.localPublicAddr == "" && this.stat == ARCL_STAT_CONNECTINT {
					this.localPublicAddr = session.GetPublicAddr()
					this.RequestServerConnectClient()
				} else {
					this.localPublicAddr = session.GetPublicAddr()
				}

			}
		} else if remoteAddr.Port == 3479 {
			continue
		} else { // 对端连接
			var serverConn *AdvancedRelayConn
			this.serverConnLock.Lock()
			serverConn = this.GetConnByRemoteAddr(remoteAddrString)
			if serverConn == nil { //新建连接
				if this.stat == ARCL_STAT_CONNECTED {
					this.serverConnLock.Unlock()
					continue //已经建立连接，拒绝新建连接
				}
				// 连接之前先进行解包，和校验时间戳
				packet, err := DecryptPacket(data[:read_count], this.password, this.encryptMethod, this.hashHeaderOnly)
				if err != nil {
					log.Printf("Client %s received an invaild packet from %s %s", this.localName,
						remoteAddrString, err.Error())
					this.serverConnLock.Unlock()
					continue
				}
				if packet.MsgType != MSG_REQ_CREATE_CONN || len(packet.Data) != 8 {
					log.Printf("Client %s received a not MSG_REQ_CREATE_CONN packet from %s, drop", this.localName,
						remoteAddrString)
					this.serverConnLock.Unlock()
					continue
				}
				ts := int64(binary.BigEndian.Uint64(packet.Data))
				currentTime := int64(time.Now().Unix())
				if currentTime-ts > 5 || ts-currentTime > 5 {
					log.Printf("Client %s received a MSG_REQ_CREATE_CONN packet from %s, but timestamp error", this.localName,
						remoteAddrString)
					this.serverConnLock.Unlock()
					continue
				}
				log.Printf("Client %s received a MSG_REQ_CREATE_CONN packet from %s, start a new connection to server", this.localName,
					remoteAddrString)
				serverConn = NewAdvancedRelayConn(this.conn, remoteAddr, this.password, this.encryptMethod, this.encryptHeaderOnly, this.hashHeaderOnly, this.localName, this.otherData, this.sessionTimeout, this.bufSize)
				serverConn.ClientInit(this.target, this.targetIPVersion, this.compressType, this.clientListener)
				this.serverConn[remoteAddrString] = serverConn
			}
			if serverConn.IsClosed() == false {
				serverConn.recvChan <- data[:read_count]
			}
			this.serverConnLock.Unlock()
		}
	}
}

func (this *AdvancedRelayClient) GetConnByRemoteAddr(addr string) *AdvancedRelayConn {
	clientConn, ok := this.serverConn[addr]
	if !ok {
		return nil
	}
	return clientConn
}

//连接到服务端 没有锁。调用这个函数的函数要加锁
func (this *AdvancedRelayClient) connectToServer(remoteAddr *net.UDPAddr) {
	serverConn := this.GetConnByRemoteAddr(remoteAddr.String())
	if serverConn != nil {
		return
	}
	serverConn = NewAdvancedRelayConn(this.conn, remoteAddr, this.password, this.encryptMethod, this.encryptHeaderOnly, this.hashHeaderOnly, this.localName, this.otherData, this.sessionTimeout, this.bufSize)
	serverConn.ClientInit(this.target, this.targetIPVersion, this.compressType, this.clientListener)
	//serverConn.Connect()
	log.Printf("Client %s try to connect server: %s", this.localName, remoteAddr.String())
	this.serverConn[remoteAddr.String()] = serverConn
}

func (this *AdvancedRelayClient) SetTimeout(timeout int) {
	this.sessionTimeout = timeout
	log.Printf("Client %s session timeout set to %d", this.localName, this.sessionTimeout)
}

func (this *AdvancedRelayClient) SetSessionSave(timeout int) {
	log.Printf("Client %s save closed session duration set to %d", this.localName, this.saveClosedSession)
	this.saveClosedSession = timeout
}

//设置stun服务器
func (this *AdvancedRelayClient) SetStunServer(stun_server string) error {
	target, err := ResolveUDPAddr("udp4", stun_server, 100)
	if err != nil {
		return err
	}
	stun_session := NewStunSession(target, this.conn)
	this.session[target.String()] = stun_session
	this.stunServer = target.String()
	log.Printf("Client %s stun server set to %s", this.localName, this.stunServer)

	go stun_session.SendStunBindReqProc()
	return nil
}

//检查资源线程
func (this *AdvancedRelayClient) check_session_proc() {
	var stunSession CommonSession
	var currentTime int64
	for this.stat != ARCL_STAT_CLOSED {
		if this.stat == ARCL_STAT_CONNECTINT {
			//检查哪个子连接连接上了，并及时删除没有成功的连接
			time.Sleep(200 * time.Microsecond)
			connected := "no"
			this.serverConnLock.Lock()
			for k := range this.serverConn {
				if this.serverConn[k].connStat == ARC_STAT_ESTABLISHED {
					connected = k
					log.Printf("Client %s connect to server success via connection %s", this.localName, this.serverConn[k].remoteAddr.String())
					go this.serverConn[k].RecvUserClientDataProc()
					this.stat = ARCL_STAT_CONNECTED
				}
			}
			if connected != "no" {
				for k := range this.serverConn {
					if k != connected {
						log.Printf("Client %s delete unsuccessful connection %s", this.localName, this.serverConn[k].remoteAddr.String())
						this.serverConn[k].Close("other connection success")
						delete(this.serverConn, k)
					}
				}
			}
			this.serverConnLock.Unlock()
			//每60秒从tracker更新ip，并将自己的ip同步到tracker
			currentTime = time.Now().Unix()
			if (currentTime+this.randomOffset)%60 == 0 {
				this.GetAddrListFromTrackerAndConnect()
				this.RequestServerConnectClient()
				time.Sleep(1 * time.Second)
			}
		} else if this.stat == ARCL_STAT_CONNECTED {
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
			this.serverConnLock.Lock()
			reConnect := false
			for k := range this.serverConn {
				// if this.serverConn[k].connStat != ARC_STAT_ESTABLISHED {
				// 	log.Printf("Client %s delete unsuccessful connection %s", this.localName, this.serverConn[k].remoteAddr.String())
				// 	this.serverConn[k].Close("other connection success")
				// 	delete(this.serverConn, k)
				// }
				if this.serverConn[k].IsClosed() == true {
					// 与服务器的连接超时，断开连接。并重新开始连接
					delete(this.serverConn, k)
					if this.stat == ARCL_STAT_CLOSED {
						break
					}
					log.Printf("Client %s connection to server timeout.Retry connect to available server.", this.localName)
					reConnect = true
					break
				}
			}
			this.serverConnLock.Unlock()
			if reConnect == true {
				this.RestartConnection()
			}

			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

// 停止服务
func (this *AdvancedRelayClient) Close() {
	this.stat = ARCL_STAT_CLOSED

	for k := range this.session {
		this.session[k].Close("advanced relay client closed")
		log.Printf("Client %s closed.", this.localName)
	}
	this.serverConnLock.Lock()
	for k := range this.serverConn {
		this.serverConn[k].Close("advanced relay client closed")
		log.Printf("Client %s closed.", this.localName)
	}
	defer this.serverConnLock.Unlock()
	this.conn.Close()
	this.clientListener.Close()
}
