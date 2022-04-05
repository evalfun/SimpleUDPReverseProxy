package udprelay

const (
	connectionTimeout      = 60
	connectionPingDuration = 10

	ARC_STAT_INIT        uint16 = 0  // 等待初始化
	ARC_STAT_WAIT_CONN   uint16 = 1  // 等待客户端开始连接  服务器才会有此状态
	ARC_STAT_WAIT_ACK    uint16 = 3  // 等待确认连接        客户端才会有此状态
	ARC_STAT_READY       uint16 = 4  // 该发的都发了，但不知道客户端有没有收到 服务器才会有此状态  准备进入连接成功
	ARC_STAT_ESTABLISHED uint16 = 5  // 连接成功     收到正常传输的数据或ping后会转化为此状态
	ARC_STAT_CLOSED      uint16 = 20 // 连接断开

	ARC_ACT_CLIENT uint8 = 0
	ARC_ACT_SERVER uint8 = 1

	commonSignRecvAck  = 0   // 收到了ack报文
	commonSignClose    = 199 //连接关闭
	commonSignRecvConn = 24  //收到了创建连接请求
)
const ( // 客户端状态
	ARCL_STAT_INIT       = 0
	ARCL_STAT_CONNECTINT = 1
	ARCL_STAT_CONNECTED  = 2
	ARCL_STAT_CLOSED     = 20
)
const (
	MSG_REQ_CREATE_CONN uint8 = 2 // 请求创建连接

	MSG_CREATE_CONN     uint8 = 5 // 连接创建
	MSG_CREATE_CONN_ACK uint8 = 6 // 确认创建连接

	MSG_CLOSE_CONN uint8 = 10 // 连接关闭

	MSG_DATA uint8 = 20 // 传输数据
	MSG_PING uint8 = 30 // 心跳包

	//COMPRESS_TYPE 请求压缩类型
	COMPRESS_NONE uint8 = 0 // 不压缩
	COMPRESS_GZIP uint8 = 1 // GZIP 压缩
	COMPRESS_LZ4  uint8 = 2 // LZ4 压缩

	// 协议
	PROTO_TCP4 uint8 = 1
	PROTO_UDP4 uint8 = 2
	PROTO_TCP6 uint8 = 3
	PROTO_UDP6 uint8 = 4
	PROTO_TCP  uint8 = 5
	PROTO_UDP  uint8 = 6
)
