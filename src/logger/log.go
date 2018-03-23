package logger

import (
    "fmt"
    "time"

    "HLS-Server/src/config"

    "github.com/rifflock/lfshook"
    "github.com/sirupsen/logrus"
    "gopkg.in/olivere/elastic.v5"
    "gopkg.in/sohlich/elogrus.v2"
)

var log = Init()
var cfg = config.Get()

func Init() *logrus.Logger {
    logger := logrus.New()

    logger.Hooks.Add(getElasticSearchHook())
    logger.Hooks.Add(getFileSystemHook())

    return logger
}

func getElasticSearchHook() logrus.Hook {
    client, err := elastic.NewClient(
        elastic.SetURL("http://"+cfg.ElasticSearch.Host+":"+cfg.ElasticSearch.Port),
        elastic.SetBasicAuth(cfg.ElasticSearch.User, cfg.ElasticSearch.Pass),
    )

    if err != nil {
        //logger.Panic(err)
        //log.Fatal(err)
        panic(err)
    }

    hook, err := elogrus.NewAsyncElasticHookWithFunc(
        client, cfg.ElasticSearch.Host, logrus.GetLevel(), func() string {
            t := time.Now()
            return fmt.Sprintf(t.Format(cfg.ElasticSearch.Index))
        })

    if err != nil {
        //logger.Panic(err)
        //log.Fatal(err)
        panic(err)
    }

    return hook
}

func getFileSystemHook() logrus.Hook {
    pathMap := lfshook.PathMap{
        logrus.InfoLevel:  "logs/info.log",
        logrus.ErrorLevel: "logs/error.log",
    }

    hook := lfshook.NewHook(
        pathMap,
        &logrus.JSONFormatter{},
    )

    return hook
}

func Get() *logrus.Logger {
    return log
}
