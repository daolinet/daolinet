package model

type (
	Gateway struct {
		Node       string
		HostName   string
		DatapathID string
		IntDev     string
		IntIP      string
		ExtDev     string
		ExtIP      string
	}

	Firewall struct {
		Name        string
		Container   string
		DatapathID  string
		GatewayIP   string
		GatewayPort int
		ServicePort int
	}
)

func NewGateway(node, hostname, datapath, intdev, intip, extdev, extip string) *Gateway {
	return &Gateway{
		Node:       node,
		HostName:   hostname,
		DatapathID: datapath,
		IntDev:     intdev,
		IntIP:      intip,
		ExtDev:     extdev,
		ExtIP:      extip,
	}
}
