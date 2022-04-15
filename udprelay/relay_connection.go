package udprelay

import (
	"SimpleUDPReverseProxy/crypts"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ConnectionStat struct {
	RemoteAddr      string
	ConnectStat     uint16
	TargetAddr      string
	TargetIPVersion string
	RemoteName      string
	RemoteOtherData string
	NewPassword     string
	SendSize        int64
	RecvSize        int64
	LastRecv        int64
	LastSend        int64
	RecvPacketCount uint32
	RecvPacketSN    uint32
	CompressType    uint8
	CloseResult     string
	CreateTime      int64
}

// 一个用户连接里面可以有很多个UDPSession
type AdvancedRelayConn struct {
	conn               *net.UDPConn // 与对端的连接 需要初始化
	session            map[uint16]CommonSession
	clientSessionIDMap map[uint16]string // 通过sessionid获取对端地址 客户端用
	clientAddrMap      map[string]uint16 // 通过对端地址获取sessionid 客户端用
	remoteAddr         *net.UDPAddr      // 对端地址 需要初始化
	remoteAddrString   string
	clientListener     *net.UDPConn // 本地端口 客户端需要初始化
	bufSize            int          // UDP接收缓冲区大小，需要初始化
	peerName           []byte       // 对端名称
	peerOtherData      []byte       // 对端附加信息
	localName          []byte       // 本端名称 需要初始化
	otherData          []byte       // 发送给对端的消息 需要初始化
	encryptHeaderOnly  bool         // 只加密报文头部   暂时弃用
	hashHeaderOnly     bool         // 只校验报文头部   暂时弃用
	encryptMethod      int
	act                uint8  // 我是服务端还是客户端？ 客户端要初始化
	targetAddr         string // 客户端要初始化 服务端靠自动协商
	targetIPVersion    string // 客户端要初始化 服务端靠自动协商
	password           []byte // 需要初始化
	newPasswd          []byte // 新协商的密码
	compressType       uint8  // 客户端要初始化，服务器协商
	sendPacketSN       uint32
	sendSize           int64
	recvSize           int64
	recvPacketCount    uint32
	recvPacketSn       uint32
	connStat           uint16
	lastRecv           int64
	lastSend           int64
	recvChan           chan []byte //从对端收到信息了
	commonChan         chan int
	closeResult        string
	timeout            int // 需要初始化
	sessionTimeout     int
	saveClosedSession  int
	clientSessionID    uint16
	createTime         int64
	sessionLock        sync.RWMutex
	sendLock           sync.Mutex

	cryptInstance          crypts.Cryption
	cryptInstanceNewPasswd crypts.Cryption
}

//参数：本端连接， 对端地址，密码，加密方式，是否只加密头部，本端名称，本端注释，超时时间
func NewAdvancedRelayConn(conn *net.UDPConn, addr *net.UDPAddr,
	password []byte, encryptMethod int, encryptHeaderOnly bool, hashHeaderOnly bool, localName []byte, otherData []byte,
	timeout int, bufSize int) *AdvancedRelayConn {
	arc := new(AdvancedRelayConn)
	arc.conn = conn
	arc.remoteAddr = addr
	arc.remoteAddrString = addr.String()
	arc.connStat = ARC_STAT_INIT
	arc.password = password
	arc.localName = localName
	arc.otherData = otherData
	arc.sessionTimeout = timeout
	arc.saveClosedSession = timeout
	arc.encryptMethod = encryptMethod
	arc.hashHeaderOnly = hashHeaderOnly
	arc.newPasswd = make([]byte, 16)
	rand.Read(arc.newPasswd)
	if password != nil {
		arc.encryptHeaderOnly = encryptHeaderOnly
	}
	arc.timeout = timeout
	arc.session = make(map[uint16]CommonSession)
	arc.clientSessionIDMap = make(map[uint16]string)
	arc.clientAddrMap = make(map[string]uint16)
	arc.bufSize = bufSize
	arc.lastRecv = time.Now().Unix()
	arc.lastSend = time.Now().Unix()
	arc.recvChan = make(chan []byte, 20)
	arc.commonChan = make(chan int, 5)
	arc.act = ARC_ACT_SERVER
	arc.createTime = time.Now().Unix()
	arc.cryptInstance, _ = crypts.NewCryption(arc.encryptMethod, arc.password, passwd_salt)
	arc.cryptInstanceNewPasswd, _ = crypts.NewCryption(arc.encryptMethod, arc.newPasswd, passwd_salt)
	go arc.recvPacketProc()
	go arc.checkSessionProc()
	return arc
}

//客户端初始化
// 用户目标服务器地址，协议版本，压缩类型，用户数据入站连接
func (this *AdvancedRelayConn) ClientInit(targetAddr string, targetIPVersion string, compressType uint8, clientListener *net.UDPConn) {
	this.act = ARC_ACT_CLIENT
	this.targetAddr = targetAddr
	this.targetIPVersion = targetIPVersion
	this.compressType = compressType
	this.clientListener = clientListener
	this.Connect()
}

func (this *AdvancedRelayConn) SetTargetAddr(targetAddr string, targetIPVersion string) {
	this.targetAddr = targetAddr
	this.targetIPVersion = targetIPVersion
}

//获取更多状态
func (this *AdvancedRelayConn) GetConnStat() *ConnectionStat {
	stat := &ConnectionStat{
		RemoteAddr:      this.remoteAddrString,
		ConnectStat:     this.connStat,
		TargetAddr:      this.targetAddr,
		TargetIPVersion: this.targetIPVersion,
		RemoteName:      string(this.peerName),
		RemoteOtherData: string(this.peerOtherData),
		SendSize:        this.sendSize,
		RecvSize:        this.recvSize,
		LastRecv:        this.lastRecv,
		LastSend:        this.lastSend,
		CompressType:    this.compressType,
		CloseResult:     this.closeResult,
		CreateTime:      this.createTime,
		RecvPacketCount: this.recvPacketCount,
		RecvPacketSN:    this.recvPacketSn,
		NewPassword:     fmt.Sprintf("%x", this.newPasswd),
	}
	return stat
}
func (this *AdvancedRelayConn) IsClosed() bool {
	if this.connStat == ARC_STAT_CLOSED {
		return true
	}
	return false
}

//获取状态 lastSend lastRecv connStat remoteAddr.String() peerName peerOtherData
func (this *AdvancedRelayConn) GetConnInfo() (int64, int64, uint16, string, []byte, []byte) {
	return this.lastSend, this.lastRecv, this.connStat, this.remoteAddrString, this.peerName, this.peerOtherData
}

func (this *AdvancedRelayConn) GetSessionList() []*SessionStat {
	var statlist []*SessionStat
	this.sessionLock.RLock()
	defer this.sessionLock.RUnlock()
	for k := range this.session {
		var stat SessionStat
		stat.SendBytes, stat.RecvBytes, stat.CreateTime, stat.ClosedTime, _, _ = this.session[k].GetSessionInfo()
		stat.TargetAddr = this.session[k].GetTargetAddr()
		statlist = append(statlist, &stat)
	}
	return statlist
}

//只有服务端使用 开始请求客户端连接
func (this *AdvancedRelayConn) RequestConnect() {
	if this.act != ARC_ACT_SERVER {
		return
	}
	if this.connStat != ARC_STAT_INIT {
		return
	}
	this.connStat = ARC_STAT_WAIT_CONN
	go this.requestConnectProc()
}

// 服务端使用！！
func (this *AdvancedRelayConn) requestConnectProc() {
	//一直向客户端发送请求连接报文，直到超时
	// 将当前时间写进数据包。客户端收到之后要对比当前时间，防止重放攻击
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(time.Now().Unix()))
	this.sendPacket(MSG_REQ_CREATE_CONN, 0, data)
	timer := time.NewTimer(2 * time.Second)
	for this.connStat == ARC_STAT_WAIT_CONN {
		select {
		case <-timer.C:
			timer = time.NewTimer(4 * time.Second)
			binary.BigEndian.PutUint64(data, uint64(time.Now().Unix()))
			this.sendPacket(MSG_REQ_CREATE_CONN, 0, data)
		case <-this.commonChan:
			timer.Stop()
			return
		}
	}
}

