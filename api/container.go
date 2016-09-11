package api

import (
    "fmt"
    "net/http"

    log "github.com/Sirupsen/logrus"
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
