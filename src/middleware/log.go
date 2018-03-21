package middleware

import (
    "log"
    "net/http"
)

func Log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if cfg.Debug.VerbosityLevel >= 2 {
            log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.RequestURI)
        }

        next.ServeHTTP(w, r)
    })
}
