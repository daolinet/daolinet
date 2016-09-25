package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/daolinet/daolinet/model"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

const (
	CONNECTED    = "ACCEPT"
	DISCONNECTED = "DROP"
)

const (
	PathGateway      = "daolinet/gateways"
	pathGroup        = "daolinet/groups"
	pathPolicy       = "daolinet/policy"
	pathNodeFirewall = "daolinet/firewalls/node"
	pathNameFirewall = "daolinet/firewalls/name"
)

var (
	ErrGroupExists         = errors.New("group already exists")
	ErrFirewallNameExists  = errors.New("firewall name already exists")
	ErrFirewallPortExists  = errors.New("firewall gateway port already exists")
	ErrGroupDoesNotExist   = errors.New("group does not exist")
	ErrPolicyDoesNotExist  = errors.New("policy does not exist")
	ErrGatewayDoesNotExist = errors.New("gateway does not exist")
	ErrPolicyConflict      = errors.New("policy should not be same container")
	ErrPolicyFormat        = errors.New("policy format should be <CONTAINER:CONTAINER>")
)

type ContainerInfo struct {
	dockerclient.ContainerInfo
	Node struct {
		ID   string
		IP   string
		Addr string
		Name string
	}
}

func (a *Api) initPath() error {
	var paths = [...]string{pathGroup, pathPolicy, pathNodeFirewall, pathNameFirewall}
	for _, p := range paths {
		exists, _ := a.store.Exists(p)
		if !exists {
			if err := a.store.PutTree(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Api) choiceGateway(node string) (*model.Gateway, error) {
	var (
		ok          bool = false
		nodeGateway model.Gateway
		tmpGateways []model.Gateway
	)
	gateway, err := a.store.List(PathGateway)
	if err != nil {
		return nil, err
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	for _, gw := range gateway {
		var g model.Gateway
		if err := json.Unmarshal(gw.Value, &g); err != nil {
			log.Errorf("error unmarshal gateway: %v", err)
			continue
		}
		if node == g.Node || (node == "" && g.HostName == host) {
			ok = true
			nodeGateway = g
		}
		if g.IntDev != g.ExtDev || g.IntIP != g.ExtIP {
			tmpGateways = append(tmpGateways, g)
		}
	}
	if len(tmpGateways) > 0 {
		return &tmpGateways[0], nil
	}

	if !ok {
		return nil, ErrGatewayDoesNotExist
	} else {
		return &nodeGateway, nil
	}
}

func (a *Api) gateways(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	gateway, err := a.store.List(PathGateway)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gateways := []model.Gateway{}
	tmp_gateways := []model.Gateway{}
	for _, gw := range gateway {
		var g model.Gateway
		err := json.Unmarshal(gw.Value, &g)
		if err != nil {
			log.Errorf("error unmarshal gateway: %v", err)
			continue
		}
		if g.IntDev != g.ExtDev || g.IntIP != g.ExtIP {
			tmp_gateways = append(tmp_gateways, g)
		}
		gateways = append(gateways, g)
	}
	if len(tmp_gateways) > 0 {
		gateways = tmp_gateways
	}
	if err := json.NewEncoder(w).Encode(gateways); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) gateway(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	gateway, err := a.store.Get(path.Join(PathGateway, vars["id"]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(gateway.Value)
}

func (a *Api) groups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	groups, err := a.store.List(pathGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupArray := []string{}
	for _, group := range groups {
		parts := strings.Split(group.Key, "/")
		groupArray = append(groupArray, parts[len(parts)-1])
	}

	if err := json.NewEncoder(w).Encode(groupArray); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) saveGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var group = map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name, ok := group["name"]
	if !ok {
		http.Error(w, "name cannot be empty.", http.StatusInternalServerError)
		return
	}

	key := path.Join(pathGroup, name)
	exists, err := a.store.Exists(key)
	if exists {
		http.Error(w, ErrGroupExists.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.store.PutTree(key); err != nil {
		log.Errorf("error saving group: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) group(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	members, err := a.store.List(path.Join(pathGroup, name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	memberArray := []string{}
	for _, member := range members {
		parts := strings.Split(member.Key, "/")
		memberArray = append(memberArray, parts[len(parts)-1])
	}

	if err := json.NewEncoder(w).Encode(memberArray); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) deleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := a.store.DeleteTree(path.Join(pathGroup, vars["name"])); err != nil {
		log.Errorf("error deleting group: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) saveMember(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var m = map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	member, ok := m["member"]
	if !ok {
		http.Error(w, "member cannot be empty.", http.StatusInternalServerError)
		return
	}

	groupath := path.Join(pathGroup, mux.Vars(r)["name"])
	exists, err := a.store.Exists(groupath)
	if !exists {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Error(w, ErrGroupDoesNotExist.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := a.store.PutTree(path.Join(groupath, member)); err != nil {
		log.Errorf("error saving member: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) deleteMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := path.Join(pathGroup, vars["name"], vars["member"])
	if err := a.store.DeleteTree(key); err != nil {
		log.Errorf("error deleting member: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) parsePolicy(parts []string) (*dockerclient.ContainerInfo, *dockerclient.ContainerInfo, error) {
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, nil, ErrPolicyFormat
	}

	pInfo, err := a.client.InspectContainer(parts[0])
	if err != nil {
		return nil, nil, err
	}

	qInfo, err := a.client.InspectContainer(parts[1])
	if err != nil {
		return nil, nil, err
	}

	switch strings.Compare(pInfo.Id, qInfo.Id) {
	case -1:
		pInfo, qInfo = pInfo, qInfo
	case +1:
		pInfo, qInfo = qInfo, pInfo
	default:
		return nil, nil, ErrPolicyConflict
	}

	return pInfo, qInfo, nil
}

func (a *Api) policys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	policies, err := a.store.List(pathPolicy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
        var data = map[string]string{}
	for _, policy := range policies {
                peer := strings.Split(policy.Key, "/")
                parts := strings.Split(peer[len(peer)-1], ":")
                pInfo, qInfo, err := a.parsePolicy(parts)
		if err != nil {
                    log.Errorf(err.Error())
                    continue
		}
                key := strings.Join([]string{
                        strings.TrimLeft(pInfo.Name, "/"),
                        strings.TrimLeft(qInfo.Name, "/")}, ":")
                data[key] = string(policy.Value)
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) policy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(mux.Vars(r)["peer"], ":")
	pInfo, qInfo, err := a.parsePolicy(parts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var val []byte
        key := fmt.Sprintf("%s:%s", pInfo.Id, qInfo.Id)
	// pair, err := a.store.Get(path.Join(pathPolicy, pInfo.Id, qInfo.Id))
	pair, err := a.store.Get(path.Join(pathPolicy, key))
	if err != nil {
		val = []byte("")
	} else {
		val = pair.Value
		if string(val) != CONNECTED && string(val) != DISCONNECTED {
			val = []byte("")
		}
	}

	w.Write(val)
}

func (a *Api) savePolicy(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	action := data["action"]
	if action != CONNECTED && action != DISCONNECTED {
		http.Error(w, "error to find action method", http.StatusInternalServerError)
		return
	}

	parts := strings.Split(mux.Vars(r)["peer"], ":")
	pInfo, qInfo, err := a.parsePolicy(parts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    if action == DISCONNECTED {
        data := map[string]string{
            "sid": pInfo.Id,
            "did": qInfo.Id,
        }

        value, err := json.Marshal(data)
        if err != nil {
	    	log.Fatalf("json marshal error: %v", err)
	    	http.Error(w, err.Error(), http.StatusInternalServerError)
	    	return
        }
        body := bytes.NewBuffer(value)
	    client := newClientAndScheme(a.client.TLSConfig)

	    resp, err := client.Post(a.ofcUrl + "/v1/policy", "application/json", body)
	    if err != nil {
	    	http.Error(w, err.Error(), http.StatusInternalServerError)
	    	return
	    }
        resp.Body.Close()

    }

	// if err := a.store.Put(path.Join(pathPolicy, pInfo.Id, qInfo.Id), []byte(action), nil); err != nil {
	key := fmt.Sprintf("%s:%s", pInfo.Id, qInfo.Id)
	if err := a.store.Put(path.Join(pathPolicy, key), []byte(action), nil); err != nil {
		//log.Errorf("error saving policy: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) deletePolicy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(mux.Vars(r)["peer"], ":")
	pInfo, qInfo, err := a.parsePolicy(parts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// if err := a.store.Delete(path.Join(pathPolicy, pInfo.Id, qInfo.Id)); err != nil {
	key := fmt.Sprintf("%s:%s", pInfo.Id, qInfo.Id)
	if err := a.store.Delete(path.Join(pathPolicy, key)); err != nil {
		//log.Errorf("error deleting policy: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) saveFirewall(w http.ResponseWriter, r *http.Request) {
	firewall := model.Firewall{}
	if err := json.NewDecoder(r.Body).Decode(&firewall); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := firewall.Name
	container := firewall.Container
	gatewayIP := firewall.GatewayIP

	if name == "" || container == "" {
		http.Error(w, "name or container cannot be empty.", http.StatusInternalServerError)
		return
	}

	nameurl := path.Join(pathNameFirewall, name)
	exists, err := a.store.Exists(nameurl)
	if exists {
		http.Error(w, ErrFirewallNameExists.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := newClientAndScheme(a.client.TLSConfig)
	resp, err := client.Get(a.dUrl + "/containers/" + container + "/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//cleanup
	defer resp.Body.Close()
	defer closeIdleConnections(client)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode >= 400 {
		http.Error(w, string(data), http.StatusInternalServerError)
		return
	}

	var info ContainerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

        if gatewayIP == "" {
            gatewayIP = info.Node.IP
        }
	gateway, err := a.choiceGateway(gatewayIP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	firewall.Container = info.Id
	firewall.DatapathID = gateway.DatapathID
	firewall.GatewayIP = gateway.ExtIP

	value, err := json.Marshal(firewall)
	if err != nil {
		log.Fatalf("json marshal error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodeurl := path.Join(pathNodeFirewall, gateway.DatapathID, strconv.Itoa(firewall.GatewayPort))
	exists, err = a.store.Exists(nodeurl)
	if exists {
		http.Error(w, ErrFirewallPortExists.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.store.Put(nodeurl, value, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.store.Put(nameurl, value, nil); err != nil {
		a.store.Delete(nodeurl)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(firewall); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) firewalls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	containers, err := a.client.ListContainers(true, true, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	containerMap := map[string]string{}
	for _, container := range containers {
		var name = ""
		if len(container.Names) > 0 {
			name = container.Names[0]
		}
		containerMap[container.Id] = name
	}

	firewalls, err := a.store.List(pathNameFirewall)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var values = []model.Firewall{}
	for _, firewall := range firewalls {
		var fw model.Firewall
		if err := json.Unmarshal(firewall.Value, &fw); err != nil {
			continue
		}
		fw.Container = containerMap[fw.Container]
		values = append(values, fw)
	}

	if err := json.NewEncoder(w).Encode(values); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) firewallByContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("content-type", "application/json")

	containerInfo, err := a.client.InspectContainer(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	firewalls, err := a.store.List(pathNameFirewall)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var values = []model.Firewall{}
	for _, firewall := range firewalls {
		var fw model.Firewall
		if err := json.Unmarshal(firewall.Value, &fw); err != nil {
			continue
		}
		if fw.Container == containerInfo.Id {
			values = append(values, fw)
		}
	}

	if err := json.NewEncoder(w).Encode(values); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) firewall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := path.Join(pathNodeFirewall, vars["node"], vars["port"])
	firewall, err := a.store.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(firewall.Value)
}

func (a *Api) deleteFirewall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	firewall, err := a.store.Get(path.Join(pathNameFirewall, vars["name"]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fw model.Firewall
	if err := json.Unmarshal(firewall.Value, &fw); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodeurl := path.Join(pathNodeFirewall, fw.DatapathID)
	if err := a.store.Delete(path.Join(nodeurl, strconv.Itoa(fw.GatewayPort))); err != nil {
		log.Warnf("deleting firewall %s: %v", vars["name"], err)
	}

	nameurl := path.Join(pathNameFirewall, fw.Name)
	if err := a.store.Delete(nameurl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodes, err := a.store.List(nodeurl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		if len(nodes) == 0 {
			if err := a.store.DeleteTree(nodeurl); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}

	w.Header().Set("content-type", "application/json")
	w.Write(firewall.Value)
}
