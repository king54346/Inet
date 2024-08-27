package main

import (
	"InNet/common"
	"fmt"
	"github.com/xtaci/smux"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	clientID   string
	serverAddr string
}

func NewClient(clientID string, serverAddr string) *Client {
	return &Client{
		clientID,
		serverAddr,
	}
}

func (c *Client) Run() {
	for {
		err := c.run()
		if err != nil && err != io.EOF {
			log.Error("%v", err)
		}
		log.Warn("reconnect %s", c.serverAddr)
		time.Sleep(time.Second * 1)
	}
}

func (c *Client) run() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 发送handshake包
	handshakeReq := common.HandshakeReq{ClientID: c.clientID}
	buf, err := handshakeReq.Encode()
	if err != nil {
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	_, err = conn.Write(buf)
	conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}

	// 创建mux session
	mux, err := smux.Client(conn, nil)
	if err != nil {
		return err
	}
	defer mux.Close()

	// 等待mux stream
	for {
		stream, err := mux.AcceptStream()
		if err != nil {
			return err
		}

		go c.handleStream(stream)
	}
}

func (c *Client) handleStream(stream net.Conn) {
	// Proxy Protocol 解码
	pp := &common.ProxyProtocol{}
	if err := pp.Decode(stream); err != nil {
		log.Printf("decode pp fail: %v", err)
		return
	}
	log.Printf("pp %+v", pp)

	// 与本地建立连接
	localConn, err := c.connectToLocal(pp)
	if err != nil {
		log.Printf("failed to connect to local: %v", err)
		return
	}
	// 双向数据拷贝
	c.proxyData(stream, localConn)
}

func (c *Client) connectToLocal(pp *common.ProxyProtocol) (net.Conn, error) {
	switch pp.InternalProtocol {
	case "tcp":
		return net.Dial("tcp", fmt.Sprintf("%s:%d", pp.InternalIP, pp.InternalPort))
	default:
		log.Printf("unsupported protocol %s", pp.InternalProtocol)
		return nil, fmt.Errorf("unsupported protocol")
	}
}

func (c *Client) proxyData(stream, localConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(localConn, stream)
		localConn.Close() // 关闭写端，通知另一方向数据传输完成
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(stream, localConn)
		stream.Close() // 关闭写端，通知另一方向数据传输完成
	}()

	wg.Wait() // 等待所有数据传输完成
}
