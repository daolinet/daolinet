package netutils

import "os/exec"

type IPtable struct{}

func (i IPtable) runNat(action, addr string) (string, error) {
	out, err := exec.Command("iptables", "-t", "nat", action, "POSTROUTING",
		"-s", addr, "!", "-d", addr, "-j", "MASQUERADE").Output()
	return string(out), err
}

func (i IPtable) runForward(action, target, addr string) (string, error) {
	out, err := exec.Command("iptables", action, "FORWARD",
		target, addr, "-j", "ACCEPT").Output()
	return string(out), err
}

func (i IPtable) AddRule(addr string) {
	i.runNat("-A", addr)
	i.runForward("-I", "-s", addr)
	i.runForward("-I", "-d", addr)
}

func (i IPtable) DropRule(addr string) {
	i.runNat("-D", addr)
	i.runForward("-D", "-s", addr)
	i.runForward("-D", "-d", addr)
}