//开始尝试连接
func (this *AdvancedRelayConn) Connect() {
	go this.tryConnectProc()
}

//客户端使用！！！
func (this *AdvancedRelayConn) tryConnectProc() {
	if this.act != ARC_ACT_CLIENT {
		return
	}
	if this.connStat != ARC_STAT_INIT {
		return
	}
	this.connStat = ARC_STAT_WAIT_ACK
	//第一阶段，发送请求连接报文
	createConnInfo := &CreateConnInfo{
		ReqCompressType: COMPRESS_NONE,
		NetworkType:     this.targetIPVersion,
		TargetAddr:      []byte(this.targetAddr),
		TimeStamp:       uint64(time.Now().Unix()),
		PeerName:        this.localName,
		OtherData:       this.otherData,
		NewPasswd:       make([]byte, 16),
	}
	rand.Read(createConnInfo.NewPasswd)
	this.newPasswd = createConnInfo.NewPasswd
	this.cryptInstanceNewPasswd.SetPassword(this.newPasswd, passwd_salt)
	createConnPacket, err := createConnInfo.PackCreateConnInfo(this.password)
	if err != nil {
		this.log(fmt.Sprintf("pack create connection info error:%s", err.Error()))
		return
	}
	this.sendPacket(MSG_CREATE_CONN, 0, createConnPacket)
	this.log(fmt.Sprintf("Try connect to %s,push new password %x\n", this.remoteAddrString, this.newPasswd))
	timer := time.NewTimer(5 * time.Second)
	for this.connStat == ARC_STAT_WAIT_ACK {
		select {
		case <-timer.C: //好像超时了，再次发送建立连接报文
			rand.Read(createConnInfo.NewPasswd)
			this.newPasswd = createConnInfo.NewPasswd
			this.cryptInstanceNewPasswd.SetPassword(this.newPasswd, passwd_salt)
			createConnInfo.TimeStamp = uint64(time.Now().Unix())
			createConnPacket, err = createConnInfo.PackCreateConnInfo(this.password)
			this.log(fmt.Sprintf("Try connect to %s,push new password %x\n", this.remoteAddrString, this.newPasswd))
			if err != nil {
				this.log(fmt.Sprintf("Pack create connection info error:%s", err.Error()))
				return
			}
			this.sendPacket(MSG_CREATE_CONN, 0, createConnPacket)
			timer = time.NewTimer(5 * time.Second)
		case <-this.commonChan: // 收到了ACK报文，进入第二阶段 接收ack 或者退出
			timer.Stop()
			return
		}
	}
}

