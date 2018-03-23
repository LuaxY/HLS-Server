package middleware

import (
    "net/http"

    "HLS-Server/src/logger"

    "github.com/sirupsen/logrus"
)

var log = logger.Get()

func Log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.WithFields(logrus.Fields{
            "remote_addr": r.RemoteAddr,
            "method":      r.Method,
            "request_uri": r.RequestURI,
            "referer":     r.Referer(),
            "user_agent":  r.UserAgent(),
        }).Info("")

        //log.Infof("%s %s %s %s %s %s", r.RemoteAddr, r.Method, "000", r.RequestURI, r.Referer(), r.UserAgent())

        next.ServeHTTP(w, r)
    })
}
