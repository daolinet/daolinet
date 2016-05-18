package cli

import (
        "crypto/tls"
        "strings"
        "time"

        log "github.com/Sirupsen/logrus"
        "github.com/daolinet/daolinet/discovery"
        "github.com/daolinet/daolinet/discovery/kv"
        "github.com/samalba/dockerclient"
        "github.com/codegangsta/cli"
        "github.com/daolinet/daolinet/api"
)

func server(c *cli.Context) {
    listenAddr := c.String("listen")
    swarmUrl := c.String("swarm")
    allowInsecure := c.Bool("allow-insecure")

    ofcUrl := c.String("ofc")
    if ofcUrl == "" {
        log.Fatalf("The openflow controller url '%s' is invalid.", ofcUrl)
    }

    uri := getDiscovery(c)
    if uri == "" {
        log.Fatalf("discovery required to manage a cluster. See '%s server --help'.", c.App.Name)
    }
    //kv.Init()
    discovery := createDiscovery(uri, c, c.StringSlice("discovery-opt"))
    kvDiscovery, ok := discovery.(*kv.Discovery)
    if !ok {
        log.Fatal("Discovery service is only supported with consul, etcd and zookeeper discovery.")
    }

    var tlsConfig *tls.Config
    client, err := dockerclient.NewDockerClient(swarmUrl, tlsConfig)
    if err != nil {
        log.Fatal(err)
    }

    log.Debugf("connected to swarm: url=%s", swarmUrl)

    apiConfig := api.ApiConfig{
        ListenAddr: listenAddr,
        OfcUrl: ofcUrl, 
        Client: client,
        Store: kvDiscovery,
        AllowInsecure: allowInsecure,
    }

    daolinetApi, err := api.NewApi(apiConfig)
    if err != nil {
        log.Fatal(err)
    }

    if err := daolinetApi.Run(); err != nil {
        log.Fatal(err)
    }
}

// Initialize the discovery service.
func createDiscovery(uri string, c *cli.Context, discoveryOpt []string) discovery.Backend {
    hb, err := time.ParseDuration(c.String("heartbeat"))
    if err != nil {
        log.Fatalf("invalid --heartbeat: %v", err)
    }
    if hb < 1*time.Second {
        log.Fatal("--heartbeat should be at least one second")
    }

    // Set up discovery.
    discovery, err := discovery.New(uri, hb, 0, getDiscoveryOpt(c))
    if err != nil {
        log.Fatal(err)
    }

    return discovery
}

func getDiscoveryOpt(c *cli.Context) map[string]string {
    // Process the store options
    options := map[string]string{}
    for _, option := range c.StringSlice("discovery-opt") {
        if !strings.Contains(option, "=") {
            log.Fatal("--discovery-opt must contain key=value strings")
        }
        kvpair := strings.SplitN(option, "=", 2)
        options[kvpair[0]] = kvpair[1]
    }
    if _, ok := options["kv.path"]; !ok {
        options["kv.path"] = api.PathGateway
    }
    return options
}
