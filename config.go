package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"main/crypts"
	"main/udprelay"
	"strconv"
	"strings"
)

type AdvancedUDPRelayServerConfig struct {
	InstanceID        int
	LocalPort         int64
	BufSize           int
	Target            string
	TargetIPVersion   string
	SessionTimeout    int
	SaveClosedSession int
	Password          string
	CryptMethod       string
	EncryptHeaderOnly bool
	HashHeaderOnly    bool
	LocalName         string
	OtherData         string
	StunServer        string
	Tracker           *udprelay.TrackerConfig
}

type AdvancedUDPRelayClientConfig struct {
	InstanceID        int
	ListenerPort      int64
	LocalPort         int64
	BufSize           int
	Target            string
	TargetIPVersion   string
	SessionTimeout    int
	SaveClosedSession int
	Password          string
	CryptMethod       string
	EncryptHeaderOnly bool
	HashHeaderOnly    bool
	LocalName         string
	OtherData         string
	CompressType      uint8
	StunServer        string
	ServerAddr        []string
	Tracker           *udprelay.TrackerConfig
}

type UDPRelayConfig struct {
	InstanceID        int
	LocalPort         int
	StunServer        string
	BufSize           int
	SessionTimeout    int
	SaveClosedSession int
	Target            string
	TargetIPVersion   string
}

type RConfig struct {
	Server []*AdvancedUDPRelayServerConfig
	Client []*AdvancedUDPRelayClientConfig
	Relay  []*UDPRelayConfig
}

func loadConfig() error {
	configFileData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	var rconfig RConfig
	err = json.Unmarshal(configFileData, &rconfig)
	if err != nil {
		return err
	}
	for i := range rconfig.Client {
		err = createAdvancedRelayClient(rconfig.Client[i])
		if err != nil {
			log.Println(err)
		}
	}
	for i := range rconfig.Server {
		err = createAdvancedRelayServer(rconfig.Server[i])
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func saveConfig() error {
	var rconfig RConfig
	for i := range advancedRelayClientMap {
		var config AdvancedUDPRelayClientConfig
		stat := advancedRelayClientMap[i].GetClientStat()
		config.InstanceID = i
		_t := strings.Split(stat.LocalAddr, ":")
		l := _t[len(_t)-1]
		var err error
		config.LocalPort, err = strconv.ParseInt(l, 10, 32)
		if err != nil {
			log.Println("config port error", err, l, stat.LocalAddr)
		}
		_t = strings.Split(stat.ListenerAddr, ":")
		l = _t[len(_t)-1]
		config.ListenerPort, err = strconv.ParseInt(l, 10, 32)
		config.BufSize = stat.BufSize
		config.Target = stat.Target
		config.TargetIPVersion = stat.TargetIPVersion
		config.SessionTimeout = stat.SessionTimeout
		config.SaveClosedSession = stat.SaveClosedSession
		config.Password = stat.Password
		config.CryptMethod = crypts.GetCryptMethodStr(stat.CryptMethod)
		config.EncryptHeaderOnly = stat.EncryptHeaderOnly
		config.LocalName = stat.LocalName
		config.OtherData = stat.OtherData
		config.StunServer = stat.StunServer
		config.ServerAddr = stat.ServerAddrList
		config.HashHeaderOnly = stat.HashHeaderOnly
		config.Tracker = advancedRelayClientMap[i].GetTrackerConfig()
		rconfig.Client = append(rconfig.Client, &config)
	}
	for i := range advancedRelayServerMap {
		var config AdvancedUDPRelayServerConfig
		var err error
		stat := advancedRelayServerMap[i].GetServerStat()
		config.InstanceID = i
		_t := strings.Split(stat.LocalAddr, ":")
		l := _t[len(_t)-1]
		config.LocalPort, err = strconv.ParseInt(l, 10, 32)
		if err != nil {
			log.Println("config port error", err)
		}
		config.BufSize = stat.BufSize
		config.Target = stat.Target
		config.TargetIPVersion = stat.TargetIPVersion
		config.SessionTimeout = stat.SessionTimeout
		config.SaveClosedSession = stat.SaveClosedSession
		config.Password = stat.Password
		config.CryptMethod = crypts.GetCryptMethodStr(stat.EncryptMethod)
		config.EncryptHeaderOnly = stat.EncryptHeaderOnly
		config.LocalName = stat.LocalName
		config.OtherData = stat.OtherData
		config.StunServer = stat.StunServer
		config.HashHeaderOnly = stat.HashHeaderOnly
		config.Tracker = advancedRelayServerMap[i].GetTrackerConfig()
		rconfig.Server = append(rconfig.Server, &config)
	}
	data, err := json.Marshal(rconfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFilePath, data, 644)
	if err != nil {
		return err
	}
	return nil
}