func (this *AdvancedRelayConn) sendPacket(msgType uint8, sessionID uint16, data []byte) {
	this.sendLock.Lock()
	defer this.sendLock.Unlock()
	this.sendPacketSN = this.sendPacketSN + 1
	packet := &Packet{
		MsgType:   msgType,
		SessionID: sessionID,
		Data:      data,
		SN:        this.sendPacketSN,
	}
	//time.Sleep(1 * time.Second)
	var err error
	var encryptedPacket []byte
	var cryptInstance crypts.Cryption
	if this.connStat == ARC_STAT_ESTABLISHED || this.connStat == ARC_STAT_READY {
		cryptInstance = this.cryptInstanceNewPasswd
	} else {
		cryptInstance = this.cryptInstance
	}

	encryptedPacket, err = packet.EncryptPacket(cryptInstance, this.compressType, this.hashHeaderOnly)

	if err != nil {
		this.log(fmt.Sprintf("encrypt packet error:%s", err.Error()))
		return
	}
	sendSize, err := this.conn.WriteToUDP(encryptedPacket, this.remoteAddr)
	if err != nil {
		this.log(fmt.Sprintf("Send encrypted packet error:%s", err.Error()))
		return
	}
	//this.log(fmt.Sprintf("EncryptPacket %x, %d, %v", passwd, this.encryptMethod, this.hashHeaderOnly))
	this.sendSize = this.sendSize + int64(sendSize)
	if msgType != MSG_DATA {
		this.lastSend = time.Now().Unix()
	}
	//this.lastSend = time.Now().Unix() 由于性能问题，不统计
}

// 只有服务器才要发送ack报文
// 正常传输数据后不能再发送ack报文
func (this *AdvancedRelayConn) sendAck() {
	if this.connStat == ARC_STAT_CLOSED || this.connStat == ARC_STAT_ESTABLISHED {
		return
	}
	ackInfo := &AckInfo{
		Result:    0,
		PeerName:  this.localName,
		OtherData: this.otherData,
	}
	ackPacket, err := ackInfo.PackAckInfo(this.newPasswd)
	if err != nil {
		this.log(fmt.Sprintf("Pack ack info error:%s", err.Error()))
		return
	}
	this.sendPacket(MSG_CREATE_CONN_ACK, 0, ackPacket)
}

//关闭连接
func (this *AdvancedRelayConn) Close(r string) {
	for i := 0; i < 3; i++ {
		this.sendPacket(MSG_CLOSE_CONN, 0, nil)
		time.Sleep(50 * time.Microsecond)
	}
	this.closeWithoutSendMSG(r)
	return
}

