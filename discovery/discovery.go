package discovery

import (
	"errors"
	"time"

	//"github.com/docker/libkv/store"
)

var (
	// ErrNotSupported is returned when a discovery service is not supported.
	ErrNotSupported = errors.New("discovery service not supported")

	// ErrNotImplemented is returned when discovery feature is not implemented
	// by discovery backend.
	ErrNotImplemented = errors.New("not implemented in this discovery service")
)

// Backend is implemented by discovery backends which manage cluster entries.
type Backend interface {
	// Initialize the discovery with URIs, a heartbeat, a ttl and optional settings.
	Initialize(string, time.Duration, time.Duration, map[string]string) error

	// Register to the discovery
	Register(string, []byte) error
	Watch(string, <-chan struct{}) (<-chan [][]byte, <-chan error)
	Exists(string) (bool, error)
	PutTree(string) error
}
