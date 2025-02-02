package main

import (
    "crypto/tls"
    "math/rand"
    "net/http"
    "time"

    "HLS-Server/src/config"
    "HLS-Server/src/handler"
    "HLS-Server/src/logger"
    "HLS-Server/src/middleware"

    "github.com/gorilla/mux"
    "golang.org/x/net/http2"
)

var log = logger.Get()
var cfg = config.Get()

func main() {
    rand.Seed(time.Now().Unix())

    handler.LoadAdvert("0") // Intro

    router := mux.NewRouter()
    router.HandleFunc("/{category:movie|tv}/{token:[A-Za-z0-9]+}/master.m3u8", handler.MasterPlaylist).Methods("GET")
    router.HandleFunc("/{category:movie|tv}/{token:[A-Za-z0-9]+}/{quality:[0-9]{3,4}}/index.m3u8", handler.StreamPlaylist).Methods("GET")
    router.HandleFunc("/{category:movie|tv}/{token:[A-Za-z0-9]+}/sub/subtitle.m3u8", handler.StreamSubtitlePlaylist).Methods("GET")
    router.HandleFunc("/{category:movie|tv}/{token:[A-Za-z0-9]+}/file.key", handler.StreamKey).Methods("GET")
    router.HandleFunc("/movie/{id:[0-9]{1,10}}/{quality:[0-9]{3,4}}/s/{segment:[0-9]{4,5}.ts}", handler.StreamMovieSegment).Methods("GET")
    router.HandleFunc("/movie/{id:[0-9]{1,10}}/sub/s/{segment:[0-9]{4,5}.vtt}", handler.StreamMovieSubtitle).Methods("GET")
    router.HandleFunc("/tv/{id:[0-9]{1,10}}/{season:[0-9]{1,10}}/{episode:[0-9]{1,10}}/{quality:[0-9]{3,4}}/s/{segment:[0-9]{4,5}.ts}", handler.StreamTVSegment).Methods("GET")
    router.HandleFunc("/tv/{id:[0-9]{1,10}}/{season:[0-9]{1,10}}/{episode:[0-9]{1,10}}/sub/s/{segment:[0-9]{4,5}.vtt}", handler.StreamTVSubtitle).Methods("GET")

    router.Use(middleware.PanicRecover)
    router.Use(middleware.Secure)
    router.Use(middleware.AES)
    router.Use(middleware.Log)

    tlsConf := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
        },
    }

    srv := &http.Server{
        Addr:         ":" + cfg.Ports.HTTPS,
        Handler:      router,
        TLSConfig:    tlsConf,
        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
    }

    http2.ConfigureServer(srv, &http2.Server{})

    if cfg.Ports.HTTP != "" {
        go http.ListenAndServe(":"+cfg.Ports.HTTP, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
        }))
    }

    log.Infof("Start listening on port HTTP %s and HTTPS %s", cfg.Ports.HTTP, cfg.Ports.HTTPS)
    log.Fatal(srv.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key))
}
