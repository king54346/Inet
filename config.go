package main

import (
	"encoding/json"
	"os"
)

type GatewayConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}
type ListenerConfig struct {
	ID               string                 `json:"id"`
	ClientID         string                 `json:"client_id"`
	PublicProtocol   string                 `json:"public_protocol"`
	PublicIP         string                 `json:"public_ip"`
	PublicPort       uint16                 `json:"public_port"`
	InternalProtocol string                 `json:"internal_protocol"`
	InternalIP       string                 `json:"internal_ip"`
	InternalPort     uint16                 `json:"internal_port"`
	HTTPRouteType    string                 `json:"http_route_type"`
	HTTPParam        map[string]interface{} `json:"http_param"`
}

func ParseListenerConfig(confFile string) ([]*ListenerConfig, error) {
	content, err := os.ReadFile(confFile)
	if err != nil {
		return nil, err
	}

	var cfg = make([]*ListenerConfig, 0)
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
