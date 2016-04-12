package netutils

import (
	"errors"
	"net"
	"os"
	"os/exec"
)

type IP struct{}

func (i IP) run(args ...string) (string, error) {
	out, err := exec.Command("ip", args...).Output()
	return string(out), err
}

func (i IP) DeleteDevice(dev string) {
	_, err := os.Stat("/sys/class/net/" + dev)
	if err == nil {
		i.run("link", "delete", dev)
	}
}

func (i IP) SetDeviceUP(dev string) {
	i.run("link", "set", dev, "up")
}

func (i IP) GetAddress(dev string) (string, error) {
	iface, err := net.InterfaceByName(dev)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if net, ok := addr.(*net.IPNet); ok {
			if net.IP.To4() != nil {
				return addr.String(), nil
			}
		}
	}

	return "", errors.New("error to get address.")
}

func (i IP) SetAddress(dev, address string) error {
	_, err := i.run("addr", "replace", address, "dev", dev)
	return err
}
