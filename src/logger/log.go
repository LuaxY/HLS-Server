package logger

import (
    "fmt"
    "time"

    "HLS-Server/src/config"

    "github.com/sirupsen/logrus"
    "gopkg.in/olivere/elastic.v5"
    "gopkg.in/sohlich/elogrus.v2"
)

var log = Init()
var cfg = config.Get()

func Init() *logrus.Logger {
    logger := logrus.New()

    client, err := elastic.NewClient(
        elastic.SetURL("http://"+cfg.ElasticSearch.Host+":"+cfg.ElasticSearch.Port),
        elastic.SetBasicAuth(cfg.ElasticSearch.User, cfg.ElasticSearch.Pass),
    )

    if err != nil {
        logger.Panic(err)
    }

    hook, err := elogrus.NewAsyncElasticHookWithFunc(
        client, cfg.ElasticSearch.Host, logrus.GetLevel(), func() string {
            t := time.Now()
            return fmt.Sprintf(t.Format(cfg.ElasticSearch.Index))
        })

    if err != nil {
        logger.Panic(err)
    }

    logger.Hooks.Add(hook)
    return logger
}

func Get() *logrus.Logger {
    return log
}