//关闭连接
func (this *AdvancedRelayConn) closeWithoutSendMSG(r string) {
	if this.connStat == ARC_STAT_CLOSED {
		return
	}
	this.closeResult = r
	this.connStat = ARC_STAT_CLOSED
	this.commonChan <- commonSignClose
	close(this.recvChan)
	close(this.commonChan)
	this.sessionLock.RLock()
	defer this.sessionLock.RUnlock()
	//服务端还要关闭session
	if this.act == ARC_ACT_SERVER {
		for k := range this.session {
			this.session[k].Close("father connection closed")
		}
	}
	return
}

func (this *AdvancedRelayConn) invalidOriginalPasswd() {
	if this.password != nil {
		this.password = nil
		nilPassword := make([]byte, 16)
		rand.Read(nilPassword)
		this.cryptInstance.SetPassword(nilPassword, passwd_salt)
	}
}

// 接收报文！！！此协程每个实例一个
func (this *AdvancedRelayConn) recvPacketProc() {
	var decrypted_packet *Packet
	var err error
	for this.connStat != ARC_STAT_CLOSED {
		encrypted_packet, ok := <-this.recvChan
		if !ok {
			return
		}
		//先开始解密,使用新的密码
		decrypted_packet, err = DecryptPacket(encrypted_packet, this.cryptInstanceNewPasswd, this.hashHeaderOnly)

		if err != nil {
			// 新密码无法解密时使用旧密码

			decrypted_packet, err = DecryptPacket(encrypted_packet, this.cryptInstance, this.hashHeaderOnly)

			if err != nil { // 旧密码也无法解密，那就gun吧
				this.log(fmt.Sprintf("Decrypt packet error:%s", err.Error()))
				continue
			}
		}

		//对一下序列号，防止重放攻击
		if this.recvPacketSn >= decrypted_packet.SN+5 {
			this.log(fmt.Sprintf("Replay packet detected:%d, current %d, msg type %d\n",
				decrypted_packet.SN, this.recvPacketSn, decrypted_packet.MsgType))
			continue
		}
		this.recvPacketCount = this.recvPacketCount + 1
		this.recvPacketSn = decrypted_packet.SN
		this.recvSize = this.recvSize + int64(len(encrypted_packet))
		switch decrypted_packet.MsgType {
		//客户端收到
		case MSG_CLOSE_CONN: //关闭连接
			if this.connStat != ARC_STAT_ESTABLISHED {
				continue
			}
			if this.connStat == ARC_STAT_CLOSED {
				return
			}
			this.log("Connection closed by remote")
			this.closeWithoutSendMSG("Connection closed by remote")
			return
		case MSG_DATA: //收到了要转发的数据
			//this.lastRecv = time.Now().Unix() 性能问题，不统计
			if this.connStat != ARC_STAT_ESTABLISHED && this.connStat != ARC_STAT_READY {
				continue
			}
			if this.act == ARC_ACT_SERVER { // 服务端
				if this.connStat == ARC_STAT_READY {
					this.log("Recv data from client, connection enter ESTABLISHED stat.")
					this.invalidOriginalPasswd()
					this.connStat = ARC_STAT_ESTABLISHED
				}
				var session *UDPSession
				// 收到了一个新的session后就创建session
				this.sessionLock.Lock()
				_session, ok := this.session[decrypted_packet.SessionID]
				if !ok { //新的session
					targetAddr, err := ResolveUDPAddr(this.targetIPVersion, this.targetAddr, 10)
					if err != nil {
						this.log(fmt.Sprintf("Can not resolve remote addr %s: %s\n", this.targetAddr, err.Error()))
						this.sessionLock.Unlock()
						continue
					}
					session, err = NewUDPSession(this.targetIPVersion, targetAddr, this.bufSize)
					if err != nil {
						this.log(fmt.Sprintf("Can not create session %d to target user server %s %s: %s\n", decrypted_packet.SessionID, this.targetIPVersion, this.targetAddr, err.Error()))
						this.sessionLock.Unlock()
						continue
					}
					go this.RecvUserServerDataProc(decrypted_packet.SessionID, session)
					this.session[decrypted_packet.SessionID] = session
					this.log(fmt.Sprintf("Created a new udp session %d to %s %s", decrypted_packet.SessionID, this.targetIPVersion,
						this.targetAddr))
				} else {
					session = _session.(*UDPSession)
				}
				this.sessionLock.Unlock()
				// 将数据传给客户端
				err = session.Send(decrypted_packet.Data)
				if err != nil {
					this.log(fmt.Sprintf("Can not send data to target user server %s: %s", this.targetAddr, err.Error()))
					continue
				}
			} else { //客户端
				// 收到未知session后直接丢弃
				var session *IncomeUDPSession
				this.sessionLock.RLock()
				_session, ok := this.session[decrypted_packet.SessionID]
				this.sessionLock.RUnlock()
				if !ok {
					this.log(fmt.Sprintf("Recv unknow session from srever %d", decrypted_packet.SessionID))
					continue
				}
				//将数据传给用户客户端
				session = _session.(*IncomeUDPSession)
				err = session.Send(decrypted_packet.Data)
				if err != nil {
					this.log(fmt.Sprintf("Can not send data to target user client %s: %s", this.targetAddr, err.Error()))
					continue
				}
			}
		case MSG_CREATE_CONN: // 创建连接，只有服务端才会收到
			// 此时服务端的状态是INIT或者WAIT_CONN
			if this.act != ARC_ACT_SERVER {
				continue
			}
			if this.connStat == ARC_STAT_ESTABLISHED {
				continue
			}
			createConnInfo, err := UnpackCreateConnInfo(decrypted_packet.Data, this.password)
			if err != nil {
				this.log(fmt.Sprintf("Unpack create connection info error:%s\n", err.Error()))
				continue
			}
			// 给你5秒钟的延迟
			currentTime := uint64(time.Now().Unix())
			if currentTime-10 > createConnInfo.TimeStamp || currentTime+10 < createConnInfo.TimeStamp {
				this.log("Unable to establish connection: wrong time stamp")
				continue
			}
			// 更新一下目标地址信息 如果需要
			this.lastRecv = time.Now().Unix()
			if this.targetAddr == "" {
				this.targetAddr = string(createConnInfo.TargetAddr)
				this.targetIPVersion = createConnInfo.NetworkType
				this.log(fmt.Sprintf("Set target user server to %s %s", createConnInfo.NetworkType, createConnInfo.TargetAddr))
			}
			this.newPasswd = createConnInfo.NewPasswd
			this.cryptInstanceNewPasswd.SetPassword(this.newPasswd, passwd_salt)
			this.compressType = createConnInfo.ReqCompressType
			this.peerName = createConnInfo.PeerName
			this.peerOtherData = createConnInfo.OtherData
			this.log(fmt.Sprintf("Remote create connection info: udp:%s, newpass:%x, otherdata:%s, peername:%s, target: %s, ts:%d", createConnInfo.NetworkType, this.newPasswd, this.peerOtherData, this.peerName, this.targetAddr, createConnInfo.TimeStamp))
			// 发送ack信息 进入READY状态
			this.connStat = ARC_STAT_READY
			this.log(fmt.Sprintf("Accepted a connection from remote,new password set to %x, enter Ready stat", this.newPasswd))
			this.sendAck()
		case MSG_CREATE_CONN_ACK: //客户端 收到了ACK，与服务器握手成功, 状态变为已建立连接
			//客户端建立连接后，可以开始转发用户客户端数据了。
			if this.act != ARC_ACT_CLIENT {
				continue
			}
			if this.connStat != ARC_STAT_WAIT_ACK {
				continue
			}
			ackInfo, err := UnpackAckInfo(decrypted_packet.Data, this.newPasswd)
			if err != nil {
				this.log(fmt.Sprintf("Recv invaild ack packet from remote: %s", err.Error()))
			}
			this.lastRecv = time.Now().Unix()
			this.connStat = ARC_STAT_ESTABLISHED
			this.commonChan <- commonSignRecvAck
			this.peerName = ackInfo.PeerName
			this.peerOtherData = ackInfo.OtherData
			//开始ping和传输数据
			this.sendPacket(MSG_PING, 0, nil)
			this.invalidOriginalPasswd()
			this.log("Recv ack from server, enter ESTABLISHED stat. remote name: " + string(this.peerName))
			//go this.RecvUserClientDataProc()
		case MSG_PING:
			if this.connStat != ARC_STAT_ESTABLISHED && this.connStat != ARC_STAT_READY {
				continue
			}
			this.lastRecv = time.Now().Unix()
			if this.connStat == ARC_STAT_READY { //服务端，收到ping报文后会话建立成功
				this.log("Recv ping from client, enter ESTABLISHED stat")
				this.connStat = ARC_STAT_ESTABLISHED
			}
		}
	}
}

