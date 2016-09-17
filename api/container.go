package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "path"
    "strconv"
    "strings"

    log "github.com/Sirupsen/logrus"
    "github.com/daolinet/daolinet/model"
    "github.com/gorilla/mux"
    "github.com/samalba/dockerclient"
)


func (a *Api) resetContainer(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }

    oldId := mux.Vars(r)["id"]
    info, err := a.client.InspectContainer(oldId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if info.State.Running {
        err := a.client.StopContainer(info.Id, 5)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    err = a.client.RenameContainer(info.Id, info.Name + "old")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    config := &dockerclient.ContainerConfig{}
    config.Tty = info.Config.Tty
    config.Env = info.Config.Env
    config.Cmd = info.Config.Cmd
    config.Image = info.Config.Image

    netMode := info.HostConfig.NetworkMode
    net := info.NetworkSettings.Networks[netMode]
    if net != nil {
        config.MacAddress = net.MacAddress
        config.HostConfig = dockerclient.HostConfig{
            NetworkMode: netMode,
        }
        endpointsConfig := map[string]*dockerclient.EndpointSettings {
            netMode: &dockerclient.EndpointSettings{
                IPAMConfig: &dockerclient.EndpointIPAMConfig {
                   IPv4Address: net.IPAddress,
                },
                //IPAddress: net.IPAddress,
                //MacAddress: net.MacAddress,
            },
        }

        config.NetworkingConfig = dockerclient.NetworkingConfig{
            EndpointsConfig: endpointsConfig,
        }
    }

    // data, _ := json.Marshal(config)
    // log.Warn(string(data))
    hostConfig := &dockerclient.HostConfig{
        NetworkMode: netMode,
    }
    newId, err := a.client.CreateContainer(config, info.Name, nil)
    if err != nil {
        a.client.StartContainer(info.Id, hostConfig)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    err = a.client.RemoveContainer(info.Id, true, false)
    if err != nil {
        log.Warnf("Remove container: %v", err)
    }

    go func() {
        client := newClientAndScheme(a.client.TLSConfig)
        url := fmt.Sprintf("%s/v1/containers/%s", a.ofcUrl, info.Id)
        resp, err := client.Post(url, "application/json", nil)
        if err != nil {
            log.Warnf("Remove container from openflow controller: %v", err)
        }

	if resp != nil && resp.Body != nil {
            resp.Body.Close()
        }
    }()

    err = a.client.StartContainer(newId, hostConfig)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write([]byte(newId))
}

func (a *Api) resetContainerById(oldId , newId string) {
    // Firewall, Group, Policy
    firewalls, err := a.store.List(pathNameFirewall)
    if err != nil {
        log.Errorf("error to get all firewalls: %v", err)
    } else {
        for _, fw := range firewalls {
            var firewall model.Firewall
            err := json.Unmarshal(fw.Value, &firewall)
            if err != nil {
                log.Errorf("error unmarshal firewall: %v", err)
                continue
            }
            if firewall.Container == oldId {
                firewall.Container = newId
                value, err := json.Marshal(firewall)
                if err != nil {
                    log.Errorf("json marshal error: %v", err)
                    continue
                }
                if err := a.store.Put(fw.Key, value, nil); err != nil {
                    log.Errorf("error to update container firewall by name: %v", err)
                    continue
                }
                nodeurl := path.Join(pathNodeFirewall, firewall.DatapathID, strconv.Itoa(firewall.GatewayPort))
                if err := a.store.Put(nodeurl, value, nil); err != nil {
                    log.Errorf("error to update container firewall by node: %v", err)
                    continue
                }
            }
        }
    }

    policies, err := a.store.List(pathPolicy)
    if err != nil {
        log.Errorf("error to get all policies : %v", err)
    } else {
        for _, policy := range policies {
            peer := strings.Split(policy.Key, "/")
            parts := strings.Split(peer[len(peer)-1], ":")
            if len(parts) != 2 {
                log.Error(ErrPolicyFormat.Error())
                continue
            }
            if oldId == parts[0] || oldId == parts[1] {
                if err := a.store.Delete(policy.Key); err != nil {
                    log.Warnf("error deleting policy: %v", err)
                }
                if oldId == parts[0] {
                    parts[0] = newId
                } else {
                    parts[1] = newId
                }

                var key string
                if strings.Compare(parts[0], parts[1]) > 0 {
                    key = fmt.Sprintf("%s:%s", parts[1], parts[0])
                } else {
                    key = fmt.Sprintf("%s:%s", parts[0], parts[1])
                }

                if err := a.store.Put(path.Join(pathPolicy, key), policy.Value, nil); err != nil {
                    log.Errorf("error to save policy: %v", err)
                }
            }
        }
    }
}
