package main

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	log.SetLevel(logrus.InfoLevel)
	sessionMgr := NewSessionManager()
	listenerConfig := &ListenerConfig{
		ClientID:         "1",
		PublicProtocol:   "tcp",
		PublicIP:         "0.0.0.0",
		PublicPort:       10000,
		InternalProtocol: "tcp",
		InternalIP:       "127.0.0.1",
		InternalPort:     2000,
	}
	listener := NewListener(listenerConfig, sessionMgr)
	go func() {
		listener.ListenAndServe()
	}()
	listenerMgr := NewListenerManager()
	listenerMgr.AddListener(listenerConfig.ID, listener)
	clientIDs := make([]string, 0)
	clientIDs = append(clientIDs, "1")

	//建立公网服务端和内网客户端的连接
	gc := &GatewayConfig{
		ListenAddr: "127.0.0.1:8000",
	}
	gw := NewGateway(gc, sessionMgr)
	gw.SetAvailableClientIDs(clientIDs)
	err := gw.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
