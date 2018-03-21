package middleware

import (
    "net/http"
    "strings"

    "HLS-Server/src/config"
)

var cfg = config.Get()

func Secure(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.EqualFold(r.Host, cfg.Host) {
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte("403 Forbidden"))
            return
        }

        w.Header().Set("Access-Control-Allow-Origin", "*") // TEMP, for dev
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains") // WARN: use 'preload' only when project is in stable production

        if cfg.TLS.HPKP != "" {
            w.Header().Set("Public-Key-Pins", "pin-sha256=\""+cfg.TLS.HPKP+"\"; max-age=31536000; includeSubDomains")
        }

        next.ServeHTTP(w, r)
    })
}
