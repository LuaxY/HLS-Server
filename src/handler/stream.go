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

type Advert map[string]*m3u8.MediaPlaylist

var log = logger.Get()
var cfg = config.Get()
var adverts []Advert

func MasterPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    master := m3u8.NewMasterPlaylist()

    params720p := m3u8.VariantParams{
        Bandwidth:    6000000,
        Resolution:   "1280x720",
        Name:         "720p HD",
    }

    params480p := m3u8.VariantParams{
        Bandwidth:  1200000,
        Resolution: "854x480",
        Name:       "480p SD",
    }

    var subtitleFile string

    if vars["category"] == "tv" {
        subtitleFile = cfg.Path + "tvs/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/sub/subtitle.m3u8"
    } else {
        subtitleFile = cfg.Path + "movies/" + vars["id"] + "/sub/subtitle.m3u8"
    }

    if _, err := os.Stat(subtitleFile); err == nil {
        params720p.Alternatives = []*m3u8.Alternative{&m3u8.Alternative{
            Type:       "SUBTITLES",
            GroupId:    "subs",
            Name:       "Français",
            Language:   "fr",
            Default:    true,
            Forced:     "NO",
            Autoselect: "YES",
            URI:        fmt.Sprintf("/%s/%s/sub/subtitle.m3u8", vars["category"], vars["token"]),
        }}
    }

    master.Append(fmt.Sprintf("/%s/%s/%d/index.m3u8", vars["category"], vars["token"], 720), nil, params720p)
    master.Append(fmt.Sprintf("/%s/%s/%d/index.m3u8", vars["category"], vars["token"], 480), nil, params480p)

    w.Header().Set("Content-Type", "application/x-mpegURL")
    w.Write(master.Encode().Bytes())
}

func StreamPlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var file string

    if vars["category"] == "tv" {
        file = cfg.Path + "tvs/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/" + vars["quality"] + "/index.m3u8"
    } else {
        file = cfg.Path + "movies/" + vars["id"] + "/" + vars["quality"] + "/index.m3u8"
    }

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
        advert = adverts[random(0, len(adverts))][vars["quality"]]

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

func StreamSubtitlePlaylist(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var file string

    if vars["category"] == "tv" {
        file = cfg.Path + "tvs/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/sub/subtitle.m3u8"
    } else {
        file = cfg.Path + "movies/" + vars["id"] + "/sub/subtitle.m3u8"
    }

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "application/x-mpegURL")
    http.ServeFile(w, r, file)
}

func StreamKey(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var file string

    if vars["category"] == "tv" {
        file = cfg.Path + "movies/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/" + vars["quality"] + "/file.key"
    } else {
        file = cfg.Path + "tvs/" + vars["id"] + "/" + vars["quality"] + "/file.key"
    }

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, r, file)
}

func StreamMovieSegment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.Path + "movies/" + vars["id"] + "/" + vars["quality"] + "/s/" + vars["segment"]

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "video/MP2T")
    http.ServeFile(w, r, file)
}

func StreamTVSegment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.Path + "tvs/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/" + vars["quality"] + "/s/" + vars["segment"]

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "video/MP2T")
    http.ServeFile(w, r, file)
}

func StreamMovieSubtitle(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.Path + "movies/" + vars["id"] + "/sub/s/" + vars["segment"]

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "text/vtt")
    http.ServeFile(w, r, file)
}

func StreamTVSubtitle(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    file := cfg.Path + "tvs/" + vars["id"] + "/" + vars["season"] + "/" + vars["episode"] + "/sub/s/" + vars["segment"]

    if cfg.Debug.VerbosityLevel >= 1 {
        log.Debug(file)
    }

    w.Header().Set("Content-Type", "text/vtt")
    http.ServeFile(w, r, file)
}


func LoadAdvert(id string) {
    advert := make(Advert)

    for _, quality := range []string{"480", "720"} {
        file := cfg.Path + "movies/" + id + "/" + quality + "/index.m3u8"

        if cfg.Debug.VerbosityLevel >= 1 {
            log.Debug(file)
        }

        playlist, err := openPlaylist(file)

        if err != nil {
            panic(errors.Error{
                err,
                logrus.Fields{
                    "id":      id,
                    "quality": quality,
                    "file":    file,
                },
            })
        }

        advert[quality] = playlist
    }

    adverts = append(adverts, advert)
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
