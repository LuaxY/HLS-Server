package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/grafov/m3u8"
    "bufio"
    "os"
    "time"
    "math/rand"
)

var cfg Config
var adverts []*m3u8.MediaPlaylist

func main() {
    fmt.Println("HLS Server")

    cfg = LoadConfiguration("config.json")
    rand.Seed(time.Now().Unix())

    //loadAdvert(cfg.MoviePath + "warning/index.m3u8")

    router := mux.NewRouter()
    router.HandleFunc("/{name:[A-Za-z0-9]+}/index.m3u8", streamPlaylist).Methods("GET")
    router.HandleFunc("/{name:[A-Za-z0-9]+}/s/{segment:[0-9]+.ts}", streamSegment).Methods("GET")

    log.Printf("Start listening on %s", cfg.Listen)

    log.Fatal(http.ListenAndServe(cfg.Listen, router))
}

func streamPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviePath + vars["name"] + "/index.m3u8"

    if cfg.VerboseLevel >= 1 {
        log.Print(file)
    }

    movie := openPlaylist(file)
    size := movie.Count()

    var advert *m3u8.MediaPlaylist

    if len(adverts) > 0 {
        advert = adverts[random(0, len(adverts))]

        if advert != nil {
            size += advert.Count()
        }
    }

    playlist, err := m3u8.NewMediaPlaylist(size, size)

    if err != nil {
        panic(err)
    }

    playlist.MediaType = m3u8.VOD
    isFirst := true

    if advert != nil {
        addPlaylist(playlist, advert, isFirst)
        isFirst = false
    }

    addPlaylist(playlist, movie, isFirst)

    playlist.Close()

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/x-mpegURL")

    w.Write(playlist.Encode().Bytes())
}

func streamSegment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviePath + vars["name"] + "/s/" + vars["segment"]

    if cfg.VerboseLevel >= 1 {
        log.Print(file)
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "video/MP2T")

    http.ServeFile(w, r, file)
}

func loadAdvert(file string) {
    playlist := openPlaylist(file)
    adverts = append(adverts, playlist)
}

func openPlaylist(file string) *m3u8.MediaPlaylist {
    f, err := os.Open(file)

    defer f.Close()

    if err != nil {
        panic(err)
    }

    p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), true)

    if err != nil {
        panic(err)
    }

    return p.(*m3u8.MediaPlaylist)
}

func addPlaylist(destination, playlist *m3u8.MediaPlaylist, isFirst bool) {
    destination.AppendSegment(playlist.Segments[0])

    if !isFirst {
        destination.SetDiscontinuity()
    }

    for _, segment := range playlist.Segments[1:playlist.Count()] {
        if segment != nil {
            destination.AppendSegment(segment)
        }
    }
}

func random(min, max int) int {
    return rand.Intn(max-min) + min
}