//接受用户服务器数据，发送数据给客户端 此协程由其它协程开启 服务器专用
func (this *AdvancedRelayConn) RecvUserServerDataProc(sessionID uint16, session *UDPSession) {
	data := make([]byte, this.bufSize)
	for this.connStat == ARC_STAT_ESTABLISHED {
		count, err := session.Recv(data)
		if err != nil {
			return
		}
		this.sendPacket(MSG_DATA, sessionID, data[:count])
	}
}

//接收用户客户端数据，发送数据给服务器  此协程由其它协程开启 客户端专用
//每个实例下，此协程只需开启一个
func (this *AdvancedRelayConn) RecvUserClientDataProc() {
	for this.connStat == ARC_STAT_ESTABLISHED {
		data := make([]byte, this.bufSize)
		read_count, userClientAddr, err := this.clientListener.ReadFromUDP(data)
		userClientAddrString := userClientAddr.String()
		if err != nil {
			return
		}
		// 创建或者获取session
		var session *IncomeUDPSession
		var sessionID uint16
		this.sessionLock.Lock()
		sessionID, ok := this.clientAddrMap[userClientAddrString]
		_session, ok1 := this.session[sessionID]
		if !ok || !ok1 {
			// 创建一个session
			session = &IncomeUDPSession{}
			session.InitUDPSession(this.clientListener, userClientAddr)
			for { // 找一个能用的sessionID
				sessionID = this.clientSessionID
				this.clientSessionID = this.clientSessionID + 1
				_, ok := this.clientSessionIDMap[sessionID]
				if !ok {
					break
				}
			}
			this.clientSessionIDMap[sessionID] = userClientAddrString
			this.clientAddrMap[userClientAddrString] = sessionID
			this.session[sessionID] = session
			this.log(fmt.Sprintf("Created a new session %d from %s\n", sessionID, userClientAddrString))
		} else {
			session = _session.(*IncomeUDPSession)
		}
		session.LastRecv = time.Now().Unix()
		session.RecvBytes = session.RecvBytes + int64(read_count)
		// 发送数据给对端
		this.sendPacket(MSG_DATA, sessionID, data[:read_count])
		this.sessionLock.Unlock()
	}
}

