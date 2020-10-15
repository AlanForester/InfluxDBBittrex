package main

import (
	"BitrixInflux/db"
	"BitrixInflux/libs"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

var (
	crawlerFactories = map[string]libs.CrawlerFactory{
		libs.Bittrex: libs.NewBittrex,
	}
	writerFactories = map[string]db.WriterFactory{
		"influxdb": db.NewInfluxStorage,
	}
)

type BasicWriter struct{}

func (b BasicWriter) Write(d interface{}) {
	fmt.Printf("%+v\n", d)
}

type Config struct {
	CrawlerCFGS []libs.CrawlerConfig `json:"crawlers"`
	WriterCFGS  []db.WriterConfig    `json:"writers"`
}

type NullWriter struct{}

func (n NullWriter) Write(d interface{}) {}

func getConfig(configFile *string) Config {
	if configFile == nil || *configFile == "" {
		panic("error reading config file")
	}
	s, err := os.Stat(*configFile)
	if err != nil || s.IsDir() {
		panic("error reading config file")
	}
	cfg := Config{}
	cfgBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		panic("error reading config file")
	}
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		panic("error unmarshaling config")
	}
	return cfg
}

func main() {
	configFile := flag.String("config", "config.json", "config file in json format")
	crawlerName := "bittrex"
	flag.Parse()
	mainCfg := getConfig(configFile)
	log.Debugf("working with config %+v", mainCfg)
	if cf, ok := crawlerFactories[crawlerName]; ok {
		for _, cfg := range mainCfg.CrawlerCFGS {
			if cfg.Name == crawlerName {
				var writers []libs.DataWriter
				log.Debugf("parsing writer configs: %+v", mainCfg.WriterCFGS)
				for _, v := range mainCfg.WriterCFGS {
					log.Debugf("searching factory config for %s", v.Name)
					if wrf, ok := writerFactories[v.Name]; ok {
						log.Debugf("adding new writer %s", v.Name)
						dataW, err := wrf(v.Params)
						if err != nil {
							log.Fatalf("error instantiating writer %s: %s", v.Name, err)
						}
						writers = append(writers, dataW)
					}
				}
				crawl, err := cf(writers, cfg.Pairs)
				if err != nil {
					log.Fatalf("error creating crawler with name %s: %s", cfg.Name, err)
				}
				crawl.Loop()
			}
		}
	}
}

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: time.RFC3339})
}
