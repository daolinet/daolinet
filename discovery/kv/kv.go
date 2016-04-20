package kv

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/daolinet/daolinet/discovery"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

// Discovery is exported
type Discovery struct {
	backend   store.Backend
	store     store.Store
	heartbeat time.Duration
	ttl       time.Duration
	prefix    string
	path      string
}

func init() {
	Init()
}

// Init is exported
func Init() {
	// Register to libkv
	zookeeper.Register()
	consul.Register()
	etcd.Register()

	// Register to internal discovery service
	discovery.Register("zk", &Discovery{backend: store.ZK})
	discovery.Register("consul", &Discovery{backend: store.CONSUL})
	discovery.Register("etcd", &Discovery{backend: store.ETCD})
}

// Initialize is exported
func (s *Discovery) Initialize(uris string, heartbeat time.Duration, ttl time.Duration, clusterOpts map[string]string) error {
	var (
		parts = strings.SplitN(uris, "/", 2)
		addrs = strings.Split(parts[0], ",")
		err   error
	)

	// A custom prefix to the path can be optionally used.
	if len(parts) == 2 {
		s.prefix = parts[1]
	}

	s.heartbeat = heartbeat
	s.ttl = ttl

	// Use a custom path if specified in discovery options
	dpath := clusterOpts["kv.path"]

	s.path = path.Join(s.prefix, dpath)

	var config *store.Config
	if clusterOpts["kv.cacertfile"] != "" && clusterOpts["kv.certfile"] != "" && clusterOpts["kv.keyfile"] != "" {
		log.Info("Initializing discovery with TLS")
		tlsConfig, err := Client(Options{
			CAFile:   clusterOpts["kv.cacertfile"],
			CertFile: clusterOpts["kv.certfile"],
			KeyFile:  clusterOpts["kv.keyfile"],
		})
		if err != nil {
			return err
		}
		config = &store.Config{
			// Set ClientTLS to trigger https (bug in libkv/etcd)
			ClientTLS: &store.ClientTLSConfig{
				CACertFile: clusterOpts["kv.cacertfile"],
				CertFile:   clusterOpts["kv.certfile"],
				KeyFile:    clusterOpts["kv.keyfile"],
			},
			// The actual TLS config that will be used
			TLS: tlsConfig,
		}
	} else {
		log.Info("Initializing discovery without TLS")
	}

	// Creates a new store, will ignore options given
	// if not supported by the chosen store
	s.store, err = libkv.NewStore(s.backend, addrs, config)
	s.InitKey()
	return err
}

// Register is exported
func (s *Discovery) Register(dpid string, gateway []byte) error {
	opts := &store.WriteOptions{TTL: s.ttl}
	return s.store.Put(path.Join(s.path, dpid), gateway, opts)
}

// Store returns the underlying store used by KV discovery
func (s *Discovery) Store() store.Store {
	return s.store
}

// Initialize key
func (s *Discovery) InitKey() error {
	opts := &store.WriteOptions{IsDir: true}
	return s.store.Put(s.path, nil, opts)
}

// Get the value at "key", returns the last modified
// index to use in conjunction to Atomic calls
func (s *Discovery) Get(key string) (pair *store.KVPair, err error) {
	return s.store.Get(key)
}

// Put a value at "key"
func (s *Discovery) Put(key string, value []byte, opts *store.WriteOptions) error {
	return s.store.Put(key, value, opts)
}

// Delete a value at "key"
func (s *Discovery) Delete(key string) error {
	return s.store.Delete(key)
}

// Exists checks if the key exists inside the store
func (s *Discovery) Exists(key string) (bool, error) {
	return s.store.Exists(key)
}

// List child nodes of a given directory
func (s *Discovery) List(directory string) ([]*store.KVPair, error) {
	return s.store.List(directory)
}

// Put a value at "key"
func (s *Discovery) PutTree(key string) error {
	opts := &store.WriteOptions{IsDir: true}
	return s.store.Put(key, nil, opts)
}

// DeleteTree deletes a range of keys under a given directory
func (s *Discovery) DeleteTree(directory string) error {
	return s.store.DeleteTree(directory)
}

// Watch is exported
func (s *Discovery) Watch(path string, stopCh <-chan struct{}) (<-chan [][]byte, <-chan error) {
	errCh := make(chan error)
	eventCh := make(chan [][]byte)

	go func() {
		defer close(errCh)
		defer close(eventCh)

		watchCh, err := s.store.WatchTree(path, stopCh)
		if err != nil {
			errCh <- err
		} else {
			for {
				select {
				case pairs, ok := <-watchCh:
					if !ok {
						errCh <- fmt.Errorf("Error to watch path (%s)", path)
						return
					}
					values := [][]byte{}
					for _, pair := range pairs {
						values = append(values, pair.Value)
					}
					eventCh <- values
				case <-stopCh:
					return
				}
			}

		}
	}()
	return eventCh, errCh
}
