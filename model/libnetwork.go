package model

import (
	"encoding/json"
	"net"
)

func ParseCIDR(cidr string) (n *net.IPNet, e error) {
	var i net.IP
	if i, n, e = net.ParseCIDR(cidr); e == nil {
		n.IP = i
	}
	return
}

// IPAMData represents the per-network ip related
// operational information libnetwork will send
// to the network driver during CreateNetwork()
type IPAMData struct {
	AddressSpace string
	Pool         *net.IPNet
	Gateway      *net.IPNet
}

func (i *IPAMData) UnmarshalJSON(data []byte) error {
	var (
		m   map[string]interface{}
		err error
	)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	i.AddressSpace = m["AddressSpace"].(string)
	if v, ok := m["Pool"]; ok {
		if i.Pool, err = ParseCIDR(v.(string)); err != nil {
			return err
		}
	}
	if v, ok := m["Gateway"]; ok {
		if i.Gateway, err = ParseCIDR(v.(string)); err != nil {
			return err
		}
	}
	return nil
}

type IpamInfo struct {
	PoolID string
	Meta   map[string]string
	IPAMData
}

func (i *IpamInfo) UnmarshalJSON(data []byte) error {
	var (
		m   map[string]interface{}
		err error
	)
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	i.PoolID = m["PoolID"].(string)
	if v, ok := m["IPAMData"]; ok {
		if err = json.Unmarshal([]byte(v.(string)), &i.IPAMData); err != nil {
			return err
		}
	}
	return nil
}

type Network struct {
	Id          string
	NetworkType string
	IPAMV4Info  []*IpamInfo
}

func (n *Network) UnmarshalJSON(b []byte) error {
	var netMap map[string]interface{}
	if err := json.Unmarshal(b, &netMap); err != nil {
		return err
	}
	n.Id = netMap["id"].(string)
	n.NetworkType = netMap["networkType"].(string)
	if v, ok := netMap["ipamV4Info"]; ok {
		if err := json.Unmarshal([]byte(v.(string)), &n.IPAMV4Info); err != nil {
			return err
		}
	}
	return nil
}
