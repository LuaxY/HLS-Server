package middleware

import (
    //"github.com/sirupsen/logrus"
    "net/http"

    "HLS-Server/src/logger"
)

var log = logger.Get()

func Log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        /*log.WithFields(logrus.Fields{
            "remote_addr": r.RemoteAddr,
            "method":      r.Method,
            "request_uri": r.RequestURI,
            "status":      "000",
            "referer":     r.Referer(),
            "user_agent":  r.UserAgent(),
        }).Info("")*/
        log.Infof("%s %s %s %s %s %s", r.RemoteAddr, r.Method, "000", r.RequestURI, r.Referer(), r.UserAgent())
        next.ServeHTTP(w, r)
    })
}
