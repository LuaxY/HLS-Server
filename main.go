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

var adverts []*m3u8.MediaPlaylist

func main() {
    fmt.Println("HLS Server")

    rand.Seed(time.Now().Unix())

    loadAdvert("./movies/advert1/index.m3u8")
    loadAdvert("./movies/advert2/index.m3u8")
    loadAdvert("./movies/advert3/index.m3u8")

    router := mux.NewRouter()
    router.HandleFunc("/{name:[A-Za-z0-9]+}/index.m3u8", streamPlaylist).Methods("GET")
    router.HandleFunc("/{name:[A-Za-z0-9]+}/s/{segment:[0-9]+.ts}", streamSegment).Methods("GET")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
}

func streamPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := fmt.Sprintf("./movies/%s/index.m3u8", vars["name"])
    //fmt.Println(file)
    movie  := openPlaylist(file)
    advert := adverts[random(0, len(adverts))]
    size := movie.Count() + advert.Count()
    playlist, err := m3u8.NewMediaPlaylist(size, size)

    if err != nil {
        panic(err)
    }

    playlist.MediaType = m3u8.VOD

    addPlaylist(playlist, advert, true)
    addPlaylist(playlist, movie, false)

    playlist.Close()

    w.Header().Set("Content-Type", "application/x-mpegURL")
    w.Write(playlist.Encode().Bytes())
}

func streamSegment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := fmt.Sprintf("./movies/%s/s/%s", vars["name"], vars["segment"])
    //fmt.Println(file)
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
    return rand.Intn(max - min) + min
}