//此协程每个实例一个
func (this *AdvancedRelayConn) checkSessionProc() {
	for this.connStat != ARC_STAT_CLOSED {
		currentTime := time.Now().Unix()
		if this.connStat == ARC_STAT_ESTABLISHED {
			if currentTime%int64(connectionPingDuration) == 0 {
				this.sendPacket(MSG_PING, 0, nil)
			}
		}

		if this.connStat == ARC_STAT_WAIT_CONN {
			if currentTime-this.lastRecv > int64(connectionTimeout*2) {
				this.log("Connection timeout.Did not receive any packet from Client.")
				this.Close("timeout. did not receive any packet from Client")
				return
			}
		} else if this.connStat != ARC_STAT_WAIT_ACK {
			if currentTime-this.lastRecv > int64(connectionTimeout) {
				this.log("Connection timeout.Did not receive any packet from remote.")
				this.Close("timeout. did not receive any packet from remote")
				return
			}
		}
		this.sessionLock.Lock()
		for k := range this.session {
			session, ok := this.session[k].(CommonSession)
			if !ok {
				this.log(fmt.Sprintf("%s Type error: %v\n", this.remoteAddrString, this.session[k]))
			}
			_, _, _, CloseTime, LastRecv, LastSend := session.GetSessionInfo()
			if currentTime-LastRecv > int64(this.sessionTimeout) && currentTime-LastSend > int64(this.sessionTimeout) && session.IsClosed() == false {
				this.log(fmt.Sprintf("Session %d time out.Last active:r%d s%d current:%d timeout:%d", k, LastRecv, LastSend, currentTime, this.sessionTimeout))
				session.Close("session timeout")
			}
			if currentTime-CloseTime >= int64(this.saveClosedSession) && this.session[k].IsClosed() {
				for clientAddr := range this.clientAddrMap {
					if this.clientAddrMap[clientAddr] == k {
						delete(this.clientAddrMap, clientAddr)
						break
					}
				}
				delete(this.clientSessionIDMap, k)
				delete(this.session, k)
			}
		}
		this.sessionLock.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (this *AdvancedRelayConn) log(s string) {
	if this.act == ARC_ACT_SERVER {
		log.Printf("Server %s connection remote %s : %s", this.localName, this.remoteAddrString, s)
	} else {
		log.Printf("Client %s connection remote %s : %s", this.localName, this.remoteAddrString, s)
	}

}
