package main

import (
	"InNet/common"
	"net"
	"time"
)

type Gateway struct {
	conf       *GatewayConfig
	clientIDs  map[string]struct{}
	sessionMgr *SessionManager
}

func NewGateway(conf *GatewayConfig, sessionMgr *SessionManager) *Gateway {
	gw := &Gateway{
		conf:       conf,
		sessionMgr: sessionMgr,
	}
	go gw.checkOnlineInterval()
	return gw
}

func (gw *Gateway) SetAvailableClientIDs(clientIDs []string) {
	clientIDsMap := make(map[string]struct{})
	for _, clientID := range clientIDs {
		clientIDsMap[clientID] = struct{}{}
	}
	gw.clientIDs = clientIDsMap
}

func (gw *Gateway) ListenAndServe() error {
	listener, err := net.Listen("tcp", gw.conf.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		//创建server和client之间的session，如果client重启会存在session过期的现象
		go gw.handleConn(conn)
	}
}

func (gw *Gateway) handleConn(conn net.Conn) {
	handshakeReq := &common.HandshakeReq{}
	err := handshakeReq.Decode(conn)
	if err != nil {
		log.Error("decode handshake fail: %v", err)
		return
	}

	if _, ok := gw.clientIDs[handshakeReq.ClientID]; !ok {
		log.Warn("client %s is not configured", handshakeReq.ClientID)
		return
	}

	log.Info("handshake from %s", handshakeReq.ClientID)

	_, err = gw.sessionMgr.CreateSession(handshakeReq.ClientID, conn)
	if err != nil {
		log.Error("create session fail: %v", err)
		return
	}
}

func (gw *Gateway) checkOnlineInterval() {
	tick := time.NewTicker(time.Second * 1)
	defer tick.Stop()
	for range tick.C {
		gw.sessionMgr.Range(func(k string, v *Session) bool {
			if v.Connection.IsClosed() {
				log.Info("session %s is offline", v.ClientID)
				return false
			}
			log.Info("session %s is online", v.ClientID)
			return true
		})
	}
}
