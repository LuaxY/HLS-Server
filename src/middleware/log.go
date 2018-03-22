package middleware

import (
    "net/http"

    "HLS-Server/src/logger"
)

var log = logger.Get()

func Log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Infof("%s %s %s", r.RemoteAddr, r.Method, r.RequestURI)
        next.ServeHTTP(w, r)
    })
}
