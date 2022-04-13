package main

import (
	"SimpleUDPReverseProxy/udprelay"
	"net"

	"github.com/gin-gonic/gin"
)

var udpRelayMap map[int]*udprelay.UDPRelay
var advancedRelayClientMap map[int]*udprelay.AdvancedRelayClient
var advancedRelayServerMap map[int]*udprelay.AdvancedRelayServer

func init() {
	advancedRelayClientMap = make(map[int]*udprelay.AdvancedRelayClient)
	advancedRelayServerMap = make(map[int]*udprelay.AdvancedRelayServer)
	udpRelayMap = make(map[int]*udprelay.UDPRelay)
}

func listAdvancedRelayServerHandler(c *gin.Context) {
	var resp map[int]*udprelay.AdvancedRelayServerStat
	resp = make(map[int]*udprelay.AdvancedRelayServerStat)
	for k := range advancedRelayServerMap {
		resp[k] = advancedRelayServerMap[k].GetServerStat()
	}
	c.JSON(200, resp)
}

func createAdvancedRelayServerHandler(c *gin.Context) {
	var req AdvancedUDPRelayServerConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}

	err := createAdvancedRelayServer(&req)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

func connectToClientHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
		ClientAddr string
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "Instance not found")
		return
	}
	if len(req.ClientAddr) < 3 {
		c.String(400, "Invalid client address")
		return
	}
	clientAddr, err := udprelay.ResolveUDPAddr("udp", req.ClientAddr, 10)
	if err != nil {
		c.String(400, "Can not resolve address:"+req.ClientAddr+" "+err.Error())
		return
	}
	instance.ReqClientConn(clientAddr)
	c.String(200, "success")
}

func getServerSessionHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	resp := instance.GetSessionDict()
	c.JSON(200, resp)
}

func getServerConnectionListHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	resp := instance.GetConnectionList()
	c.JSON(200, resp)
}

func setServerTrackerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
		ServerURL  string
		ServerID   string
		UserID     string
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	config := &udprelay.TrackerConfig{
		ServerID:  req.ServerID,
		ServerURL: req.ServerURL,
		UserID:    req.UserID,
	}
	err := instance.SetTrackerConfig(config)
	if err != nil {
		c.String(400, err.Error())
	} else {
		c.String(200, "success")
	}
}

func getServerTrackerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	type Resp struct {
		UserID    string
		ServerID  string
		ServerURL string
		Message   string
		Code      int
	}
	var resp Resp
	config := instance.GetTrackerConfig()
	if config == nil {
		c.JSON(200, resp)
		return
	}
	resp.ServerID = config.ServerID
	resp.ServerURL = config.ServerURL
	resp.UserID = config.UserID
	resp.Message, resp.Code = instance.GetTrackerStat()
	c.JSON(200, resp)
}

func deleteAdcancedRelayServerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayServerMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	instance.Close()
	delete(advancedRelayServerMap, req.InstanceID)
	c.String(200, "success")
}

func listAdvancedRelayClientHandler(c *gin.Context) {
	var resp map[int]*udprelay.AdvancedRelayClientStat
	resp = make(map[int]*udprelay.AdvancedRelayClientStat)
	for k := range advancedRelayClientMap {
		resp[k] = advancedRelayClientMap[k].GetClientStat()
	}
	c.JSON(200, resp)
}

func getClientConnectionListHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	resp := instance.GetConnectionList()
	c.JSON(200, resp)
}

func updateClientServerAddrHandler(c *gin.Context) {
	type Request struct {
		InstanceID     int
		ServerAddrList []string
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	var addrList []*net.UDPAddr
	for i := range req.ServerAddrList {
		if len(req.ServerAddrList[i]) < 4 {
			continue
		}
		addr, err := udprelay.ResolveUDPAddr("udp", req.ServerAddrList[i], 10)
		if err != nil {
			continue
		}
		addrList = append(addrList, addr)
	}
	instance.SetServerAddrList(addrList)
	c.String(200, "success")
}

func restartClientConnectionHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	instance.RestartConnection()
	c.JSON(200, "success")
}

func createAdvancedRelayClientHandler(c *gin.Context) {
	var req AdvancedUDPRelayClientConfig
	var err error
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	err = createAdvancedRelayClient(&req)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

func getClientSessionHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	resp := instance.GetSessionDict()
	c.JSON(200, resp)
}

func setClientTrackerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
		ServerURL  string
		ServerID   string
		UserID     string
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	config := &udprelay.TrackerConfig{
		ServerID:  req.ServerID,
		ServerURL: req.ServerURL,
		UserID:    req.UserID,
	}
	instance.SetTrackerConfig(config)
	c.String(200, "success")
}

func getClientTrackerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	type Resp struct {
		UserID    string
		ServerID  string
		ServerURL string
		Message   string
		AddrList  []string
	}
	var resp Resp
	config := instance.GetTrackerConfig()
	if config == nil {
		c.JSON(200, resp)
		return
	}
	resp.ServerID = config.ServerID
	resp.ServerURL = config.ServerURL
	resp.UserID = config.UserID
	resp.Message, resp.AddrList = instance.GetTrackerStat()
	c.JSON(200, resp)
}
func deleteAdcancedRelayClientHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	instance, ok := advancedRelayClientMap[req.InstanceID]
	if !ok {
		c.String(404, "instance not found")
		return
	}
	delete(advancedRelayClientMap, req.InstanceID)
	c.String(200, "success")
	instance.Close()
}

func listUDPRelayHandler(c *gin.Context) {
	type Resp struct {
		InstanceID int
		LocalAddr  string
		PublicAddr string
		StunServer string
	}
	var respList []Resp
	for k := range udpRelayMap {
		resp := Resp{
			InstanceID: k,
			LocalAddr:  udpRelayMap[k].Conn.LocalAddr().String(),
			PublicAddr: udpRelayMap[k].LocalPublicAddr,
			StunServer: udpRelayMap[k].StunServer,
		}
		respList = append(respList, resp)
	}
	c.JSON(200, respList)

}

func createUDPRelayHandler(c *gin.Context) {
	var request UDPRelayConfig
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(400, err.Error())
		return
	}
	err := createUDPRelay(&request)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

func deleteUDPRelayServerHandler(c *gin.Context) {
	type Request struct {
		InstanceID int
	}
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(400, err.Error())
		return
	}
	relayInstance, ok := udpRelayMap[req.InstanceID]
	if !ok {
		c.String(404, "target udp relay server not found")
		return
	}
	delete(udpRelayMap, req.InstanceID)
	relayInstance.Close()
	c.String(200, "success")
}

func saveConfigHandler(c *gin.Context) {
	err := saveConfig()
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.String(200, "success")
	return
}
