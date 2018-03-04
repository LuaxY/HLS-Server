package main

import (
    "os"
    "fmt"
    "encoding/json"
)

type Config struct {
    Listen       string `json:"listen"`
    MoviePath    string `json:"moviePath"`
    VerboseLevel int    `json:"verboseLevel"`
}

func LoadConfiguration(file string) Config {
    var config Config
    configFile, err := os.Open(file)
    defer configFile.Close()
    if err != nil {
        fmt.Println(err.Error())
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}
