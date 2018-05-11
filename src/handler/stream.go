package handler

import (
    "bufio"
    "fmt"
    "math/rand"
    "net/http"
    "os"

    "HLS-Server/src/config"
    "HLS-Server/src/errors"
    "HLS-Server/src/logger"

    "github.com/gorilla/mux"
    "github.com/grafov/m3u8"
    "github.com/sirupsen/logrus"
)

var log = logger.Get()
var cfg = config.Get()
var adverts []*m3u8.MediaPlaylist

//loadAdvert(cfg.MoviePath + "warning/index.m3u8")

func MasterPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    master := m3u8.NewMasterPlaylist()

    params480p := m3u8.VariantParams{
        Bandwidth:  1200000,
        Resolution: "854x480",
        Name:       "480p SD",
    }

    params720p := m3u8.VariantParams{
        Bandwidth:  6000000,
        Resolution: "1280x720",
        Name:       "720p HD",
    }

    master.Append(fmt.Sprintf("/movie/%s/%d/index.m3u8", vars["token"], 480), nil, params480p)
    master.Append(fmt.Sprintf("/movie/%s/%d/index.m3u8", vars["token"], 720), nil, params720p)

    w.Header().Set("Content-Type", "application/x-mpegURL")
    w.Write(master.Encode().Bytes())
}

func StreamPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviesPath + vars["id"] + "/" + vars["quality"] + "/index.m3u8"

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    movie, err := openPlaylist(file)

    if err != nil {
        panic(errors.Error{
            err,
            logrus.Fields{
                "id":    vars["id"],
                "token": vars["token"],
                "file":  file,
            },
        })
    }

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
        panic(errors.Error{
            err,
            logrus.Fields{
                "id":    vars["id"],
                "token": vars["token"],
                "file":  file,
                "count": movie.Count(),
            },
        })
    }

    playlist.MediaType = m3u8.VOD
    isFirst := true

    if advert != nil {
        err = addPlaylist(playlist, advert, isFirst, "???")

        if err != nil {
            panic(errors.Error{
                err,
                logrus.Fields{
                    "id":           vars["id"],
                    "token":        vars["token"],
                    "file":         file,
                    "count":        movie.Count(),
                    "count_advert": advert.Count(),
                },
            })
        }

        isFirst = false
    }

    err = addPlaylist(playlist, movie, isFirst, vars["token"])

    if err != nil {
        panic(errors.Error{
            err,
            logrus.Fields{
                "id":    vars["id"],
                "token": vars["token"],
                "file":  file,
                "count": movie.Count(),
            },
        })
    }

    playlist.Close()

    w.Header().Set("Content-Type", "application/x-mpegURL")
    w.Write(playlist.Encode().Bytes())
}

func StreamKey(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviesPath + vars["id"] + "/file.key"

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, r, file)
}

func StreamSegment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviesPath + vars["id"] + "/" + vars["quality"] + "/s/" + vars["segment"]

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "video/MP2T")
    http.ServeFile(w, r, file)
}

func loadAdvert(file string) {
    playlist, err := openPlaylist(file)

    if err != nil {
        log.Error(err)
        return
    }

    adverts = append(adverts, playlist)
}

func openPlaylist(file string) (*m3u8.MediaPlaylist, error) {
    f, err := os.Open(file)

    defer f.Close()

    if err != nil {
        return nil, err
    }

    p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), true)

    if err != nil {
        return nil, err
    }

    return p.(*m3u8.MediaPlaylist), nil
}

func addPlaylist(destination, playlist *m3u8.MediaPlaylist, isFirst bool, token string) error {
    var err error

    //re := regexp.MustCompile(`\/[a-zA-Z0-9]+\/s\/([0-9]+).ts`)
    //playlist.Segments[0].URI = re.ReplaceAllString(playlist.Segments[0].URI, "/"+token+"/s/$1.ts")

    if playlist.Key != nil {
        playlist.Segments[0].Key.URI = "/" + token + "/file.key"
    }

    playlist.Segments[0].URI = "/movie" + playlist.Segments[0].URI // FIXME

    err = destination.AppendSegment(playlist.Segments[0])

    if err != nil {
        return err
    }

    if !isFirst {
        err = destination.SetDiscontinuity()

        if err != nil {
            return err
        }
    }

    for _, segment := range playlist.Segments[1:playlist.Count()] {
        if segment != nil {
            //segment.URI = re.ReplaceAllString(segment.URI, "/"+token+"/s/$1.ts")
            segment.URI = "/movie" + segment.URI // FIXME
            err = destination.AppendSegment(segment)

            if err != nil {
                return err
            }
        }
    }
    return nil
}

func random(min, max int) int {
    return rand.Intn(max-min) + min
}
