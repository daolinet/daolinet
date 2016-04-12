package cli

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/daolicloud/daolinet/discovery"
	"github.com/daolicloud/daolinet/model"
	"github.com/daolicloud/daolinet/netutils"
)

const (
	DOCKERNETWORK = "docker/network/v1.0/network"
	DRIVERNETWORK = "daolinet"
)

func parseAddr(addr string) (string, string) {
	kpair := strings.SplitN(addr, ":", 2)
	if len(kpair) == 1 {
		return kpair[0], ""
	}

	return kpair[0], kpair[1]
}

func agent(c *cli.Context) {
	/*var (
		extdev string
		extip  string
	)*/
	dflag := getDiscovery(c)
	if dflag == "" {
		log.Fatalf("discovery required to connect a cluster. See '%s agent --help'.", c.App.Name)
	}

	ovs := netutils.NewOVS(c.String("bridge"))
	dpid, err := ovs.GetDatapath()
	if err != nil {
		log.Fatalf("error to get ovs datapath: %v", err)
	}

	/*intdev, intip := parseAddr(c.String("int-nic"))
	if intip == "" {
		log.Fatal("--int-nic should be of the form nic:ip")
	}

	extnic := c.String("ext-nic")
	if extnic == "" {
		extdev, extip = intdev, intip
	} else {
		extdev, extip = parseAddr(extnic)
		if extip == "" {
			log.Fatal("--ext-nic should be of the form nic:ip or empty")
		}
	}*/

	intdev, intip := parseAddr(c.String("iface"))
	if intip == "" {
		log.Fatal("--iface should be the form devname:ip")
	}

	extdev, extip := intdev, intip

	node := c.String("addr")
	if node == "" {
		node = intip
	}

	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("could not found hostname: %v", err)
	}

	gateway := model.NewGateway(node, host, dpid, intdev, intip, extdev, extip)
	value, err := json.Marshal(gateway)
	if err != nil {
		log.Fatalf("json marshal error: %v", err)
	}

	hb, err := time.ParseDuration(c.String("heartbeat"))
	if err != nil {
		log.Fatalf("invalid --heartbeat: %v", err)
	}
	if hb < 1*time.Second {
		log.Fatal("--heartbeat should be at least one second")
	}
	ttl, err := time.ParseDuration(c.String("ttl"))
	if err != nil {
		log.Fatalf("invalid --ttl: %v", err)
	}
	if ttl <= hb {
		log.Fatal("--ttl must be strictly superior to the heartbeat value")
	}

	//kv.Init()
	ttl = 0
	d, err := discovery.New(dflag, hb, ttl, getDiscoveryOpt(c))
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{"addr": extip, "discovery": dflag}).Infof("Registering on the discovery service every %s...", hb)
	if err := d.Register(dpid, value); err != nil {
		log.Error(err)
	}
	//time.Sleep(hb)

	stopCh := make(chan struct{})
	defer func() {
		stopCh <- struct{}{}
		close(stopCh)
	}()

	for {
		exists, err := d.Exists(DOCKERNETWORK)
		if err != nil {
			log.Fatalf("error trying to get value: %v", err)
		}
		if !exists {
			if err := d.PutTree(DOCKERNETWORK); err != nil {
				log.Fatalf("error trying to put value: %v", err)
			}
		}

		eventCh, errCh := d.Watch(DOCKERNETWORK, stopCh)
	Loop:
		for {
			select {
			case pairs := <-eventCh:
				if err := monitorGateway(ovs, pairs); err != nil {
					log.Error(err)
				}
			case err := <-errCh:
				if err != nil {
					log.Errorf("error chan: %v", err)
				}
				break Loop
			}
		}
		log.Warn("Watch to disconnected, retrying again.")
		time.Sleep(hb / 2)
	}
}

func monitorGateway(ovs *netutils.OVS, pairs [][]byte) error {
	var devMap = make(map[string]string)
	for _, pair := range pairs {
		network := model.Network{}
		if err := network.UnmarshalJSON(pair); err != nil {
			continue
		}
		if network.NetworkType == DRIVERNETWORK {
			ipamInfo := network.IPAMV4Info
			if len(ipamInfo) != 1 {
				log.Error("daolinet driver supported only one subnet")
				continue
			}
			devname := netutils.DeviceByNetwork(network.Id)
			devMap[devname] = ipamInfo[0].Gateway.String()
		}
	}

	var ovsMap = make(map[string]string)
	ip := netutils.IP{}
	iptable := netutils.IPtable{}

	out, err := ovs.FindInternal()
	if err != nil {
		return err
	}

	data := map[string][]interface{}{}
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		return err
	}

	for _, names := range data["data"] {
		name := names.([]interface{})
		dev := name[0].(string)
		if dev != DRIVERNETWORK {
			ovsMap[dev] = dev
			if _, ok := devMap[dev]; !ok {
				addr, err := ip.GetAddress(dev)
				ovs.DeleteNetwork(dev)
				if err != nil {
					log.Error(err)
					break
				}
				iptable.DropRule(addr)
			}
		}
	}

	for key, val := range devMap {
		if _, ok := ovsMap[key]; !ok {
			if err := ovs.CreateNetwork(key); err != nil {
				return err
			}
			ip.SetDeviceUP(key)
			if err := ip.SetAddress(key, val); err != nil {
				return err
			}
			iptable.AddRule(val)
		}
	}
	return nil
}
