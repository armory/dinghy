package main

import (
	"net/http"
	"os"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/armory-io/dinghy/pkg/web"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	logLevelStr := util.GetenvOrDefault("DEBUG_LEVEL", "info")
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.Panic("Invalid log level : " + logLevelStr)
	}
	log.SetLevel(logLevel)
	log.Info("Dinghy started.")
	cache.C = cache.NewRedisCacheStore()
	log.Fatal(http.ListenAndServe(":8081", web.Router()))
}
