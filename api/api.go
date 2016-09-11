package api

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/daolinet/daolinet/discovery/kv"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/samalba/dockerclient"
)

type (
	Api struct {
		listenAddr    string
		client        *dockerclient.DockerClient
		store         *kv.Discovery
		allowInsecure bool
		dUrl          string
		ofcUrl        string
		fwd           *forward.Forwarder
	}

	ApiConfig struct {
		ListenAddr    string
		OfcUrl        string
		Client        *dockerclient.DockerClient
		Store         *kv.Discovery
		AllowInsecure bool
	}
)

func NewApi(config ApiConfig) (*Api, error) {
	return &Api{
		listenAddr:    config.ListenAddr,
		ofcUrl:        config.OfcUrl,
		client:        config.Client,
		store:         config.Store,
		allowInsecure: config.AllowInsecure,
	}, nil
}

func (a *Api) Run() error {
	globalMux := http.NewServeMux()

	// forwarder for swarm
	var err error
	a.fwd, err = forward.New()
	if err != nil {
		return err
	}

	u := a.client.URL
	// setup redirect target to swarm
	scheme := "http://"

	// check if TLS is enabled and configure if so
	if a.client.TLSConfig != nil {
		log.Debug("configuring ssl for swarm redirect")
		scheme = "https://"
		// setup custom roundtripper with TLS transport
		r := forward.RoundTripper(
			&http.Transport{
				TLSClientConfig: a.client.TLSConfig,
			})
		f, err := forward.New(r)
		if err != nil {
			return err
		}

		a.fwd = f
	}

	// init key-value path
	if err := a.initPath(); err != nil {
		return err
	}

	a.dUrl = fmt.Sprintf("%s%s", scheme, u.Host)

	log.Debugf("configured docker proxy target: %s", a.dUrl)

	swarmRedirect := http.HandlerFunc(a.swarmRedirect)

	swarmHijack := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.swarmHijack(a.client.TLSConfig, a.dUrl, w, req)
	})

	mh := map[string]map[string]http.HandlerFunc{
		"GET": {
			"/api/gateways":                a.gateways,
			"/api/gateways/{id}":           a.gateway,
			"/api/groups":                  a.groups,
			"/api/groups/{name}":           a.group,
			"/api/policy":                  a.policys,
			"/api/policy/{peer}":           a.policy,
			"/api/firewalls":               a.firewalls,
			"/api/firewalls/{name}":        a.firewallByContainer,
			"/api/firewalls/{node}/{port}": a.firewall,
		},
		"POST": {
			"/api/groups":        a.saveGroup,
			"/api/groups/{name}": a.saveMember,
			"/api/policy/{peer}": a.savePolicy,
			"/api/firewalls":     a.saveFirewall,
		},
		"DELETE": {
			"/api/groups/{name}":          a.deleteGroup,
			"/api/groups/{name}/{member}": a.deleteMember,
			"/api/policy/{peer}":          a.deletePolicy,
			"/api/firewalls/{name}":       a.deleteFirewall,
		},
                "PUT": {
			"/api/containers/{id}/reset":    a.resetContainer,
                },
	}

	apiRouter := mux.NewRouter()
	for method, routes := range mh {
		for route, fct := range routes {
			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				localFct(w, r)
			}
			localMethod := method

			// add the new route
			apiRouter.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			apiRouter.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	// global handler
	globalMux.Handle("/", http.FileServer(http.Dir("static")))
	globalMux.Handle("/api/", apiRouter)
	//globalMux.Handle("/v{version:[0-9.]+}/api/", apiRouter)
	globalMux.Handle("/v1.14/api/", apiRouter)
	globalMux.Handle("/v1.15/api/", apiRouter)
	globalMux.Handle("/v1.16/api/", apiRouter)
	globalMux.Handle("/v1.17/api/", apiRouter)
	globalMux.Handle("/v1.18/api/", apiRouter)
	globalMux.Handle("/v1.19/api/", apiRouter)
	globalMux.Handle("/v1.20/api/", apiRouter)
	globalMux.Handle("/v1.21/api/", apiRouter)
	globalMux.Handle("/v1.22/api/", apiRouter)
	globalMux.Handle("/v1.23/api/", apiRouter)
	globalMux.Handle("/v1.24/api/", apiRouter)
	globalMux.Handle("/v1.25/api/", apiRouter)
	globalMux.Handle("/v1.26/api/", apiRouter)

	// swarm
	swarmRouter := mux.NewRouter()
	// these are pulled from the swarm api code to proxy and allow
	// usage with the standard Docker cli
	m := map[string]map[string]http.HandlerFunc{
		"GET": {
			"/_ping":                          swarmRedirect,
			"/events":                         swarmRedirect,
			"/info":                           swarmRedirect,
			"/version":                        swarmRedirect,
			"/images/json":                    swarmRedirect,
			"/images/viz":                     swarmRedirect,
			"/images/search":                  swarmRedirect,
			"/images/get":                     swarmRedirect,
			"/images/{name:.*}/get":           swarmRedirect,
			"/images/{name:.*}/history":       swarmRedirect,
			"/images/{name:.*}/json":          swarmRedirect,
			"/containers/ps":                  swarmRedirect,
			"/containers/json":                swarmRedirect,
			"/containers/{name:.*}/export":    swarmRedirect,
			"/containers/{name:.*}/changes":   swarmRedirect,
			"/containers/{name:.*}/json":      swarmRedirect,
			"/containers/{name:.*}/top":       swarmRedirect,
			"/containers/{name:.*}/logs":      swarmRedirect,
			"/containers/{name:.*}/stats":     swarmRedirect,
			"/containers/{name:.*}/attach/ws": swarmHijack,
			"/exec/{execid:.*}/json":          swarmRedirect,
			"/networks":                       swarmRedirect,
			"/networks/{networkid:.*}":        swarmRedirect,
			"/volumes":                        swarmRedirect,
			"/volumes/{volumename:.*}":        swarmRedirect,
		},
		"POST": {
			"/auth":                               swarmRedirect,
			"/commit":                             swarmRedirect,
			"/build":                              swarmRedirect,
			"/images/create":                      swarmRedirect,
			"/images/load":                        swarmRedirect,
			"/images/{name:.*}/push":              swarmRedirect,
			"/images/{name:.*}/tag":               swarmRedirect,
			"/containers/create":                  swarmRedirect,
			"/containers/{name:.*}/kill":          swarmRedirect,
			"/containers/{name:.*}/pause":         swarmRedirect,
			"/containers/{name:.*}/unpause":       swarmRedirect,
			"/containers/{name:.*}/rename":        swarmRedirect,
			"/containers/{name:.*}/restart":       swarmRedirect,
			"/containers/{name:.*}/start":         swarmRedirect,
			"/containers/{name:.*}/stop":          swarmRedirect,
			"/containers/{name:.*}/wait":          swarmRedirect,
			"/containers/{name:.*}/resize":        swarmRedirect,
			"/containers/{name:.*}/attach":        swarmHijack,
			"/containers/{name:.*}/copy":          swarmRedirect,
			"/containers/{name:.*}/exec":          swarmRedirect,
			"/exec/{execid:.*}/start":             swarmHijack,
			"/exec/{execid:.*}/resize":            swarmRedirect,
			"/networks/create":                    swarmRedirect,
			"/networks/{networkid:.*}/connect":    swarmRedirect,
			"/networks/{networkid:.*}/disconnect": swarmRedirect,
			"/volumes/create":                     swarmRedirect,
		},
		"PUT": {
			"/containers/{name:.*}/archive": swarmRedirect,
		},
		"DELETE": {
			"/containers/{name:.*}":    swarmRedirect,
			"/images/{name:.*}":        swarmRedirect,
			"/networks/{networkid:.*}": swarmRedirect,
			"/volumes/{name:.*}":       swarmRedirect,
		},
		"OPTIONS": {
			"": swarmRedirect,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				localFct(w, r)
			}
			localMethod := method

			// add the new route
			swarmRouter.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			swarmRouter.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	globalMux.Handle("/containers/", swarmRouter)
	globalMux.Handle("/_ping", swarmRouter)
	globalMux.Handle("/commit", swarmRouter)
	globalMux.Handle("/build", swarmRouter)
	globalMux.Handle("/events", swarmRouter)
	globalMux.Handle("/version", swarmRouter)
	globalMux.Handle("/images/", swarmRouter)
	globalMux.Handle("/exec/", swarmRouter)
	globalMux.Handle("/v1.14/", swarmRouter)
	globalMux.Handle("/v1.15/", swarmRouter)
	globalMux.Handle("/v1.16/", swarmRouter)
	globalMux.Handle("/v1.17/", swarmRouter)
	globalMux.Handle("/v1.18/", swarmRouter)
	globalMux.Handle("/v1.19/", swarmRouter)
	globalMux.Handle("/v1.20/", swarmRouter)
	globalMux.Handle("/v1.21/", swarmRouter)
	globalMux.Handle("/v1.22/", swarmRouter)
	globalMux.Handle("/v1.23/", swarmRouter)
	globalMux.Handle("/v1.24/", swarmRouter)
	globalMux.Handle("/v1.25/", swarmRouter)
	globalMux.Handle("/v1.26/", swarmRouter)

	log.Infof("controller listening on %s", a.listenAddr)

	s := &http.Server{
		Addr:    a.listenAddr,
		Handler: context.ClearHandler(globalMux),
	}

	var runErr error
	runErr = s.ListenAndServe()
	return runErr
}
