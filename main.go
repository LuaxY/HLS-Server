package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/grafov/m3u8"
    "bufio"
    "os"
)

func main() {
    fmt.Println("HLS Server")

    router := mux.NewRouter()
    router.HandleFunc("/movies/index.m3u8", serverVersion).Methods("GET")
    router.PathPrefix("/").Handler(http.FileServer(http.Dir("./movies/"))) // TODO: use sendfile from nginx
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
}

func serverVersion(w http.ResponseWriter, r *http.Request) {
    movie  := openPlaylist("./movies/film1/index.m3u8")
    advert := openPlaylist("./movies/advert/index.m3u8")

    size := movie.Count() + advert.Count()
    playlist, err := m3u8.NewMediaPlaylist(size, size)

    if err != nil {
        panic(err)
    }

    playlist.MediaType = m3u8.VOD

    addPlaylist(playlist, advert, true)
    addPlaylist(playlist, movie, false)

    playlist.Close()

    w.Write(playlist.Encode().Bytes())
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