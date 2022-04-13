package main

import (
	"SimpleUDPReverseProxy/crypts"
	"SimpleUDPReverseProxy/udprelay"
	"errors"
	"net"
)

func createAdvancedRelayClient(config *AdvancedUDPRelayClientConfig) error {
	if config.InstanceID == 0 {
		for {
			config.InstanceID = config.InstanceID + 1
			_, ok := advancedRelayClientMap[config.InstanceID]
			if !ok {
				break
			}
		}
	}
	var err error
	instance, ok := advancedRelayClientMap[config.InstanceID]
	if ok {
		return errors.New("instance already exist")
	}
	var cryptMethod int
	cryptMethod = crypts.GetCryptMethodCode(config.CryptMethod)
	if cryptMethod == 0 {
		return errors.New("unkonwn crypt method " + config.CryptMethod)
	}
	if config.BufSize < 128 {
		return errors.New("recv BufSize can not smaller than 128")
	}
	instance, err = udprelay.NewAdvancedRelayClient(int(config.ListenerPort), int(config.LocalPort), config.BufSize, config.Target, config.TargetIPVersion, config.SessionTimeout, config.SaveClosedSession, []byte(config.Password), cryptMethod, false, false, []byte(config.LocalName), []byte(config.OtherData), config.CompressType)
	if err != nil {
		return errors.New("Create instance faild:" + err.Error())
	}
	advancedRelayClientMap[config.InstanceID] = instance
	if config.StunServer != "" {
		instance.SetStunServer(config.StunServer)
	}
	var serverAddrList []*net.UDPAddr
	for i := range config.ServerAddr {
		if config.ServerAddr[i] == "" {
			continue
		}
		serverAddr, err := udprelay.ResolveUDPAddr("udp", config.ServerAddr[i], 10)
		if err == nil {
			serverAddrList = append(serverAddrList, serverAddr)
		}
	}
	instance.SetServerAddrList(serverAddrList)
	instance.StartConnect()
	instance.SetTrackerConfig(config.Tracker)
	return nil
}

func createAdvancedRelayServer(config *AdvancedUDPRelayServerConfig) error {
	if config.InstanceID == 0 {
		for {
			config.InstanceID = config.InstanceID + 1
			_, ok := advancedRelayServerMap[config.InstanceID]
			if !ok {
				break
			}
		}
	}
	instance, ok := advancedRelayServerMap[config.InstanceID]
	if ok {
		return errors.New("instance already exist")
	}
	cryptMethod := crypts.GetCryptMethodCode(config.CryptMethod)
	if cryptMethod == 0 {
		return errors.New("unkonwn crypt method " + config.CryptMethod)
	}
	instance, err := udprelay.NewAdvancedRelayServer(int(config.LocalPort), config.BufSize, config.SessionTimeout, config.SaveClosedSession, []byte(config.Password), cryptMethod, false, false, []byte(config.LocalName), []byte(config.OtherData))
	if err != nil {
		return errors.New("Create instance error:" + err.Error())
	}
	advancedRelayServerMap[config.InstanceID] = instance

	instance.SetTargetAddr(config.TargetIPVersion, config.Target)
	if config.StunServer != "" {
		instance.SetStunServer(config.StunServer)
	}
	err = instance.SetTrackerConfig(config.Tracker)
	return err
}

func createUDPRelay(config *UDPRelayConfig) error {
	_, ok := udpRelayMap[config.InstanceID]
	if ok {
		return errors.New("instance already exist")
	}
	relay, err := udprelay.NewUDPRelay(config.LocalPort, config.BufSize)
	if err != nil {
		return err
	}
	relay.SetSessionSave(config.SaveClosedSession)
	relay.SetTimeout(config.SessionTimeout)
	relay.SetStunServer(config.StunServer)
	relay.SetTargetAddr(config.TargetIPVersion, config.TargetIPVersion)
	udpRelayMap[config.InstanceID] = relay
	return nil
}
