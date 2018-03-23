package middleware

import (
    "net/http"
    "time"

    "HLS-Server/src/logger"

    "github.com/sirupsen/logrus"
)

var log = logger.Get()

func Log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        next.ServeHTTP(w, r)

        log.WithFields(logrus.Fields{
            "remote_addr": r.RemoteAddr,
            "method":      r.Method,
            "request_uri": r.RequestURI,
            "referer":     r.Referer(),
            "user_agent":  r.UserAgent(),
            "duration":    time.Since(start),
        }).Info("")
    })
}
