package api

import (
	"crypto/tls"
	"net/http"
)

func newClientAndScheme(tlsConfig *tls.Config) *http.Client {
        if tlsConfig != nil {
                return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
        }
        return &http.Client{}
}

// prevents leak with https
func closeIdleConnections(client *http.Client) {
    if tr, ok := client.Transport.(*http.Transport); ok {
        tr.CloseIdleConnections()
    }
}
