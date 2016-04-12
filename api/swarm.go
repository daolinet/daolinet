package api

import (
        "net/http"
        "net/url"
)

func (a *Api) swarmRedirect(w http.ResponseWriter, req *http.Request) {
    var err error
    req.URL, err = url.ParseRequestURI(a.dUrl)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    a.fwd.ServeHTTP(w, req)
}
