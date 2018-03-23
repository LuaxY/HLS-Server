package handler

import (
    "bufio"
    "math/rand"
    "net/http"
    "os"
    "regexp"
    "strconv"

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

func StreamPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.MoviesPath + vars["id"] + "/index.m3u8"

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
                "count": strconv.FormatUint(uint64(movie.Count()), 10),
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
                    "count":        strconv.FormatUint(uint64(movie.Count()), 10),
                    "count_advert": strconv.FormatUint(uint64(advert.Count()), 10),
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
                "count": strconv.FormatUint(uint64(movie.Count()), 10),
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
    file := cfg.MoviesPath + vars["id"] + "/s/" + vars["segment"]

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

    re := regexp.MustCompile(`\/[a-zA-Z0-9]+\/s\/([0-9]+).ts`)
    playlist.Segments[0].URI = re.ReplaceAllString(playlist.Segments[0].URI, "/"+token+"/s/$1.ts")

    if playlist.Key != nil {
        playlist.Segments[0].Key.URI = "/" + token + "/file.key"
    }

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
            segment.URI = re.ReplaceAllString(segment.URI, "/"+token+"/s/$1.ts")
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
