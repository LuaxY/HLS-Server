package config

import (
    "encoding/json"
    "flag"
    "log"
    "os"
)

var cfg = LoadConfiguration()

type Config struct {
    Ports struct {
        HTTP  string `json:"http"`
        HTTPS string `json:"https"`
    } `json:"ports"`

    TLS struct {
        Cert string `json:"cert"`
        Key  string `json:"key"`
        HPKP string `json:"hpkp"`
    } `json:"tls"`

    ElasticSearch struct {
        Host  string `json:"host"`
        Port  string `json:"port"`
        User  string `json:"user"`
        Pass  string `json:"pass"`
        Index string `json:"index"`
    } `json:"elasticsearch"`

    Host       string `json:"host"`
    AES        string `json:"aes"`
    Path       string `json:"path"`

    Debug struct {
        VerbosityLevel int `json:"verbosityLevel"`
    } `json:"debug"`
}

func LoadConfiguration() Config {
    var config Config

    file := flag.String("c", "dev.config.json", "config file")
    flag.Parse()

    log.Printf("Read config file: %s", *file)

    configFile, err := os.Open(*file)
    defer configFile.Close()

    if err != nil {
        log.Print(err)
    }

    jsonParser := json.NewDecoder(configFile)
    err = jsonParser.Decode(&config)

    if err != nil {
        log.Print(err)
    }

    return config
}

func Get() *Config {
    return &cfg
}
