package main

import (
	"io/ioutil"
	"os"
	"runtime/debug"

	"github.com/anon2anon/esterpad/internal/http"
	"github.com/anon2anon/esterpad/internal/mongo"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Log struct {
		Level     int
		Directory string
	}
	Mongo mongo.Config
	HTTP  http.Config
}

func getConfig(fname string) *Config {
	var conf Config
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.WithError(err).Fatal("config read")
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.WithError(err).Fatal("config unmarshal")
	}
	return &conf
}

func main() {
	if len(os.Args) < 2 {
		log.Infof("usage: %v [config file]", os.Args[0])
		os.Exit(1)
	}
	log.AddHook(filename.NewHook())
	config := getConfig(os.Args[1])
	log.SetLevel(log.Level(config.Log.Level))
	log.Debug("running with config: ", *config)
	defer func() {
		if err := recover(); err != nil {
			log.WithField("err", err).Fatal("Panic\n", string(debug.Stack()))
		}
	}()
	storage := mongo.New(config.Mongo)
	log.Debug(storage.LoginUser("test", "test"))
	// cacher.Init()
	err := http.Serve(config.HTTP)
	if err != nil {
		log.WithError(err).Fatal("Cannot start HTTP server")
	}
}
