package netutils

const NETPREFIX = "tap"

func DeviceByNetwork(nid string) string {
	return NETPREFIX + nid[:11]
}
