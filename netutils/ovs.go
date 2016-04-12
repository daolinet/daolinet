package netutils

import (
	"fmt"
	"os/exec"
	"strings"
)

type OVS struct {
	br      string
	timeout int
}

func (o *OVS) run(args ...string) (string, error) {
	timeout := fmt.Sprintf("--timeout=%d", o.timeout)
	format := fmt.Sprintf("--format=%s", "json")
	cmd := append([]string{timeout, format, "--no-heading", "--"}, args...)
	out, err := exec.Command("ovs-vsctl", cmd...).Output()
	//out, err := exec.Command("ovs-vsctl", args...).Output()
	return string(out), err
}

func (o *OVS) CreateNetwork(dev string) error {
	_, err := o.run("--if-exists", "del-port", dev,
		"--", "add-port", o.br, dev,
		"--", "set", "Interface", dev,
		"type=internal")
	return err
}

func (o *OVS) DeleteNetwork(dev string) error {
	_, err := o.run("--if-exists", "del-port", o.br, dev)
	return err
}

func (o *OVS) GetDatapath() (string, error) {
	out, err := o.run("get", "bridge", o.br, "datapath_id")
	if err != nil {
		return out, err
	}
	out = strings.Trim(string(out), "\"\n")
	return out, nil
}

func (o *OVS) FindInternal() (string, error) {
	out, err := o.run("--columns=name", "find", "interface", "type=internal")
	return strings.TrimRight(out, "\n "), err
}

func NewOVS(br string) *OVS {
	return &OVS{
		br:      br,
		timeout: 5,
	}
}
